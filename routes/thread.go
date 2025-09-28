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
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		http.Redirect(writer, request, "/login", 302)
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		http.Redirect(writer, request, "/login", 302)
		return
	}

	err := request.ParseForm()
	if err != nil {
		utils.Danger(err, "Cannot parse form")
	} else {
		topic := request.PostFormValue("topic")
		body := request.PostFormValue("body")
		selected := request.PostFormValue("selection1")
		selected2 := request.PostFormValue("selection2")

		if selected == selected2 {
			selected2 = ""
		}
		// Use CreateThreadByUser which accepts string categories
		idTo, err := dbManager.CreateThreadByUser(topic, body, currentUser.Id, selected, selected2)
		if err != nil {
			utils.Danger(err, "Cannot create thread")
		}

		http.Redirect(writer, request, "/thread/read?id="+strconv.Itoa(int(idTo)), http.StatusFound)
	}
}

// GET /thread/read
// show the details of the thread, including the posts and the form to write a post
func ReadThread(writer http.ResponseWriter, request *http.Request) {
	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.ErrorMessage(writer, request, "Database connection unavailable")
		return
	}

	thread := data.Thread{}
	var err error
	id := request.URL.Query().Get("id")
	resid, _ := strconv.Atoi(id)
	// userFrom := *GetCurrentUser(request)
	thread, err = dbManager.GetThreadWithPosts(resid)

	if err != nil {
		utils.ErrorMessage(writer, request, "Cannot read thread")
	} else {
		// Check authentication status to determine which template to use
		if IsAuthenticated(request) {
			utils.GenerateHTML(writer, &thread, "layout", "private.navbar", "private.thread")
		} else {
			utils.GenerateHTML(writer, &thread, "layout", "public.navbar", "public.thread")
		}
	}
}

func ThreadLikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Thread Like/Dislike POSTING processes started")
	
	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	URLID := request.URL.Query().Get("idlikes")
	URLIDConv, _ := strconv.Atoi(URLID)
	fmt.Print(URLID)
	fmt.Print(" /// ")
	
	likes, err := dbManager.GetThreadLikes(URLIDConv)
	if err != nil {
		http.Error(writer, "Cannot get thread likes", http.StatusInternalServerError)
		return
	}
	
	var props data.ThreadLikeProperties
	props.Li = likes
	
	dislikes, err := dbManager.GetThreadDislikes(URLIDConv)
	if err != nil {
		http.Error(writer, "Cannot get thread dislikes", http.StatusInternalServerError)
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
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}
	
	LIKEDID := request.URL.Query().Get("idlikes")
	postID2, _ := strconv.Atoi(LIKEDID)

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
		http.Error(writer, "Cannot get post likes", http.StatusInternalServerError)
		return
	}

	dislikes, err := dbManager.GetDislikes(postID2)
	if err != nil {
		http.Error(writer, "Cannot get post dislikes", http.StatusInternalServerError)
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
	fmt.Println("ApplyThreadLikes processes")
	
	// Get database manager from context
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		writer.Write([]byte("You are guest:)"))
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		writer.Write([]byte("You are guest:)"))
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in ApplyThreadLikes:", err)
		return
	}

	userId := request.PostFormValue("userUSER") // thread creator userid
	sook := request.PostFormValue("okay")       // thread id
	userId2, _ := strconv.Atoi(userId)
	threadID, _ := strconv.Atoi(sook)

	// Check if user is trying to like their own thread
	if currentUser.Id == userId2 {
		writer.Write([]byte("You are creator:)"))
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
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		writer.Write([]byte("You are guest:)"))
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		writer.Write([]byte("You are guest:)"))
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in ApplyThreadDislikes:", err)
		return
	}

	postID := request.PostFormValue("userUSER2") // thread creator userid
	sook := request.PostFormValue("okay2")       // thread id
	creatorID, _ := strconv.Atoi(postID)
	threadID, _ := strconv.Atoi(sook)

	// Check if user is trying to dislike their own thread
	if currentUser.Id == creatorID {
		writer.Write([]byte("You are creator:)"))
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
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		http.Redirect(writer, request, "/err?msg=You are guest, please login", 302)
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		http.Redirect(writer, request, "/err?msg=You are guest, please login", 302)
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in AcceptLike:", err)
		return
	}

	postID := request.PostFormValue("userUSER") // post id
	sook := request.PostFormValue("okay")       // post creator id
	postID2, _ := strconv.Atoi(postID)
	creatorID, _ := strconv.Atoi(sook)

	// Check if user is trying to like their own post
	if currentUser.Id == creatorID {
		http.Redirect(writer, request, "/err?msg=You cannot like your own post", 302)
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
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		http.Redirect(writer, request, "/err?msg=You are guest, please login", 302)
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		http.Redirect(writer, request, "/err?msg=You are guest, please login", 302)
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in AcceptDislike:", err)
		return
	}

	postID := request.PostFormValue("userUSER2") // post id
	sook := request.PostFormValue("okay2")       // post creator id
	postID2, _ := strconv.Atoi(postID)
	creatorID, _ := strconv.Atoi(sook)

	// Check if user is trying to dislike their own post
	if currentUser.Id == creatorID {
		http.Redirect(writer, request, "/err?msg=You cannot dislike your own post", 302)
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
		http.Error(writer, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	if !IsAuthenticated(request) {
		http.Redirect(writer, request, "/login", 302)
		return
	}

	currentUser := GetCurrentUser(request)
	if currentUser == nil {
		http.Redirect(writer, request, "/login", 302)
		return
	}

	err := request.ParseForm()
	if err != nil {
		utils.Danger(err, "Cannot parse form")
		return
	}

	body := request.PostFormValue("body")
	id := request.PostFormValue("id")
	threadID, _ := strconv.Atoi(id)
	
	thread, err := dbManager.GetThreadByID(threadID)
	if err != nil {
		utils.ErrorMessage(writer, request, "Cannot read thread")
		return
	}

	_, err = dbManager.CreatePost(thread.Id, body, currentUser.Id)
	if err != nil {
		utils.Danger(err, "Cannot create post")
		return
	}

	url := fmt.Sprint("/thread/read?id=", id)
	http.Redirect(writer, request, url, 302)
}
