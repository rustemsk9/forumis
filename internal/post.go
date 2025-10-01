package internal

import (
	"fmt"
	"forum/internal/data"
	"forum/models"
)

var postDM *data.DatabaseManager

func InitPostDM(dm *data.DatabaseManager) {
	postDM = dm
}

func CreatePost(threadID int, body string, userID int) (int64, error) {
	return postDM.CreatePostByUser(body, userID, threadID)
}

func GetLikes(postID int) (int, error) {
	return postDM.GetPostLikesCount(postID)
}

func GetDislikes(postID int) (int, error) {
	return postDM.GetPostDislikesCount(postID)
}

func LikeOnPostCreation(userID, postID int) error {
	return postDM.CreatePostLikeOnCreation(userID, postID)
}

func DislikeOnPostCreation(userID, postID int) error {
	return postDM.CreatePostDislikeOnCreation(userID, postID)
}

func PrepareLikedPosts(userID int, postID int) bool {
	likes, err := postDM.GetUserLikedPosts(userID)
	if err != nil {
		fmt.Println("Error on PrepareLikedPosts")
		return false
	}

	for _, like := range likes {
		if like.PostId == postID {
			return true
		}
	}
	return false
}

// true if user disliked this post
func PrepareDislikedPosts(userID, postID int) bool {
	dislikes, err := postDM.GetUserDislikedPosts(userID)
	if err != nil {
		fmt.Println("Error on PrepareDislikedPosts")
		return false
	}

	for _, dislike := range dislikes {
		if dislike.PostId == postID {
			return true
		}
	}
	return false
}

func GetLikesPostsFromDB(likes []models.Likes) ([]models.Post, error) {
	// This function should get full post objects from the likes
	var posts []models.Post
	for _, like := range likes {
		post, err := postDM.GetPostByID(like.PostId)
		if err != nil {
			continue // Skip posts that can't be found
		}

		// Populate User information for the post
		user, err := userDM.GetPostUser(post.UserId)
		if err != nil {
			// If user loading fails, create a placeholder
			post.User = "Unknown User"
		} else {
			post.User = user.Name
		}

		// Note: CreatedAtDate formatting will be done by the template method

		posts = append(posts, post)
	}
	return posts, nil
}
