package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"forum/data"
	"forum/utils"
)

func ReadThreadsFromAccount(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		utils.MethodNotAllowed(writer, request, "POST method not allowed")
		return
	}

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		http.Redirect(writer, request, "/login", 302)
		return
	}

	// Get user ID from URL parameters
	URLID := request.URL.Query().Get("user_id")
	if URLID == "" {
		utils.BadRequest(writer, request, "User ID is required")
		return
	}

	URLIDConv, err := strconv.Atoi(URLID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid user ID format")
		return
	}

	// Get user info first
	userInfo, err := dbManager.GetUserByID(URLIDConv)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}

	// Get user's threads
	userThreads, err := dbManager.GetThreadsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user threads: %v\n", err)
		userThreads = []data.Thread{} // Empty slice if error
	}

	// Get user's posts
	userPosts, err := dbManager.GetPostsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user posts: %v\n", err)
		userPosts = []data.Post{} // Empty slice if error
	}

	// Get user's liked posts
	likedPosts, err := dbManager.GetLikedPostsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user liked posts: %v\n", err)
		likedPosts = []data.Post{} // Empty slice if error
	}

	// Get user's liked threads
	likedThreads, err := dbManager.GetLikedThreadsByUserID(URLIDConv)
	if err != nil {
		fmt.Printf("Error getting user liked threads: %v\n", err)
		likedThreads = []data.Thread{} // Empty slice if error
	}

	// Create account data structure that matches the template expectations
	// The template expects an array where the first element has user info and contains Cards/LikedPosts
	var templateData []data.Thread

	// Always create at least one element for the template
	var firstElement data.Thread

	if len(userThreads) > 0 {
		// Use the first thread as a base
		firstElement = userThreads[0]

		// Add the rest of the threads if any
		if len(userThreads) > 1 {
			templateData = append(templateData, userThreads[1:]...)
		}
	} else {
		// No threads, create a dummy thread with user info
		firstElement = data.Thread{
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
	templateData = append([]data.Thread{firstElement}, templateData...)

	// Use the utility function to generate HTML
	utils.GenerateHTML(writer, templateData, "layout", "private.navbar", "account", "accountbypost", "cookie-consent")
}

func AccountCheck(writer http.ResponseWriter, request *http.Request) {
	sess, err := data.SessionCheck(writer, request)
	if err != nil {
		fmt.Println("Session check error:", err)
		http.Redirect(writer, request, "/login", 302)
		return
	}

	// Check if session is valid
	valid, err := sess.Valid()
	if err != nil || !valid {
		fmt.Println("Invalid session")
		http.Redirect(writer, request, "/login", 302)
		return
	}

	// Get user from session
	user, err := sess.User()
	if err != nil {
		fmt.Println("Error getting user from session:", err)
		http.Redirect(writer, request, "/login", 302)
		return
	}

	// Redirect to account page with user ID
	http.Redirect(writer, request, fmt.Sprintf("/account?user_id=%v", user.Id), 302)
}
