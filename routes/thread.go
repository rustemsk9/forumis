package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"forum/internal"
	"forum/models"
	"forum/utils"
)

// GET /threads/new
// show the new thread form page
func NewThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		utils.MethodNotAllowed(writer, request, "GET method only")
		return
	}
	// Check if user is authenticated via middleware
	if !IsAuthenticated(request) {
		http.Redirect(writer, request, "/login", http.StatusFound)
		return
	}

	utils.GenerateHTML(writer, nil, "layout", "private.navbar", "new.thread")
}

// POST /thread/create
// create the thread
func CreateThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		utils.MethodNotAllowed(writer, request, "POST method only")
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	err := request.ParseForm()
	if err != nil {
		utils.BadRequest(writer, request, "Cannot parse form data")
		return
	}

	topic := request.PostFormValue("topic")
	body := request.PostFormValue("body")
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.ReplaceAll(body, "\\n", "\n")
	selected := request.PostFormValue("selection1")
	selected2 := request.PostFormValue("selection2")

	// Validate required fields
	if topic == "" {
		utils.BadRequest(writer, request, "Thread topic is required")
		return
	}
	if body == "" {
		utils.BadRequest(writer, request, "Thread body is required")
		return
	}

	if selected == selected2 {
		selected2 = ""
	}

	// Use CreateThreadByUser which accepts string categories
	idTo, err := internal.CrThreadByUser(topic, body, currentUser.Id, selected, selected2)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	http.Redirect(writer, request, "/thread/read?id="+strconv.Itoa(int(idTo)), http.StatusFound)
}

// GET /thread/read
// show the details of the thread, including the posts and the form to write a post
func ReadThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		utils.MethodNotAllowed(writer, request, "GET method only")
		return
	}

	thread := models.Thread{}
	var err error
	id := request.URL.Query().Get("id")

	// Validate thread ID
	if id == "" {
		utils.BadRequest(writer, request, "Thread ID is required")
		return
	}

	resid, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	thread, err = internal.ThreadWithPosts(resid)
	if err != nil {
		// Check if it's a "not found" error vs other database errors
		if err.Error() == "sql: no rows in result set" {
			utils.BadRequest(writer, request, "Thread not found")
		} else {
			utils.InternalServerError(writer, request, err)
		}
		return
	}

	// Check authentication status to determine which template to use
	if IsAuthenticated(request) {
		utils.GenerateHTML(writer, &thread, "layout", "private.navbar", "private.thread")
	} else {
		utils.GenerateHTML(writer, &thread, "layout", "public.navbar", "public.thread")
	}
}

// POST /thread/post
// create the post
func PostThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		utils.MethodNotAllowed(writer, request, "POST method only")
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	currentUser := GetCurrentUser(request)

	err := request.ParseForm()
	if err != nil {
		utils.BadRequest(writer, request, "Cannot parse form data")
		return
	}

	body := request.PostFormValue("body")
	id := request.PostFormValue("id")

	// Validate required fields
	if body == "" {
		utils.BadRequest(writer, request, "Post body is required")
		return
	}
	if id == "" {
		utils.BadRequest(writer, request, "Thread ID is required")
		return
	}

	threadID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	thread, err := internal.ThreadById(threadID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.NotFound(writer, request)
		} else {
			utils.InternalServerError(writer, request, err)
		}
		return
	}

	_, err = internal.CreatePost(thread.Id, body, currentUser.Id)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	url := fmt.Sprint("/thread/read?id=", id)
	http.Redirect(writer, request, url, 302)
}
