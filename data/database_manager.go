package data

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseManager encapsulates database operations
type DatabaseManager struct {
	db *sql.DB
}

func (dm *DatabaseManager) DoExec(query string, args ...interface{}) (sql.Result, error) {
	return dm.db.Exec(query, args...)
}

func (dm *DatabaseManager) Ping() error {
	return dm.db.Ping()
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	// Add connection string parameters to handle concurrent access better
	connStr := dbPath + "?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_timeout=5000"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for better concurrent handling
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return &DatabaseManager{db: db}, nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	return dm.db.Close()
}

// GetDB returns the database connection (for migration purposes)
func (dm *DatabaseManager) GetDB() *sql.DB {
	return dm.db
}

// User operations
func (dm *DatabaseManager) CreateUser(user *User) error {
	user.Uuid = createUUID()
	user.CreatedAt = time.Now()

	stmt, err := dm.db.Prepare(
		"INSERT INTO users(uuid, name, email, password, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.Uuid, user.Name, user.Email, Encrypt(user.Password), user.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.Id = int(id)
	return nil
}

func (dm *DatabaseManager) GetUserByEmail(email string) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

func (dm *DatabaseManager) GetUserByID(id int) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE id=?", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

// Update user preferred categories
func (dm *DatabaseManager) UpdateUserPreferences(userID int, category1, category2 string) error {
	_, err := dm.db.Exec("UPDATE users SET prefered_category1=?, prefered_category2=? WHERE id=?",
		category1, category2, userID)
	return err
}

// Session operations
func (dm *DatabaseManager) CreateSession(user *User) (Session, error) {
	// Delete existing sessions for this user
	dm.db.Exec("DELETE from sessions where user_id=?", user.Id)

	// Create the session UUID
	sessionUUID := createUUID()

	// Create cookie string value in format "userId&sessionUUID"
	cookieString := fmt.Sprintf("%d&%s", user.Id, sessionUUID)

	// Calculate current time as hour*100 + minute
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	// Insert session
	stmt, err := dm.db.Prepare(
		"INSERT INTO sessions(uuid, email, user_id, created_at, cookie_string, active_last) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return Session{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sessionUUID, user.Email, user.Id, time.Now(), cookieString, currentTime)
	if err != nil {
		return Session{}, err
	}

	return Session{
		Uuid:         sessionUUID,
		Email:        user.Email,
		UserId:       user.Id,
		CreatedAt:    time.Now(),
		CookieString: cookieString,
		ActiveLast:   currentTime,
	}, nil
}

func (dm *DatabaseManager) ValidateSession(sessionUUID string) (Session, bool, error) {
	var session Session

	// Calculate current time as hour*100 + minute
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	err := dm.db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE uuid=?", sessionUUID).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)

	if err != nil {
		return session, false, err
	}

	if session.Id != 0 {
		// Only update active_last if it's been more than 5 minutes since last update
		// This reduces database writes and prevents locking issues
		timeDiff := currentTime - session.ActiveLast
		if timeDiff < 0 {
			// Handle day rollover (current time is next day)
			timeDiff = (2400 + currentTime) - session.ActiveLast
		}

		// Update if more than 5 minutes have passed (500 = 5 hours in our time format, 5 = 5 minutes)
		if timeDiff >= 5 {
			_, err = dm.db.Exec("UPDATE sessions SET active_last = ? WHERE uuid = ?", currentTime, sessionUUID)
			if err != nil {
				// Don't fail validation if we can't update timestamp, just log it
				log.Printf("Warning: Could not update active_last for session %s: %v\n", sessionUUID, err)
			} else {
				session.ActiveLast = currentTime
			}
		}
		return session, true, nil
	}

	return session, false, nil
}

func (dm *DatabaseManager) GetSessionByCookie(cookieValue string) (Session, error) {
	var session Session
	err := dm.db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE cookie_string=?", cookieValue).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	return session, err
}

func (dm *DatabaseManager) DeleteSession(sessionUUID string) error {
	_, err := dm.db.Exec("DELETE FROM sessions WHERE uuid=?", sessionUUID)
	return err
}

// Thread operations
func (dm *DatabaseManager) GetAllThreads() ([]Thread, error) {
	var threads []Thread

	rows, err := dm.db.Query("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads ORDER BY created_at DESC")
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
		if err != nil {
			continue
		}

		// Load user information for this thread
		user, err := dm.GetUserByID(thread.UserId)
		if err != nil {
			thread.User = user.Name
		}

		thread.User = user.Name
		thread.CreatedAtDate = thread.CreatedAt.Format("Jan 2, 2006 at 15:04")
		thread.NumReplies, _ = dm.GetThreadPostsCount(thread.Id)
		thread.LikesCount, _ = dm.GetThreadLikesCount(thread.Id)
		thread.DislikesCount, _ = dm.GetThreadDislikesCount(thread.Id)
		thread.Len = len(thread.Topic)
		threads = append(threads, thread)
	}

	return threads, nil
}

// Get all threads ordered by likes count (most liked first)
func (dm *DatabaseManager) GetAllThreadsByLikes() ([]Thread, error) {
	var threads []Thread

	query := `SELECT t.id, t.uuid, t.topic, t.user_id, t.created_at, t.category1, t.category2,
	          COALESCE(like_counts.like_count, 0) as like_count
	          FROM threads t
	          LEFT JOIN (
	              SELECT thread_id, COUNT(*) as like_count
	              FROM threadlikes
	              GROUP BY thread_id
	          ) like_counts ON t.id = like_counts.thread_id
	          ORDER BY like_count DESC, t.created_at DESC`

	rows, err := dm.db.Query(query)
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var likeCount int
		err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.UserId, &thread.CreatedAt,
			&thread.Category1, &thread.Category2, &likeCount)
		if err != nil {
			continue
		}

		// Load user information for this thread
		user, err := dm.GetUserByID(thread.UserId)
		if err != nil {
			thread.User = "Unknown User"
		} else {
			thread.User = user.Name
		}

		thread.CreatedAtDate = thread.CreatedAt.Format("Jan 2, 2006 at 15:04")
		thread.NumReplies, _ = dm.GetThreadPostsCount(thread.Id)
		thread.LikesCount, _ = dm.GetThreadLikesCount(thread.Id)
		thread.DislikesCount, _ = dm.GetThreadDislikesCount(thread.Id)

		threads = append(threads, thread)
	}

	return threads, nil
}

