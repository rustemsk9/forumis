package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"forum/internal"
	"forum/utils"
)

func ReadThreadsFromAccount(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		utils.MethodNotAllowed(writer, request, "POST method not allowed")
		return
	} else if request.Method == "GET" {
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

		templateData, err := internal.ReadDMFromAccount(URLIDConv)
		if err != nil {
			utils.BadRequest(writer, request, "Failed to from Account")
			return
		}
		// Use the utility function to generate HTML
		utils.GenerateHTML(writer, templateData, "layout", "private.navbar", "account", "accountbypost", "cookie-consent")
	}
}

func AccountCheck(writer http.ResponseWriter, request *http.Request) {
	sess, err := SessionCheck(writer, request)
	if err != nil || sess.Uuid == "" {
		fmt.Println("Session check error:", err)
		http.Redirect(writer, request, "/login", 302)
		return
	}

	// Redirect to account page with user ID
	http.Redirect(writer, request, fmt.Sprintf("/account?user_id=%v", sess.UserId), 302)
}
