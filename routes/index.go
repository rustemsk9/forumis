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
	if request.Method != "GET" {
		utils.MethodNotAllowed(writer, request, "GET method only")
		return
	}
	vals := request.URL.Query()

	// Use middleware to check authentication
	if IsAuthenticated(request) {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "private.navbar", "error")
	} else {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "public.navbar", "error")
	}
}

func Index(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		utils.NotFound(writer, request)
		return
	}
	var threads []data.Thread
	var err error
	category1, category2 := "", ""
	sortBy := ""
	reset := "false"
	catbool := false
	var user *data.User
	var thread data.Thread
	// Check if this is a POST request with filter parameters
	switch request.Method {
	case "POST":
		fmt.Println("POST request received for filtering/sorting")
		err = request.ParseForm()
		if err == nil {
			category1 = request.PostFormValue("selection1")
			category2 = request.PostFormValue("selection2")
			sortBy = request.URL.Query().Get("sort")

			// Save user preferences if user is authenticated
			if IsAuthenticated(request) {
				user = GetCurrentUser(request)
				if user != nil {
					err := user.UpdateUserPreferences(user.Id, category1, category2)
					if err != nil {
						fmt.Printf("Error updating user preferences: %v\n", err)
					}
				}
			}

			// Handle sorting and filtering
			if category1 != "" || category2 != "" {
				fmt.Println("Filtering by categories:", category1, category2)
				threads, err = thread.FilterThreadsByCategories(category1, category2)
				catbool = true

			} else {
				fmt.Println("No filters applied, showing all threads")
				threads, err = thread.GetAllThreads()
				catbool = false
			}

			if sortBy == "most_liked" {
				fmt.Println("Sorting by most liked")
				threads, err = data.SortThreadsByLikesDesc(threads)
			} else if sortBy == "latest" {
				threads, err = data.SortThreadsByLatest(threads)
			}
			if err != nil {
				fmt.Println("Error retrieving threads:", err)
				utils.ErrorMessage(writer, request, "Error retrieving threads")
				return
			}

		}
	case "GET":
		// GET request - check for sort parameter in URL

		// Load user preferred categories for authenticated users
		userIdFind := -1
		if IsAuthenticated(request) {
			user = GetCurrentUser(request)
			if user != nil {
				userIdFind = user.Id
				// Get fresh user data to ensure we have preferred categories
				fullUser := user.GetUserById(user.Id)
				category1 = fullUser.PreferedCategory1
				category2 = fullUser.PreferedCategory2
				if category1 != "" || category2 != "" {
					catbool = true
				}

			}
		}

		err = request.ParseForm()
		if err == nil {
			sortBy = request.URL.Query().Get("sort")
			reset = request.URL.Query().Get("reset")
		} else {
			utils.BadRequest(writer, request, "Cannot parse form data")
			return
		}

		if reset == "true" {
			threads, err = thread.GetAllThreads()
			catbool = true
			category1, category2 = "", ""
			sortBy = ""

			if userIdFind != -1 {
				err := user.UpdateUserPreferences(userIdFind, category1, category2)
				if err != nil {
					fmt.Printf("Error updating user preferences: %v\n", err)
				}
			}
		} else {

			if category1 != "" || category2 != "" {
				threads, _ = thread.FilterThreadsByCategories(category1, category2)
				catbool = true
			} else {
				fmt.Println("No filters applied, showing all threads GET")
				threads, err = thread.GetAllThreads()
				catbool = false
			}

			if sortBy == "most_liked" {
				threads, err = data.SortThreadsByLikesDesc(threads)
			} else if sortBy == "latest" {
				threads, err = data.SortThreadsByLatest(threads)
			}
			if err != nil {
				utils.InternalServerError(writer, request, err)
				return
			}
		}
	default:
		utils.BadRequest(writer, request, "Unsupported request method")
		return
	}

	if err == nil {
		// Get current user from middleware
		user = GetCurrentUser(request)
		userName := ""
		if user != nil {
			userName = user.Name
			// Populate user-specific vote information for each thread
			for i := range threads {
				threads[i].UserLiked = user.HasThreadLiked(user.Id, threads[i].Id)
				threads[i].UserDisliked = user.HasThreadDisliked(user.Id, threads[i].Id)
			}
		}

		// Create expanded data structure
		pageData := struct {
			Threads           []data.Thread
			Title             string
			Message           string
			User              string
			Count             int
			Online            int
			PreferedCategory1 string
			PreferedCategory2 string
			SortBy            string
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
				online, err := data.CheckOnlineUsers(10)
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
		utils.NotFound(writer, request)
	}
}
