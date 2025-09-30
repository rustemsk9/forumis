package data

import (
	"forum/models"
	"forum/utils"
	"time"
)

func (dm *DatabaseManager) GetTotalPostsCount() (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	return count, err
}

func (dm *DatabaseManager) CreatePostByUser(body string, userID, threadID int) (int64, error) {
	result, err := dm.db.Exec("INSERT INTO posts(uuid, body, user_id, thread_id, created_at) VALUES(?, ?, ?, ?, ?)",
		utils.CreateUUID(), body, userID, threadID, time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Additional methods needed by account routes
func (dm *DatabaseManager) GetUserCreatedPosts(userID int) ([]models.Post, error) {
	var posts []models.Post
	rows, err := dm.db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE user_id=? ORDER BY created_at DESC", userID)
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
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

// GetPostsByUserID returns all posts created by a specific user
func (dm *DatabaseManager) GetPostsByUserID(userID int) ([]models.Post, error) {
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

	var posts []models.Post
	for rows.Next() {
		var post models.Post
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
func (dm *DatabaseManager) GetLikedPostsByUserID(userID int) ([]models.Post, error) {
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

	var posts []models.Post
	for rows.Next() {
		var post models.Post
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

func (dm *DatabaseManager) GetThreadPosts(threadID int) ([]models.Post, error) {
	var posts []models.Post
	rows, err := dm.db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE thread_id=?", threadID)
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
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

func (dm *DatabaseManager) CreatePost(threadID int, body string, userID int) (int64, error) {
	stmt, err := dm.db.Prepare("INSERT INTO posts(uuid, body, user_id, thread_id, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	uuid := utils.CreateUUID()
	result, err := stmt.Exec(uuid, body, userID, threadID, time.Now())
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	return postID, err
}

// Post like/dislike operations
func (dm *DatabaseManager) GetLikes(postID int) ([]models.Likes, error) {
	var likes []models.Likes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM likedposts WHERE post_id=?", postID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	var likeLength int
	for rows.Next() {
		var like models.Likes
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

func (dm *DatabaseManager) GetDislikes(postID int) ([]models.Dislikes, error) {
	var dislikes []models.Dislikes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM dislikes WHERE post_id=?", postID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	var dislikeLength int
	for rows.Next() {
		var dislike models.Dislikes
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

// GetLikedThreadsByUserID returns all threads liked by a specific user
func (dm *DatabaseManager) GetLikedThreadsByUserID(userID int) ([]models.Thread, error) {
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

// Post operations
func (dm *DatabaseManager) GetPostUser(postUserId int) (models.User, error) {
	var user models.User
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
func (dm *DatabaseManager) GetPostByID(id int) (models.Post, error) {
	var post models.Post
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

func (dm *DatabaseManager) GetPostLikes(postID int) ([]models.Likes, error) {
	var likes []models.Likes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM likedposts WHERE post_id=?", postID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like models.Likes
		err = rows.Scan(&like.Type, &like.UserId, &like.PostId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetPostDislikes(postID int) ([]models.Dislikes, error) {
	var dislikes []models.Dislikes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM dislikes WHERE post_id=?", postID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike models.Dislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.PostId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}
