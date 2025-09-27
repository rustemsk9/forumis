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

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
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
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) GetUserByID(id int) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE id=?", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
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
		// Update active_last with current time
		_, err = dm.db.Exec("UPDATE sessions SET active_last = ? WHERE uuid = ?", currentTime, sessionUUID)
		if err != nil {
			log.Printf("Warning: Could not update active_last for session %s: %v\n", sessionUUID, err)
		}
		session.ActiveLast = currentTime
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
	
	rows, err := dm.db.Query("SELECT id, uuid, topic, user_id, created_at, category1, category2 FROM threads ORDER BY created_at DESC")
	if err != nil {
		return threads, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Topic, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
		if err != nil {
			continue
		}
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
			hourDiff := (currentTime/100) - (activeLast/100)
			minuteDiff := (currentTime%100) - (activeLast%100)
			timeDiff = hourDiff*60 + minuteDiff
		} else {
			hourDiff := (24 + currentTime/100) - (activeLast/100)
			minuteDiff := (currentTime%100) - (activeLast%100)
			timeDiff = hourDiff*60 + minuteDiff
		}
		
		if timeDiff <= considerOnline {
			users = append(users, user)
		}
	}
	
	return users, nil
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