func (dm *DatabaseManager) CheckOnlineUsers(considerOnline int) ([]User, error) {
	var users []User
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	query := `
		SELECT DISTINCT u.id, u.uuid, u.name, u.email, u.created_at, s.active_last
		FROM users u 
		INNER JOIN sessions s ON u.id = s.user_id 
		WHERE s.active_last > 0`

	rows, err := dm.db.Query(query)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var activeLast int

		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &activeLast)
		if err != nil {
			continue
		}

		// Calculate time difference in minutes
		var timeDiff int
		if currentTime >= activeLast {
			hourDiff := (currentTime / 100) - (activeLast / 100)
			minuteDiff := (currentTime % 100) - (activeLast % 100)
			timeDiff = hourDiff*60 + minuteDiff
		} else {
			hourDiff := (24 + currentTime/100) - (activeLast / 100)
			minuteDiff := (currentTime % 100) - (activeLast % 100)
			timeDiff = hourDiff*60 + minuteDiff
		}

		if timeDiff <= considerOnline {
			users = append(users, user)
		}
	}

	return users, nil
}

// Thread operations
func (dm *DatabaseManager) CreateThread(topic, body string, userID, categoryID, subcategoryID int) (int64, error) {
	stmt, err := dm.db.Prepare("INSERT INTO threads(uuid, topic, body, user_id, created_at, category1, category2) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	uuid := createUUID()
	result, err := stmt.Exec(uuid, topic, body, userID, time.Now(), categoryID, subcategoryID)
	if err != nil {
		return 0, err
	}

	threadID, err := result.LastInsertId()
	return threadID, err
}

func (dm *DatabaseManager) GetThreadByID(id int) (Thread, error) {
	var thread Thread
	err := dm.db.QueryRow("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE id = ?", id).Scan(
		&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
	return thread, err
}

func (dm *DatabaseManager) GetThreadWithPosts(id int) (Thread, error) {
	// First get the thread
	thread, err := dm.GetThreadByID(id)
	if err != nil {
		return thread, err
	}

	// Then get the posts for this thread
	posts, err := dm.GetThreadPosts(id)
	if err != nil {
		return thread, err
	}

	// Attach posts to thread
	user, err := dm.GetUserByID(thread.UserId)
	if err == nil {
		thread.User = user.Name
		thread.Email = user.Email
		thread.CreatedAtDate = thread.CreatedAt.Format("Jan 2, 2006 at 15:04")
	} else {
		thread.User = "Unknown User"
		thread.Email = ""
		thread.CreatedAtDate = thread.CreatedAt.Format("Jan 2, 2006 at 15:04")
	}

	thread.Cards = posts
	return thread, nil
}

func (dm *DatabaseManager) GetThreadPosts(threadID int) ([]Post, error) {
	var posts []Post
	rows, err := dm.db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE thread_id=?", threadID)
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt)
		if err != nil {
			continue
		}

		// Load user information for this post
		user, err := dm.GetPostUser(post.UserId)
		if err != nil {
			// If user loading fails, create a placeholder
			post.User = "Unknown User"
		} else {
			post.User = user.Name
		}

		posts = append(posts, post)
	}
	return posts, nil
}

