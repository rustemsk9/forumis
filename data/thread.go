package data

import (
	"errors"
	"fmt"
	"net/http"
)

// Global DatabaseManager instance for thread operations
var threadDM *DatabaseManager

// InitThreadDM initializes the global DatabaseManager for thread operations
func InitThreadDM(dm *DatabaseManager) {
	threadDM = dm
}

func SortThreadsByLikesDesc(threads []Thread) ([]Thread, error) {
	// Simple bubble sort for demonstration; consider more efficient sorting for large datasets
	n := len(threads)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if threads[j].LikesCount < threads[j+1].LikesCount {
				threads[j], threads[j+1] = threads[j+1], threads[j]
			}
		}
	}
	return threads, nil
}

func SortThreadsByLatest(threads []Thread) ([]Thread, error) {
	// Simple bubble sort for demonstration; consider more efficient sorting for large datasets
	n := len(threads)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if threads[j].CreatedAt.Before(threads[j+1].CreatedAt) {
				threads[j], threads[j+1] = threads[j+1], threads[j]
			}
		}
	}
	return threads, nil
}
func GetCookieValue(request *http.Request) int {
	// Debug: Print all cookies
	fmt.Printf("DEBUG: All cookies for request: ")
	for _, cookie := range request.Cookies() {
		fmt.Printf("[%s=%s] ", cookie.Name, cookie.Value)
	}
	fmt.Println()

	// Get the _cookie from request
	cook, err := request.Cookie("_cookie")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			fmt.Println("Cookie '_cookie' not found")
		} else {
			fmt.Println("Error on Get Cookie:", err)
		}
		return -1
	}

	fmt.Printf("Found cookie: %s = %s\n", cook.Name, cook.Value)

	// Check if session exists in database with this cookie string
	session, err := GetSessionByCookie(cook.Value)
	if err != nil {
		fmt.Printf("No valid session found for cookie: %s\n", cook.Value)
		return -1
	}

	if session.UserId == 0 {
		fmt.Println("Session found but no valid user ID")
		return -1
	}

	fmt.Printf("Authenticated user ID %d via cookie string\n", session.UserId)
	return session.UserId
}

func GetThreadLikes(findid int) (THLI []ThreadLikes) {
	likes, err := threadDM.GetThreadLikes(findid)
	if err != nil {
		fmt.Println("Error on select GetThreadLikes")
		return
	}
	return likes
}

func GetThreadDislikes(findid int) (THDI []ThreadDislikes) {
	dislikes, err := threadDM.GetThreadDislikes(findid)
	if err != nil {
		fmt.Println("Error on select GetThreadDislikes")
		return
	}
	return dislikes
}

func ApplyThreadLike(stateLike string, userID int, threadId int) {
	fmt.Println("Apply thread likes in database proccess")
	threadDM.AddThreadLike(userID, threadId)
	// can apply only if threadId is right // TODO check sequence

}

func ApplyThreadDislike(stateLike string, userID int, threadId int) {
	fmt.Println("Apply thread dislikes in database proccess")
	threadDM.AddThreadDislike(userID, threadId)
	// can apply only if threadId is right // TODO check sequence
}

// Check if user has already liked a thread
func HasUserLikedThread(userID int, threadID int) bool {
	return threadDM.HasUserLikedThread(userID, threadID) > 0
}

// Check if user has already disliked a thread
func HasUserDislikedThread(userID int, threadID int) bool {
	return threadDM.HasUserDislikedThread(userID, threadID) > 0
}

// Remove user's like from a thread
func RemoveThreadLike(userID int, threadID int) error {
	return threadDM.RemoveThreadLike(userID, threadID)
}

// Remove user's dislike from a thread
func RemoveThreadDislike(userID int, threadID int) error {
	return threadDM.RemoveThreadDislike(userID, threadID)
}

// Smart like function - handles vote switching
func SmartApplyThreadLike(userID int, threadID int) error {
	// Check if user already liked this thread
	if HasUserLikedThread(userID, threadID) {
		// User already liked, so remove the like (toggle off)
		return RemoveThreadLike(userID, threadID)
	}

	// Check if user disliked this thread, if so remove the dislike first
	if HasUserDislikedThread(userID, threadID) {
		err := RemoveThreadDislike(userID, threadID)
		if err != nil {
			return err
		}
	}

	// Add the like
	ApplyThreadLike("like", userID, threadID)
	return nil
}

// Smart dislike function - handles vote switching
func SmartApplyThreadDislike(userID int, threadID int) error {
	// Check if user already disliked this thread
	if HasUserDislikedThread(userID, threadID) {
		// User already disliked, so remove the dislike (toggle off)
		return RemoveThreadDislike(userID, threadID)
	}

	// Check if user liked this thread, if so remove the like first
	if HasUserLikedThread(userID, threadID) {
		err := RemoveThreadLike(userID, threadID)
		if err != nil {
			return err
		}
	}

	// Add the dislike
	ApplyThreadDislike("dislike", userID, threadID)
	return nil
}

// }
func GetLikes(postID int) (Li []Likes) {
	Li, err := threadDM.GetLikes(postID)
	if err != nil {
		return
	}
	return Li
}

func GetDislikes(postID int) (Di []Dislikes) {
	Di, err := threadDM.GetDislikes(postID)
	if err != nil {
		return
	}
	return Di
}

// Thread methods needed by API routes
func (thread *Thread) GetLikesCount() int {
	likes, err := threadDM.GetThreadLikes(thread.Id)
	if err != nil {
		return 0
	}
	return len(likes)
}

func (thread *Thread) GetDislikesCount() int {
	dislikes, err := threadDM.GetThreadDislikes(thread.Id)
	if err != nil {
		return 0
	}
	return len(dislikes)
}

func ApplyLikes(stateLike string, userID int, postID int) {
	fmt.Println("---------------------")
	fmt.Println("in database proccess")
	threadDM.ApplyLikes(userID, postID)
}

func ApplyDislikes(stateLike string, userID int, postID int) {
	fmt.Println("---------------------")
	fmt.Println("in database proccess")
	threadDM.ApplyDislikes(userID, postID)

}
