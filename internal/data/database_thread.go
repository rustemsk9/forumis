package data

import (
	"fmt"
	"forum/models"
	"forum/utils"
	"time"
)

func (dm *DatabaseManager) GetTotalThreadsCount() (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threads").Scan(&count)
	return count, err
}

// Thread operations
func (dm *DatabaseManager) GetAllThreads() ([]models.Thread, error) {
	var threads []models.Thread

	rows, err := dm.db.Query("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads ORDER BY created_at DESC")
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread models.Thread
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
func (dm *DatabaseManager) GetAllThreadsByLikes() ([]models.Thread, error) {
	var threads []models.Thread

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
		var thread models.Thread
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

// Thread operations
func (dm *DatabaseManager) CreateThread(topic, body string, userID, categoryID, subcategoryID int) (int64, error) {
	stmt, err := dm.db.Prepare("INSERT INTO threads(uuid, topic, body, user_id, created_at, category1, category2) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	uuid := utils.CreateUUID()
	result, err := stmt.Exec(uuid, topic, body, userID, time.Now(), categoryID, subcategoryID)
	if err != nil {
		return 0, err
	}

	threadID, err := result.LastInsertId()
	return threadID, err
}

func (dm *DatabaseManager) GetThreadByID(id int) (models.Thread, error) {
	var thread models.Thread
	err := dm.db.QueryRow("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE id = ?", id).Scan(
		&thread.Id, &thread.Uuid, &thread.Topic, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Category1, &thread.Category2)
	return thread, err
}

func (dm *DatabaseManager) GetThreadWithPosts(id int) (models.Thread, error) {
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

func (dm *DatabaseManager) GetThreadLikes(threadID int) ([]models.ThreadLikes, error) {
	var likes []models.ThreadLikes
	rows, err := dm.db.Query("SELECT * FROM threadlikes WHERE thread_id=?", threadID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like models.ThreadLikes
		err = rows.Scan(&like.Type, &like.UserId, &like.ThreadId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetThreadDislikes(threadID int) ([]models.ThreadDislikes, error) {
	var dislikes []models.ThreadDislikes
	rows, err := dm.db.Query("SELECT * FROM threaddislikes WHERE thread_id=?", threadID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike models.ThreadDislikes
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

// Additional thread like/dislike helper methods
func (dm *DatabaseManager) PrepareThreadLikedPosts(userID, threadID int) bool {
	return dm.HasThreadLiked(userID, threadID)
}

func (dm *DatabaseManager) PrepareThreadDislikedPosts(userID, threadID int) bool {
	return dm.HasThreadDisliked(userID, threadID)
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
func (dm *DatabaseManager) GetThreads() ([]models.Thread, error) {
	var threads []models.Thread
	rows, err := dm.db.Query("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads ORDER BY created_at DESC")
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread models.Thread
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

func (dm *DatabaseManager) GetThreadPostCount(threadID int) (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT count(*) FROM posts where thread_id=?", threadID).Scan(&count)
	return count, err
}

// User management methods needed by user.go
func (dm *DatabaseManager) CreateThreadByUser(topic, body string, userID int, category1, category2 string) (int64, error) {
	result, err := dm.db.Exec("INSERT INTO threads(uuid, topic, body, user_id, created_at, category1, category2) VALUES(?, ?, ?, ?, ?, ?, ?)",
		utils.CreateUUID(), topic, body, userID, time.Now(), category1, category2)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (dm *DatabaseManager) GetUserCreatedThreads(userID int) ([]models.Thread, error) {
	var threads []models.Thread
	rows, err := dm.db.Query("SELECT id, uuid, topic, body, user_id, created_at, category1, category2 FROM threads WHERE user_id=? ORDER BY created_at DESC", userID)
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread models.Thread
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

func (dm *DatabaseManager) GetThreadsByCategories(category1, category2 string) ([]models.Thread, error) {
	var threads []models.Thread
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
		var thread models.Thread
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
func (dm *DatabaseManager) GetThreadsByUserID(userID int) ([]models.Thread, error) {
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

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
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
