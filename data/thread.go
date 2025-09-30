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

func (thread *Thread) GetAllThreads() (threads []Thread, err error) {
	threads, err = threadDM.GetAllThreads()
	if err != nil {
		fmt.Println("Error on select GetAllThreads")
		return
	}
	return
}

func (thread *Thread) FilterThreadsByCategories(category1, category2 string) ([]Thread, error) {
	// We need to add this method to DatabaseManager
	return threadDM.GetThreadsByCategories(category1, category2)
}

// Additional functions needed by API routes
func (thread *Thread) ThreadById(threadID int) (Thread, error) {
	return threadDM.GetThreadByID(threadID)
}

func (thread *Thread) GetThreadLikesCount(threadID int) (int, error) {
	return threadDM.GetThreadLikesCount(threadID)
}

func (thread *Thread) GetThreadDislikesCount(threadID int) (int, error) {
	return threadDM.GetThreadDislikesCount(threadID)
}

func ReadDMFromAccount(URLIDConv int) ([]Thread, error) {
	// Get user info first
	userInfo, err := threadDM.GetUserByID(URLIDConv)
	if err != nil {
		return []Thread{}, err
	}

	// Get user's threads
	userThreads, err := threadDM.GetThreadsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user threads: %v\n", err)
		userThreads = []Thread{} // Empty slice if error
	}

	// Get user's posts
	userPosts, err := threadDM.GetPostsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user posts: %v\n", err)
		userPosts = []Post{} // Empty slice if error
	}

	// Get user's liked posts
	likedPosts, err := threadDM.GetLikedPostsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user liked posts: %v\n", err)
		likedPosts = []Post{} // Empty slice if error
	}

	// Get user's liked threads
	likedThreads, err := threadDM.GetLikedThreadsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user liked threads: %v\n", err)
		likedThreads = []Thread{} // Empty slice if error
	}

	// Create account data structure that matches the template expectations
	// The template expects an array where the first element has user info and contains Cards/LikedPosts
	var templateData []Thread

	// Always create at least one element for the template
	var firstElement Thread

	if len(userThreads) > 0 {
		// Use the first thread as a base
		firstElement = userThreads[0]

		// Add the rest of the threads if any
		if len(userThreads) > 1 {
			templateData = append(templateData, userThreads[1:]...)
		}
	} else {
		// No threads, create a dummy thread with user info
		firstElement = Thread{
			User:      userInfo.Name,
			Email:     userInfo.Email,
			CreatedAt: userInfo.CreatedAt,
		}
	}

	// Set the Cards and LikedPosts on the first element
	firstElement.Cards = userPosts
	firstElement.LikedPosts = likedPosts
	firstElement.UserLikedThreads = likedThreads

	// Insert the first element at the beginning
	templateData = append([]Thread{firstElement}, templateData...)

	return templateData, nil
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