func (dm *DatabaseManager) GetThreadLikes(threadID int) ([]ThreadLikes, error) {
	var likes []ThreadLikes
	rows, err := dm.db.Query("SELECT * FROM threadlikes WHERE thread_id=?", threadID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like ThreadLikes
		err = rows.Scan(&like.Type, &like.UserId, &like.ThreadId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetThreadDislikes(threadID int) ([]ThreadDislikes, error) {
	var dislikes []ThreadDislikes
	rows, err := dm.db.Query("SELECT * FROM threaddislikes WHERE thread_id=?", threadID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike ThreadDislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.ThreadId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}

func (dm *DatabaseManager) ApplyThreadLike(userID, threadID int) error {
	stmt, err := dm.db.Prepare("INSERT INTO threadlikes(type, user_id, thread_id) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec("like", userID, threadID)
	return err
}

func (dm *DatabaseManager) ApplyThreadDislike(userID, threadID int) error {
	stmt, err := dm.db.Prepare("INSERT INTO threaddislikes(type, user_id, thread_id) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec("dislike", userID, threadID)
	return err
}

func (dm *DatabaseManager) HasThreadLiked(userID, threadID int) bool {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threadlikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	return err == nil && count > 0
}

func (dm *DatabaseManager) HasThreadDisliked(userID, threadID int) bool {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threaddislikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	return err == nil && count > 0
}

// Alternative method names for compatibility with thread.go
func (dm *DatabaseManager) HasUserLikedThread(userID, threadID int) int {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threadlikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	if err != nil {
		return count
	}
	return 0
}

func (dm *DatabaseManager) HasUserDislikedThread(userID, threadID int) int {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threaddislikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	if err != nil {
		return count
	}
	return 0
}

func (dm *DatabaseManager) DeleteThreadLikes(userID, threadID int) error {
	stmt, err := dm.db.Prepare("DELETE FROM threadlikes WHERE user_id=? AND thread_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, threadID)
	return err
}

func (dm *DatabaseManager) DeleteThreadDislikes(userID, threadID int) error {
	stmt, err := dm.db.Prepare("DELETE FROM threaddislikes WHERE user_id=? AND thread_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, threadID)
	return err
}

func (dm *DatabaseManager) CreatePost(threadID int, body string, userID int) (int64, error) {
	stmt, err := dm.db.Prepare("INSERT INTO posts(uuid, body, user_id, thread_id, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	uuid := createUUID()
	result, err := stmt.Exec(uuid, body, userID, threadID, time.Now())
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	return postID, err
}

func (dm *DatabaseManager) GetThreadPostsCount(threadID int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT count(*) FROM posts where thread_id=?", threadID).Scan(&count)
	return count, err
}

func (dm *DatabaseManager) GetThreadLikesCount(threadID int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threadlikes WHERE thread_id=?", threadID).Scan(&count)
	return count, err
}

func (dm *DatabaseManager) GetThreadDislikesCount(threadID int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threaddislikes WHERE thread_id=?", threadID).Scan(&count)
	return count, err
}

// Post like/dislike operations
func (dm *DatabaseManager) GetLikes(postID int) ([]Likes, error) {
	var likes []Likes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM likedposts WHERE post_id=?", postID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	var likeLength int
	for rows.Next() {
		var like Likes
		likeLength++
		if err = rows.Scan(&like.Type, &like.UserId, &like.PostId); err != nil {
			continue
		}
		likes = append(likes, like)
	}

	for i := range likes {
		likes[i].LengthOfLikes = likeLength - 1
	}
	return likes, nil
}

func (dm *DatabaseManager) GetDislikes(postID int) ([]Dislikes, error) {
	var dislikes []Dislikes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM dislikes WHERE post_id=?", postID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	var dislikeLength int
	for rows.Next() {
		var dislike Dislikes
		dislikeLength++
		if err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.PostId); err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}

	for i := range dislikes {
		dislikes[i].LengthOfDislikes = dislikeLength - 1
	}
	return dislikes, nil
}

// Additional thread like/dislike helper methods
func (dm *DatabaseManager) PrepareThreadLikedPosts(userID, threadID int) bool {
	return dm.HasThreadLiked(userID, threadID)
}

func (dm *DatabaseManager) PrepareThreadDislikedPosts(userID, threadID int) bool {
	return dm.HasThreadDisliked(userID, threadID)
}

// Post like/dislike helper methods
func (dm *DatabaseManager) PrepareLikedPosts(userID, postID int) bool {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM likedposts WHERE user_id=? AND post_id=?", userID, postID).Scan(&count)
	return err == nil && count > 0
}

func (dm *DatabaseManager) PrepareDislikedPosts(userID, postID int) bool {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM dislikes WHERE user_id=? AND post_id=?", userID, postID).Scan(&count)
	return err == nil && count > 0
}

func (dm *DatabaseManager) DeleteLikes(userID, postID int) error {
	stmt, err := dm.db.Prepare("DELETE FROM likedposts WHERE user_id=? AND post_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, postID)
	return err
}

func (dm *DatabaseManager) DeleteDislikes(userID, postID int) error {
	stmt, err := dm.db.Prepare("DELETE FROM dislikes WHERE user_id=? AND post_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, postID)
	return err
}

func (dm *DatabaseManager) ApplyLikes(userID, postID int) error {
	stmt, err := dm.db.Prepare("INSERT INTO likedposts(type, user_id, post_id) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec("like", userID, postID)
	return err
}

func (dm *DatabaseManager) ApplyDislikes(userID, postID int) error {
	stmt, err := dm.db.Prepare("INSERT INTO dislikes(type, user_id, post_id) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec("dislike", userID, postID)
	return err
}

// Thread retrieval methods
func (dm *DatabaseManager) GetThreads() ([]Thread, error) {
	var threads []Thread
	rows, err := dm.db.Query("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads ORDER BY created_at DESC")
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
		if err != nil {
			continue
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

// Smart thread voting methods
func (dm *DatabaseManager) SmartApplyThreadLike(userID, threadID int) error {
	// Check if user already liked this thread
	if dm.HasThreadLiked(userID, threadID) {
		// User already liked, so remove the like (toggle off)
		return dm.DeleteThreadLikes(userID, threadID)
	}

	// Check if user disliked this thread, if so remove the dislike first
	if dm.HasThreadDisliked(userID, threadID) {
		err := dm.DeleteThreadDislikes(userID, threadID)
		if err != nil {
			return err
		}
	}

	// Add the like
	return dm.ApplyThreadLike(userID, threadID)
}

func (dm *DatabaseManager) SmartApplyThreadDislike(userID, threadID int) error {
	// Check if user already disliked this thread
	if dm.HasThreadDisliked(userID, threadID) {
		// User already disliked, so remove the dislike (toggle off)
		return dm.DeleteThreadDislikes(userID, threadID)
	}

	// Check if user liked this thread, if so remove the like first
	if dm.HasThreadLiked(userID, threadID) {
		err := dm.DeleteThreadLikes(userID, threadID)
		if err != nil {
			return err
		}
	}

	// Add the dislike
	return dm.ApplyThreadDislike(userID, threadID)
}

// Post operations
func (dm *DatabaseManager) GetPostUser(postUserId int) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", postUserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) GetPostLikesCount(postId int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM likedposts WHERE post_id = ?", postId).Scan(&count)
	return count, err
}

func (dm *DatabaseManager) GetPostDislikesCount(postId int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM dislikes WHERE post_id = ?", postId).Scan(&count)
	return count, err
}

// Smart post voting methods
func (dm *DatabaseManager) SmartApplyPostLike(userID, postID int) error {
	// Check if user already liked this post
	hasLiked, err := dm.HasUserLikedPost(userID, postID)
	if err != nil {
		return err
	}
	if hasLiked {
		// User already liked, so remove the like (toggle off)
		return dm.RemovePostLike(userID, postID)
	}

	// Check if user disliked this post, if so remove the dislike first
	hasDisliked, err := dm.HasUserDislikedPost(userID, postID)
	if err != nil {
		return err
	}
	if hasDisliked {
		err := dm.RemovePostDislike(userID, postID)
		if err != nil {
			return err
		}
	}

	// Add the like
	return dm.AddPostLike(userID, postID)
}

func (dm *DatabaseManager) SmartApplyPostDislike(userID, postID int) error {
	// Check if user already disliked this post
	hasDisliked, err := dm.HasUserDislikedPost(userID, postID)
	if err != nil {
		return err
	}
	if hasDisliked {
		// User already disliked, so remove the dislike (toggle off)
		return dm.RemovePostDislike(userID, postID)
	}

	// Check if user liked this post, if so remove the like first
	hasLiked, err := dm.HasUserLikedPost(userID, postID)
	if err != nil {
		return err
	}
	if hasLiked {
		err := dm.RemovePostLike(userID, postID)
		if err != nil {
			return err
		}
	}

	// Add the dislike
	return dm.AddPostDislike(userID, postID)
}

// Get a post by ID
func (dm *DatabaseManager) GetPostByID(id int) (Post, error) {
	var post Post
	err := dm.db.QueryRow("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE id=?", id).
		Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt)
	return post, err
}

// Check if user has already liked a post
func (dm *DatabaseManager) HasUserLikedPost(userID int, postID int) (bool, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM likedposts WHERE user_id=? AND post_id=?", userID, postID).Scan(&count)
	return count > 0, err
}

// Check if user has already disliked a post
func (dm *DatabaseManager) HasUserDislikedPost(userID int, postID int) (bool, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM dislikes WHERE user_id=? AND post_id=?", userID, postID).Scan(&count)
	return count > 0, err
}

// Remove user's like from a post
func (dm *DatabaseManager) RemovePostLike(userID int, postID int) error {
	_, err := dm.db.Exec("DELETE FROM likedposts WHERE user_id=? AND post_id=?", userID, postID)
	return err
}

// Remove user's dislike from a post
func (dm *DatabaseManager) RemovePostDislike(userID int, postID int) error {
	_, err := dm.db.Exec("DELETE FROM dislikes WHERE user_id=? AND post_id=?", userID, postID)
	return err
}

// Add user's like to a post
func (dm *DatabaseManager) AddPostLike(userID int, postID int) error {
	_, err := dm.db.Exec("INSERT INTO likedposts (user_id, post_id) VALUES (?, ?)", userID, postID)
	return err
}

// Add user's dislike to a post
func (dm *DatabaseManager) AddPostDislike(userID int, postID int) error {
	_, err := dm.db.Exec("INSERT INTO dislikes (user_id, post_id) VALUES (?, ?)", userID, postID)
	return err
}

// Session management methods needed by session.go
func (dm *DatabaseManager) UpdateSessionCookieString(uuid, cookieValue string) error {
	_, err := dm.db.Exec("UPDATE sessions SET cookie_string=? WHERE uuid=?", cookieValue, uuid)
	return err
}

func (dm *DatabaseManager) DeleteSessionByUUID(uuid string) error {
	_, err := dm.db.Exec("DELETE FROM sessions WHERE uuid=?", uuid)
	return err
}

func (dm *DatabaseManager) DeleteAllSessions() error {
	_, err := dm.db.Exec("DELETE FROM sessions")
	return err
}

func (dm *DatabaseManager) GetSessionUser(userID int) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", userID).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return user, err
}

// Thread management methods needed by thread.go
func (dm *DatabaseManager) RemoveThreadLike(userID, threadID int) error {
	_, err := dm.db.Exec("DELETE FROM threadlikes WHERE user_id=? AND thread_id=?", userID, threadID)
	return err
}

func (dm *DatabaseManager) RemoveThreadDislike(userID, threadID int) error {
	_, err := dm.db.Exec("DELETE FROM threaddislikes WHERE user_id=? AND thread_id=?", userID, threadID)
	return err
}

func (dm *DatabaseManager) AddThreadLike(userID, threadID int) error {
	_, err := dm.db.Exec("INSERT INTO threadlikes(type, user_id, thread_id) VALUES(?, ?, ?)", "like", userID, threadID)
	return err
}

func (dm *DatabaseManager) AddThreadDislike(userID, threadID int) error {
	_, err := dm.db.Exec("INSERT INTO threaddislikes(type, user_id, thread_id) VALUES(?, ?, ?)", "dislike", userID, threadID)
	return err
}

func (dm *DatabaseManager) GetPostLikes(postID int) ([]Likes, error) {
	var likes []Likes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM likedposts WHERE post_id=?", postID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like Likes
		err = rows.Scan(&like.Type, &like.UserId, &like.PostId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetPostDislikes(postID int) ([]Dislikes, error) {
	var dislikes []Dislikes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM dislikes WHERE post_id=?", postID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike Dislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.PostId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}

func (dm *DatabaseManager) GetThreadPostCount(threadID int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT count(*) FROM posts where thread_id=?", threadID).Scan(&count)
	return count, err
}

// User management methods needed by user.go
func (dm *DatabaseManager) CreateThreadByUser(topic, body string, userID int, category1, category2 string) (int64, error) {
	result, err := dm.db.Exec("INSERT INTO threads(uuid, topic, body, user_id, created_at, category1, category2) VALUES(?, ?, ?, ?, ?, ?, ?)",
		createUUID(), topic, body, userID, time.Now(), category1, category2)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (dm *DatabaseManager) CreatePostByUser(body string, userID, threadID int) (int64, error) {
	result, err := dm.db.Exec("INSERT INTO posts(uuid, body, user_id, thread_id, created_at) VALUES(?, ?, ?, ?, ?)",
		createUUID(), body, userID, threadID, time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Statistics methods needed by statistics.go
func (dm *DatabaseManager) GetUserCount() (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (dm *DatabaseManager) GetMostActiveUsers(limit int) ([]User, error) {
	var users []User
	rows, err := dm.db.Query(`
		SELECT u.id, u.uuid, u.name, u.email, u.password, u.created_at, COUNT(t.id) AS thread_count
		FROM users u
		LEFT JOIN threads t ON u.id = t.user_id
		GROUP BY u.id
		ORDER BY thread_count DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var threadCount int
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &threadCount)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (dm *DatabaseManager) GetTotalPostsCount() (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	return count, err
}

func (dm *DatabaseManager) GetTotalThreadsCount() (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threads").Scan(&count)
	return count, err
}

// Additional methods needed by user.go
func (dm *DatabaseManager) CreateThreadLikeOnCreation(userID, threadID int) error {
	_, err := dm.db.Exec("INSERT INTO threadlikes(type, user_id, thread_id) VALUES(?, ?, ?)", "creator", userID, threadID)
	return err
}

func (dm *DatabaseManager) CreateThreadDislikeOnCreation(userID, threadID int) error {
	_, err := dm.db.Exec("INSERT INTO threaddislikes(type, user_id, thread_id) VALUES(?, ?, ?)", "creator", userID, threadID)
	return err
}

func (dm *DatabaseManager) CreatePostLikeOnCreation(userID, postID int) error {
	_, err := dm.db.Exec("INSERT INTO likedposts(type, user_id, post_id) VALUES(?, ?, ?)", "creator", userID, postID)
	return err
}

func (dm *DatabaseManager) CreatePostDislikeOnCreation(userID, postID int) error {
	_, err := dm.db.Exec("INSERT INTO dislikes(type, user_id, post_id) VALUES(?, ?, ?)", "creator", userID, postID)
	return err
}

func (dm *DatabaseManager) GetUserByEmailDetailed(email string) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) GetUserByUUID(uuid string) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE uuid=?", uuid).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) DeleteUserByID(userID int) error {
	_, err := dm.db.Exec("DELETE FROM users WHERE id=?", userID)
	return err
}

func (dm *DatabaseManager) DeleteAllUsers() error {
	_, err := dm.db.Exec("DELETE FROM users")
	return err
}

func (dm *DatabaseManager) GetAllUsers() ([]User, error) {
	var users []User
	rows, err := dm.db.Query("SELECT id, uuid, name, email, password, created_at FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (dm *DatabaseManager) GetUserLikedPosts(userID int) ([]Likes, error) {
	var likes []Likes
	rows, err := dm.db.Query("SELECT COALESCE(type, 'like') as type, user_id, post_id FROM likedposts WHERE user_id=?", userID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like Likes
		err = rows.Scan(&like.Type, &like.UserId, &like.PostId)
		if err != nil {
			fmt.Println("Error scanning like:", err)
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetUserLikedThreads(userID int) ([]ThreadLikes, error) {
	var likes []ThreadLikes
	rows, err := dm.db.Query("SELECT type, user_id, thread_id FROM threadlikes WHERE user_id=?", userID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like ThreadLikes
		err = rows.Scan(&like.Type, &like.UserId, &like.ThreadId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetUserDislikedThreads(userID int) ([]ThreadDislikes, error) {
	var dislikes []ThreadDislikes
	rows, err := dm.db.Query("SELECT type, user_id, thread_id FROM threaddislikes WHERE user_id=?", userID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike ThreadDislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.ThreadId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}

func (dm *DatabaseManager) GetUserDislikedPosts(userID int) ([]Dislikes, error) {
	var dislikes []Dislikes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM dislikes WHERE user_id=?", userID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike Dislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.PostId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}

func (dm *DatabaseManager) CheckUserExists(email, name string) (bool, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR name = ?", email, name).Scan(&count)
	return count > 0, err
}

// Additional methods needed by account routes
func (dm *DatabaseManager) GetUserCreatedPosts(userID int) ([]Post, error) {
	var posts []Post
	rows, err := dm.db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE user_id=? ORDER BY created_at DESC", userID)
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt)
		if err != nil {
			continue
		}

		// Get user info for this post
		user, err := dm.GetPostUser(post.UserId)
		if err == nil {
			post.User = user.Name
		} else {
			post.User = "Unknown User"
		}
		// post.CreatedAt = post.CreatedAt.Local()
		posts = append(posts, post)
	}
	return posts, nil
}

func (dm *DatabaseManager) GetUserCreatedThreads(userID int) ([]Thread, error) {
	var threads []Thread
	rows, err := dm.db.Query("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE user_id=? ORDER BY created_at DESC", userID)
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
		if err != nil {
			fmt.Println("Error scanning thread for Account:", err)
			continue
		}

		// Get user info for this thread
		user, _ := dm.GetUserByID(thread.UserId)

		thread.User = user.Name
		// Format the creation date for display
		thread.CreatedAtDate = thread.CreatedAt.Format("Jan 2, 2006 at 15:04")

		// Get number of replies (posts) for this thread
		thread.NumReplies, _ = dm.GetThreadPostsCount(thread.Id)

		// Get posts for this thread
		posts, err := dm.GetThreadPosts(thread.Id)
		// TODO: User.Name
		if err == nil {
			thread.Cards = posts
			thread.LengthOfPosts = len(posts)
			thread.Email = user.Email
		}

		threads = append(threads, thread)
	}
	return threads, nil
}

func (dm *DatabaseManager) GetThreadsByCategories(category1, category2 string) ([]Thread, error) {
	var threads []Thread
	var query string
	var args []interface{}

	if category1 != "" && category2 != "" {
		query = "SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE category1=? OR category2=? ORDER BY created_at DESC"
		args = []interface{}{category1, category2}
	} else if category1 != "" {
		query = "SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE category1=? ORDER BY created_at DESC"
		args = []interface{}{category1}
	} else if category2 != "" {
		query = "SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE category2=? ORDER BY created_at DESC"
		args = []interface{}{category2}
	} else {
		// Return all threads if no categories specified
		return dm.GetAllThreads()
	}

	rows, err := dm.db.Query(query, args...)
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
		if err != nil {
			continue
		}

		// Get user info for this thread
		user, err := dm.GetUserByID(thread.UserId)
		if err == nil {
			thread.User = user.Name
			thread.NumReplies, _ = dm.GetThreadPostsCount(thread.Id)
			thread.CreatedAtDate = thread.CreatedAt.Format("Jan 2, 2006 at 15:04")
			thread.LikesCount, _ = dm.GetThreadLikesCount(thread.Id)
			thread.DislikesCount, _ = dm.GetThreadDislikesCount(thread.Id)
			thread.Len = len(thread.Topic)
		}

		threads = append(threads, thread)
	}
	return threads, nil
}

// GetThreadsByUserID returns all threads created by a specific user
func (dm *DatabaseManager) GetThreadsByUserID(userID int) ([]Thread, error) {
	rows, err := dm.db.Query(`
		SELECT t.id, t.uuid, t.topic, t.body, t.user_id, u.name, u.email, 
		       t.created_at, t.category1,
		       COALESCE(p.reply_count, 0) as num_replies
		FROM threads t 
		JOIN users u ON t.user_id = u.id 
		LEFT JOIN (
			SELECT thread_id, COUNT(*) as reply_count 
			FROM posts 
			GROUP BY thread_id
		) p ON t.id = p.thread_id
		WHERE t.user_id = ?
		ORDER BY t.created_at DESC`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []Thread
	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body,
			&thread.UserId, &thread.User, &thread.Email,
			&thread.CreatedAt, &thread.Category1, &thread.NumReplies)
		if err != nil {
			return nil, err
		}

		// Format created date
		thread.CreatedAtDate = thread.CreatedAt.Format("2006-01-02 15:04:05")

		threads = append(threads, thread)
	}

	return threads, rows.Err()
}

// GetPostsByUserID returns all posts created by a specific user
func (dm *DatabaseManager) GetPostsByUserID(userID int) ([]Post, error) {
	rows, err := dm.db.Query(`
		SELECT p.id, p.uuid, p.body, p.user_id, p.thread_id, p.created_at, u.name
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId,
			&post.ThreadId, &post.CreatedAt, &post.User)
		if err != nil {
			return nil, err
		}

		// Format created date for template
		post.FormattedDate = post.CreatedAt.Format("2006-01-02 15:04:05")

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

// GetLikedPostsByUserID returns all posts liked by a specific user
func (dm *DatabaseManager) GetLikedPostsByUserID(userID int) ([]Post, error) {
	rows, err := dm.db.Query(`
		SELECT p.id, p.uuid, p.body, p.user_id, p.thread_id, p.created_at, u.name
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		JOIN likedposts l ON p.id = l.post_id
		WHERE l.user_id = ?
		ORDER BY p.created_at DESC`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId,
			&post.ThreadId, &post.CreatedAt, &post.User)
		if err != nil {
			return nil, err
		}

		// Format created date for template
		post.FormattedDate = post.CreatedAt.Format("2006-01-02 15:04:05")

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

// GetLikedThreadsByUserID returns all threads liked by a specific user
func (dm *DatabaseManager) GetLikedThreadsByUserID(userID int) ([]Thread, error) {
	rows, err := dm.db.Query(`
		SELECT t.id, t.uuid, t.topic, t.body, t.user_id, u.name, u.email, 
		       t.created_at, t.category1,
		       COALESCE(p.reply_count, 0) as num_replies
		FROM threads t 
		JOIN users u ON t.user_id = u.id 
		LEFT JOIN (
			SELECT thread_id, COUNT(*) as reply_count 
			FROM posts 
			GROUP BY thread_id
		) p ON t.id = p.thread_id
		JOIN threadlikes tl ON t.id = tl.thread_id
		WHERE tl.user_id = ?
		ORDER BY t.created_at DESC`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []Thread
	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body,
			&thread.UserId, &thread.User, &thread.Email,
			&thread.CreatedAt, &thread.Category1, &thread.NumReplies)
		if err != nil {
			return nil, err
		}

		// Format created date
		thread.CreatedAtDate = thread.CreatedAt.Format("2006-01-02 15:04:05")

		threads = append(threads, thread)
	}

	return threads, rows.Err()
}
