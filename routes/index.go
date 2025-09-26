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
	_, err := data.SessionCheck(writer, request)
	if err != nil {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "public.navbar", "error")
	} else {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "private.navbar", "error")
	}
}

func Index(writer http.ResponseWriter, request *http.Request) {
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
		// GET request - show all threads
		threads, err = data.Threads()
	}

	if err == nil {
		// Create expanded data structure
		user, _ := data.CurrentUser(request)
		pageData := struct {
			Threads []data.Thread
			Title   string
			Message string
			User    string
			Count   int
			Online int
			// Users  []string
		}{
			Threads: threads,
			Title:   "Forum Home",
			Message: "Welcome to the Forum",
			User:    user,
			Count: func() int {
				count, err := data.UserCount()
				if err != nil {
					return 0
				}
				return count
			}(),
			Online: func() int {
				online, err := data.CheckOnlineUsers(10)
				if err != nil {
					return 0
				}
				return len(online)
			}(),
		}
		
		

		_, err := data.SessionCheck(writer, request)
		if err != nil {
			utils.GenerateHTML(writer, pageData, "layout", "public.navbar", "index")
		} else {
			// You could add user data here if available
			utils.GenerateHTML(writer, pageData, "layout", "private.navbar", "index")
		}
	} else {
		utils.ErrorMessage(writer, request, "Cannot get threads")
	}
}
