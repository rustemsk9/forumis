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

	// // Get database manager from context
	// dbManager := GetDatabaseManager(request)
	// if dbManager == nil {
	// 	utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
	// 	return
	// }

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

	templateData, err := data.ReadDMFromAccount(URLIDConv)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}
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
