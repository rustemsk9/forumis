package routes

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"forum/data"
	"forum/utils"
)

// GET /threads/new
// show the new thread form page
func NewThread(writer http.ResponseWriter, request *http.Request) {
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
	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
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
	idTo, err := dbManager.CreateThreadByUser(topic, body, currentUser.Id, selected, selected2)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	http.Redirect(writer, request, "/thread/read?id="+strconv.Itoa(int(idTo)), http.StatusFound)
}

// GET /thread/read
// show the details of the thread, including the posts and the form to write a post
func ReadThread(writer http.ResponseWriter, request *http.Request) {
	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	thread := data.Thread{}
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

	thread, err = dbManager.GetThreadWithPosts(resid)
	if err != nil {
		// Check if it's a "not found" error vs other database errors
		if err.Error() == "sql: no rows in result set" {
			utils.NotFound(writer, request)
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

func ThreadLikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Thread Like/Dislike POSTING processes started")

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	URLID := request.URL.Query().Get("idlikes")
	if URLID == "" {
		utils.BadRequest(writer, request, "Thread ID is required")
		return
	}

	URLIDConv, err := strconv.Atoi(URLID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	fmt.Print(URLID)
	fmt.Print(" /// ")

	likes, err := dbManager.GetThreadLikes(URLIDConv)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	var props data.ThreadLikeProperties
	props.Li = likes

	dislikes, err := dbManager.GetThreadDislikes(URLIDConv)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	fmt.Println(dislikes)
	props.Di = dislikes
	templates := template.Must(template.ParseFiles("templates/threadLikes.html"))
	templates.Execute(writer, &props)
}

// POST /likes
func PostLike(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Post Like processes ")

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	LIKEDID := request.URL.Query().Get("idlikes")
	if LIKEDID == "" {
		utils.BadRequest(writer, request, "Post ID is required")
		return
	}

	postID2, err := strconv.Atoi(LIKEDID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid post ID format")
		return
	}

	// Get current user ID for checking likes
	var currentUserID int = -1
	if IsAuthenticated(request) {
		user := GetCurrentUser(request)
		if user != nil {
			currentUserID = user.Id
		}
	}

	likes, err := dbManager.GetLikes(postID2)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	dislikes, err := dbManager.GetDislikes(postID2)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	var props data.LikeProperties
	props.Li = likes
	props.Di = dislikes

	for i, g := range props.Li {
		if g.UserId == currentUserID {
			props.Li[i].UserLiked = true
		}
	}
	for i, g := range props.Di {
		if g.UserId == currentUserID {
			props.Di[i].UserDisliked = true
		}
	}

	templates := template.Must(template.ParseFiles("templates/likes.html"))
	templates.Execute(writer, &props)
}

func ApplyThreadLikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Thread Like process")

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
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

	postID := request.PostFormValue("userUSER2") // thread creator userid
	sook := request.PostFormValue("okay2")       // thread id

	if postID == "" || sook == "" {
		utils.BadRequest(writer, request, "Missing required parameters")
		return
	}

	creatorID, err := strconv.Atoi(postID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid creator ID format")
		return
	}

	threadID, err := strconv.Atoi(sook)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Check if user is trying to like their own thread
	if currentUser.Id == creatorID {
		utils.BadRequest(writer, request, "You cannot like your own thread")
		return
	}

	// Handle the like logic
	if dbManager.PrepareThreadLikedPosts(currentUser.Id, threadID) {
		// User already liked, remove the like
		err = dbManager.DeleteThreadLikes(currentUser.Id, threadID)
		if err != nil {
			log.Println("[ERROR] deleting thread like:", err)
		}
	} else {
		// Check if user disliked this thread, if so remove dislike first
		if dbManager.PrepareThreadDislikedPosts(currentUser.Id, threadID) {
			err = dbManager.DeleteThreadDislikes(currentUser.Id, threadID)
			if err != nil {
				log.Println("[ERROR] deleting thread dislike:", err)
			}
		}
		// Add the like
		err = dbManager.ApplyThreadLike(currentUser.Id, threadID)
		if err != nil {
			log.Println("[ERROR] applying thread like:", err)
		}
	}

	ourstr := fmt.Sprintf("/threadLikes?idlikes=%v", threadID)
	http.Redirect(writer, request, ourstr, 302)
}

func ApplyThreadDislikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Thread Dislike processes")

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
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

	postID := request.PostFormValue("userUSER2") // thread creator userid
	sook := request.PostFormValue("okay2")       // thread id

	if postID == "" || sook == "" {
		utils.BadRequest(writer, request, "Missing required parameters")
		return
	}

	creatorID, err := strconv.Atoi(postID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid creator ID format")
		return
	}

	threadID, err := strconv.Atoi(sook)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Check if user is trying to dislike their own thread
	if currentUser.Id == creatorID {
		utils.BadRequest(writer, request, "You cannot dislike your own thread")
		return
	}

	// Handle the dislike logic
	if dbManager.PrepareThreadDislikedPosts(currentUser.Id, threadID) {
		// User already disliked, remove the dislike
		err = dbManager.DeleteThreadDislikes(currentUser.Id, threadID)
		if err != nil {
			log.Println("[ERROR] deleting thread dislike:", err)
		}
	} else {
		// Check if user liked this thread, if so remove like first
		if dbManager.PrepareThreadLikedPosts(currentUser.Id, threadID) {
			err = dbManager.DeleteThreadLikes(currentUser.Id, threadID)
			if err != nil {
				log.Println("[ERROR] deleting thread like:", err)
			}
		}
		// Add the dislike
		err = dbManager.ApplyThreadDislike(currentUser.Id, threadID)
		if err != nil {
			log.Println("[ERROR] applying thread dislike:", err)
		}
	}

	ourstr := fmt.Sprintf("/threadLikes?idlikes=%v", threadID)
	http.Redirect(writer, request, ourstr, 302)
}

func AcceptLike(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("handler AcceptLike just started.")

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
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

	postID := request.PostFormValue("userUSER") // post id
	sook := request.PostFormValue("okay")       // post creator id

	if postID == "" || sook == "" {
		utils.BadRequest(writer, request, "Missing required parameters")
		return
	}

	postID2, err := strconv.Atoi(postID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid post ID format")
		return
	}

	// Handle the like logic
	if dbManager.PrepareLikedPosts(currentUser.Id, postID2) {
		// User already liked, remove the like
		err = dbManager.DeleteLikes(currentUser.Id, postID2)
		if err != nil {
			log.Println("[ERROR] deleting like:", err)
		}
	} else {
		// Check if user disliked this post, if so remove dislike first
		if dbManager.PrepareDislikedPosts(currentUser.Id, postID2) {
			err = dbManager.DeleteDislikes(currentUser.Id, postID2)
			if err != nil {
				log.Println("[ERROR] deleting dislike:", err)
			}
		}
		// Add the like
		err = dbManager.ApplyLikes(currentUser.Id, postID2)
		if err != nil {
			log.Println("[ERROR] applying like:", err)
		}
	}

	ourstr := fmt.Sprintf("/likes?idlikes=%v", postID)
	http.Redirect(writer, request, ourstr, 302)
}

func AcceptDislike(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("handler AcceptDislike just started.")

	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
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

	postID := request.PostFormValue("userUSER2") // post id
	sook := request.PostFormValue("okay2")       // post creator id

	if postID == "" || sook == "" {
		utils.BadRequest(writer, request, "Missing required parameters")
		return
	}

	postID2, err := strconv.Atoi(postID)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid post ID format")
		return
	}

	creatorID, err := strconv.Atoi(sook)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid creator ID format")
		return
	}

	// Check if user is trying to dislike their own post
	if currentUser.Id == creatorID {
		utils.BadRequest(writer, request, "You cannot dislike your own post")
		return
	}

	// Handle the dislike logic
	if dbManager.PrepareDislikedPosts(currentUser.Id, postID2) {
		// User already disliked, remove the dislike
		err = dbManager.DeleteDislikes(currentUser.Id, postID2)
		if err != nil {
			log.Println("[ERROR] deleting dislike:", err)
		}
	} else {
		// Check if user liked this post, if so remove like first
		if dbManager.PrepareLikedPosts(currentUser.Id, postID2) {
			err = dbManager.DeleteLikes(currentUser.Id, postID2)
			if err != nil {
				log.Println("[ERROR] deleting like:", err)
			}
		}
		// Add the dislike
		err = dbManager.ApplyDislikes(currentUser.Id, postID2)
		if err != nil {
			log.Println("[ERROR] applying dislike:", err)
		}
	}

	ourstr := fmt.Sprintf("/likes?idlikes=%v", postID)
	http.Redirect(writer, request, ourstr, 302)
}

// POST /thread/post
// create the post
func PostThread(writer http.ResponseWriter, request *http.Request) {
	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
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

	thread, err := dbManager.GetThreadByID(threadID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.NotFound(writer, request)
		} else {
			utils.InternalServerError(writer, request, err)
		}
		return
	}

	_, err = dbManager.CreatePost(thread.Id, body, currentUser.Id)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	url := fmt.Sprint("/thread/read?id=", id)
	http.Redirect(writer, request, url, 302)
}
