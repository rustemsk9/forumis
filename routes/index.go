package routes

import (
	"fmt"
	"net/http"

	"forum/data"
	"forum/utils"
)

// GET /err?msg=
// shows the error message page
func Err(writer http.ResponseWriter, request *http.Request) {
	vals := request.URL.Query()
	
	// Use middleware to check authentication
	if IsAuthenticated(request) {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "private.navbar", "error")
	} else {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "public.navbar", "error")
	}
}

func Index(writer http.ResponseWriter, request *http.Request) {
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.ErrorMessage(writer, request, "Database not available")
		return
	}

	var threads []data.Thread
	var err error
	category1, category2 := "", ""
	sortBy := ""
	catbool := false
	// Check if this is a POST request with filter parameters
	if request.Method == "POST" {
		err = request.ParseForm()
		if err == nil {
			category1 = request.PostFormValue("selection1")
			category2 = request.PostFormValue("selection2")
			sortBy = request.PostFormValue("sort_by")

			// Save user preferences if user is authenticated
			if IsAuthenticated(request) {
				user := GetCurrentUser(request)
				if user != nil {
					err := dbManager.UpdateUserPreferences(user.Id, category1, category2)
					if err != nil {
						fmt.Printf("Error updating user preferences: %v\n", err)
					}
				}
			}
			
			// Handle sorting and filtering
			if category1 != "" || category2 != "" {
				threads, err = data.FilterThreadsByCategories(category1, category2)
			} else if sortBy == "likes" {
				threads, err = dbManager.GetAllThreadsByLikes()
			} else {
				threads, err = dbManager.GetAllThreads()
			}
		}
	} else {
		// GET request - check for sort parameter in URL
		
		// Load user preferred categories for authenticated users
		if IsAuthenticated(request) {
			user := GetCurrentUser(request)
			if user != nil {
				// Get fresh user data to ensure we have preferred categories
				fullUser, userErr := dbManager.GetUserByID(user.Id)
				if userErr == nil {
					category1 = fullUser.PreferedCategory1
					category2 = fullUser.PreferedCategory2
					if category1 != "" || category2 != "" {
						catbool = true
					}
				}
			}
		}

		err = request.ParseForm()
		if err == nil {
			// category1 = request.PostFormValue("selection1")
			// category2 = request.PostFormValue("selection2")
			sortBy = request.PostFormValue("sort_by")
		}
		if category1 != "" || category2 != "" {
			threads, _ = data.FilterThreadsByCategories(category1, category2)
			catbool = true
			
		} else if sortBy == "likes" {
			threads, _ = dbManager.GetAllThreadsByLikes()
		} else {
			threads, _ = dbManager.GetAllThreads()
		}
		
	}

	if err == nil {
		// Get current user from middleware
		user := GetCurrentUser(request)
		userName := ""
		if user != nil {
			userName = user.Name
		}

		// Create expanded data structure
		pageData := struct {
			Threads []data.Thread
			Title   string
			Message string
			User    string
			Count   int
			Online  int
			PreferedCategory1 string
			PreferedCategory2 string
			SortBy string
		}{
			Threads: threads,
			Title:   "Forum Home",
			Message: "Welcome to the Forum",
			User:    userName,
			SortBy:  sortBy,
			Count: func() int {
				count, err := data.UserCount()
				if err != nil {
					return 0
				}
				return count
			}(),
			Online: func() int {
				online, err := dbManager.CheckOnlineUsers(10)
				if err != nil {
					return 0
				}
				return len(online)
			}(),
		}

		if catbool {
			pageData.PreferedCategory1 = category1
			pageData.PreferedCategory2 = category2
		}

		// Use middleware authentication check
		if IsAuthenticated(request) {
			utils.GenerateHTML(writer, pageData, "layout", "private.navbar", "index")
		} else {
			utils.GenerateHTML(writer, pageData, "layout", "public.navbar", "index")
		}
	} else {
		utils.ErrorMessage(writer, request, "Empty forum. No threads found.")
	}
}
