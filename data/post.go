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

// get the user who wrote the post using DatabaseManager (preferred method)
func (post *Post) UserWithDB(dm *DatabaseManager) (User, error) {
	return dm.GetPostUser(post.UserId)
}

// get the number of likes for this post using DatabaseManager
func (post *Post) GetLikesCountWithDB(dm *DatabaseManager) (int, error) {
	return dm.GetPostLikesCount(post.Id)
}

// get the number of dislikes for this post using DatabaseManager
func (post *Post) GetDislikesCountWithDB(dm *DatabaseManager) (int, error) {
	return dm.GetPostDislikesCount(post.Id)
}

// Smart like function using DatabaseManager - handles vote switching
func (dm *DatabaseManager) SmartApplyPostLike(userID int, postID int) error {
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

// Smart dislike function using DatabaseManager - handles vote switching
func (dm *DatabaseManager) SmartApplyPostDislike(userID int, postID int) error {
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
