package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"forum/data"
	"forum/utils"
)

func ReadThreadsFromAccount(writer http.ResponseWriter, request *http.Request) {
	_, err := data.SessionCheck(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
		return
	}
	
	URLID := request.URL.Query().Get("user_id")
	URLIDConv, _ := strconv.Atoi(URLID)
	fmt.Println(URLIDConv)
	
	// Get user threads
	thread, err := data.AccountThreads(URLIDConv)
	if err != nil {
		fmt.Println("Error on routes ReadAccount:", err)
		http.Redirect(writer, request, "/", 302)
		return
	}

		likedposts, err := data.GetUserLikedPosts(URLIDConv)
	if err != nil {
		fmt.Println("Error getting user liked posts:", err)
	}

	userINFO := data.GetUserById(URLIDConv)
	if len(thread) == 0 && len(likedposts) == 0 {
		utils.GenerateHTML(writer, &userINFO, "layout", "private.navbar", "justaccount")
		return
	}

	// Get user posts and liked posts
	posts, err := data.GetUserPosts(URLIDConv)
	if err != nil {
		fmt.Println("Error getting user posts:", err)
	}
	


	fmt.Println("Liked posts IDs:", likedposts)
	getFinal, err := data.GetLikesPostsFromDB(likedposts)
	if err != nil {
		fmt.Println("Error getting liked posts details:", err)
	}
	
	// Set posts and liked posts on the first thread (for template compatibility)
	if len(thread) > 0 {
		thread[0].Cards = posts
		thread[0].LikedPosts = getFinal
	} else {
		// Create a dummy thread entry if no threads exist, but user has posts or liked posts
		if len(posts) > 0 || len(getFinal) > 0 {
			dummyThread := data.Thread{
				Cards:      posts,
				LikedPosts: getFinal,
				User:       userINFO.Name, // Use the user info we already have
				Email:     userINFO.Email,
			}
			thread = append(thread, dummyThread)
		}
	}
	filesFrom := []string{"layout", "private.navbar", "account", "accountbypost", "cookie-consent"}
	var files []string
	for _, file := range filesFrom {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}

	fmt.Println("Templates files are: ", files)
	templates, err := template.ParseFiles(files...)
	if err != nil {
		fmt.Println("Template parsing error:", err)
		http.Error(writer, "Template parsing error", http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(writer, "layout", &thread)
	if err != nil {
		fmt.Println("Template execution error:", err)
		http.Error(writer, "Template execution error", http.StatusInternalServerError)
		return
	}
	fmt.Println("Template executed successfully!")
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
