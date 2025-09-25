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
	}
	URLID := request.URL.Query().Get("user_id")
	URLIDConv, _ := strconv.Atoi(URLID)
	fmt.Println(URLIDConv)
	// and user template define content here
	thread, err := data.AccountThreads(URLIDConv)
	if err != nil {
		fmt.Println("Error on routes ReadAccount")
		http.Redirect(writer, request, "/", 302)
		return
	}

	userINFO := data.GetUserById(URLIDConv)
	if thread == nil {
		utils.GenerateHTML(writer, &userINFO, "layout", "private.navbar", "justaccount")
		// http.Redirect(writer, request, "/", 302)
		return
	}

	posts, _ := data.GetUserPosts(URLIDConv)
	likedposts, _ := data.GetUserLikedPosts(URLIDConv)

	fmt.Println(likedposts)
	getFinal, _ := data.GetLikesPostsFromDB(likedposts)
	thread[0].Cards = posts
	// fmt.Println(getFinal)
	thread[0].LikedPosts = getFinal
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
