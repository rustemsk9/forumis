package routes

import (
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
	
	// Check if this is a POST request with filter parameters
	if request.Method == "POST" {
		err = request.ParseForm()
		if err == nil {
			category1 := request.PostFormValue("selection1")
			category2 := request.PostFormValue("selection2")
			threads, err = data.FilterThreadsByCategories(category1, category2)
		}
	} else {
		// GET request - show all threads using database manager
		threads, err = dbManager.GetAllThreads()
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
		}{
			Threads: threads,
			Title:   "Forum Home",
			Message: "Welcome to the Forum",
			User:    userName,
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

		// Use middleware authentication check
		if IsAuthenticated(request) {
			utils.GenerateHTML(writer, pageData, "layout", "private.navbar", "index")
		} else {
			utils.GenerateHTML(writer, pageData, "layout", "public.navbar", "index")
		}
	} else {
		utils.ErrorMessage(writer, request, "Cannot get threads")
	}
}
