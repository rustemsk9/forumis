package data

import (
	"time"
)

type Post struct {
	Id        int
	Uuid      string
	Body      string
	UserId    int
	ThreadId  int
	CreatedAt time.Time
}

// format the CreatedAt date to display nicely on the screen
func (post *Post) CreatedAtDate() string {
	return post.CreatedAt.Format("Jan/2/2006 3:04pm")
}

// get the user who wrote the post
func (post *Post) User() (user User) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", post.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return
}

// get the number of likes for this post
func (post *Post) GetLikesCount() int {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM likedposts WHERE post_id=?", post.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// get the number of dislikes for this post
func (post *Post) GetDislikesCount() int {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM dislikes WHERE post_id=?", post.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// Get a post by ID
func PostById(id int) (post Post, err error) {
	post = Post{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE id=?", id).
		Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt)
	return
}

// Check if user has already liked a post
func HasUserLikedPost(userID int, postID int) bool {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM likedposts WHERE user_id=? AND post_id=?", userID, postID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Check if user has already disliked a post
func HasUserDislikedPost(userID int, postID int) bool {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM dislikes WHERE user_id=? AND post_id=?", userID, postID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Remove user's like from a post
func RemovePostLike(userID int, postID int) error {
	stmt, err := Db.Prepare("DELETE FROM likedposts WHERE user_id=? AND post_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID, postID)
	return err
}

// Remove user's dislike from a post
func RemovePostDislike(userID int, postID int) error {
	stmt, err := Db.Prepare("DELETE FROM dislikes WHERE user_id=? AND post_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID, postID)
	return err
}

// Smart like function for posts - handles vote switching
func SmartApplyPostLike(userID int, postID int) error {
	// Check if user already liked this post
	if HasUserLikedPost(userID, postID) {
		// User already liked, so remove the like (toggle off)
		return RemovePostLike(userID, postID)
	}

	// Check if user disliked this post, if so remove the dislike first
	if HasUserDislikedPost(userID, postID) {
		err := RemovePostDislike(userID, postID)
		if err != nil {
			return err
		}
	}

	// Add the like
	ApplyLikes("like", userID, postID)
	return nil
}

// Smart dislike function for posts - handles vote switching
func SmartApplyPostDislike(userID int, postID int) error {
	// Check if user already disliked this post
	if HasUserDislikedPost(userID, postID) {
		// User already disliked, so remove the dislike (toggle off)
		return RemovePostDislike(userID, postID)
	}

	// Check if user liked this post, if so remove the like first
	if HasUserLikedPost(userID, postID) {
		err := RemovePostLike(userID, postID)
		if err != nil {
			return err
		}
	}

	// Add the dislike
	ApplyDislikes("dislike", userID, postID)
	return nil
}
