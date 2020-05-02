package routes

import (
	"fmt"
	"forum/data"
	"forum/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func AccountCheck(writer http.ResponseWriter, request *http.Request) {
	// var alsoid int
	fmt.Println("OK")
	cook, err := request.Cookie("_cookie")
	if err != nil {
		fmt.Println("Error") // or redirect
	}
	cooPart := strings.Split(cook.Value, "&")
	http.Redirect(writer, request, fmt.Sprintf("/account?user_id=%v", cooPart[0]), 302)
}

// GET /threads/new
// show the new thread form page
func NewThread(writer http.ResponseWriter, request *http.Request) {
	_, err := data.SessionCheck(writer, request)
	// MyId := data.GetCookieValue(request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		utils.GenerateHTML(writer, nil, "layout", "private.navbar", "new.thread")
	}
}

// POST /signup
// create the user account
func CreateThread(writer http.ResponseWriter, request *http.Request) {
	sess, err := data.SessionCheck(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			utils.Danger(err, "Cannot parse form")
		}
		selected := request.PostFormValue("selection")
		user, err := sess.User()
		if err != nil {
			utils.Danger(err, "Cannot get user from session")
		}
		alsoid := data.GetCookieValue(request)
		topic := request.PostFormValue("topic")
		ourID, _, err := user.CreateThread(topic, alsoid, selected)
		if err != nil {
			utils.Danger(err, "Cannot create thread")
		}

		err = data.LikeOnThreadCreation(alsoid, int(ourID)) // only with PostId
		err = data.DislikeOnThreadCreation(alsoid, int(ourID))
		if err != nil {
			utils.Danger(err, "Cannot Create Like/Dislike on Post Creation")
		}
		http.Redirect(writer, request, "/", 302)
	}
}

// GET /thread/read
// show the details of the thread, including the posts and the form to write a post
func ReadThread(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	resid, _ := strconv.Atoi(id)
	thread, err := data.ThreadById(resid)
	if err != nil {
		utils.ErrorMessage(writer, request, "Cannot read thread")
	} else {
		_, err := data.SessionCheck(writer, request)
		if err != nil {
			utils.GenerateHTML(writer, &thread, "layout", "public.navbar", "public.thread")
		} else {
			utils.GenerateHTML(writer, &thread, "layout", "private.navbar", "private.thread")
		}
	}
}

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
	getFinal, _ := data.GetFromLikedDB(likedposts)
	thread[0].Cards = posts
	// fmt.Println(getFinal)
	thread[0].LikedPosts = getFinal
	filesFrom := []string{"layout", "private.navbar", "account", "accountbypost"}
	var files []string
	for _, file := range filesFrom {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(writer, "layout", &thread)

}

func ThreadLikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Thread Like/Dislike POSTING procceses started")
	URLID := request.URL.Query().Get("idlikes")
	URLIDConv, _ := strconv.Atoi(URLID)
	fmt.Print(URLID)
	fmt.Print(" /// ")
	likes := data.GetThreadLikes(URLIDConv)
	// fmt.Println(likes)
	var props data.ThreadLikeProperties
	props.Li = likes
	dislikes := data.GetThreadDislikes(URLIDConv)
	fmt.Println(dislikes)
	props.Di = dislikes
	templates := template.Must(template.ParseFiles("templates/threadLikes.html"))
	templates.Execute(writer, &props)
}

// POST /likes
func PostLike(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Post Like procceses ")
	LIKEDID := request.URL.Query().Get("idlikes")
	postID2, _ := strconv.Atoi(LIKEDID)
	// fmt.Println("also likes proccess started")
	// fmt.Println(likes)

	alsoid := data.GetCookieValue(request)
	likes := data.GetLikes(postID2)

	dislikes := data.GetDislikes(postID2)

	var props data.LikeProperties
	props.Li = likes
	props.Di = dislikes
	for i, g := range props.Li {
		if g.UserId == alsoid {
			props.Li[i].UserLiked = true
		}
	}
	for i, g := range props.Di {
		if g.UserId == alsoid {
			props.Di[i].UserDisliked = true
		}
	}
	templates := template.Must(template.ParseFiles("templates/likes.html"))
	templates.Execute(writer, &props)
}

func ApplyThreadLikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("ApplyThreadLikes procceses")
	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in AcceptLike")
	}

	alsoid := data.GetCookieValue(request)
	user := data.GetUserById(alsoid)

	userId := request.PostFormValue("userUSER") // userid
	sook := request.PostFormValue("okay")       // postid/threadid
	userId2, _ := strconv.Atoi(userId)
	sook2, _ := strconv.Atoi(sook)
	// data.ApplyThreadLike(user.Name, alsoid, sook2)

	if alsoid == -1 {
		writer.Write([]byte("You are guest:)"))
	} else if alsoid == sook2 {
		writer.Write([]byte("You are creator:)"))
	} else {
		if data.PrepareThreadLikedPosts(alsoid, userId2) {
			data.DeleteThreadLikes(alsoid, userId2)
		} else {
			if data.PrepareThreadDislikedPosts(alsoid, userId2) {
				data.DeleteThreadDislikes(alsoid, userId2)
			}
			data.ApplyThreadLike(user.Name, alsoid, userId2)
		}
	}

	ourstr := fmt.Sprintf("/threadLikes?idlikes=%v", userId) // lol thread or user id
	http.Redirect(writer, request, ourstr, 302)
}

func ApplyThreadDislikes(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Thread Dislike procceses")
	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in AcceptLike")
	}

	alsoid := data.GetCookieValue(request)
	user := data.GetUserById(alsoid)

	postID := request.PostFormValue("userUSER2") // userid
	sook := request.PostFormValue("okay2")       // thread
	postID2, _ := strconv.Atoi(postID)
	fmt.Println(postID2)
	sook2, _ := strconv.Atoi(sook)
	// data.ApplyThreadDislike(user.Name, alsoid, sook2)

	if alsoid == sook2 {
		writer.Write([]byte("You are creator:)"))
	} else {
		if data.PrepareThreadDislikedPosts(alsoid, postID2) {
			data.DeleteThreadDislikes(alsoid, postID2)
		} else {
			if data.PrepareThreadLikedPosts(alsoid, postID2) {
				data.DeleteThreadLikes(alsoid, postID2)
			}
			data.ApplyThreadDislike(user.Name, alsoid, postID2)
		}
	}
	ourstr := fmt.Sprintf("/threadLikes?idlikes=%v", postID) // lol thread or user id
	http.Redirect(writer, request, ourstr, 302)
}

func AcceptLike(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("handler AcceptLike just started.")
	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in AcceptLike")
	}
	alsoid := data.GetCookieValue(request)
	user := data.GetUserById(alsoid)

	postID := request.PostFormValue("userUSER")
	sook := request.PostFormValue("okay")
	postID2, _ := strconv.Atoi(postID)
	sook2, _ := strconv.Atoi(sook)
	// eqLike := false
	if alsoid == sook2 {
		writer.Write([]byte("You are creator:)"))
	} else {
		if data.PrepareLikedPosts(alsoid, postID2) {
			data.DeleteLikes(alsoid, postID2)
		} else {
			if data.PrepareDislikedPosts(alsoid, postID2) {
				data.DeleteDislikes(alsoid, postID2)
			}
			data.ApplyLikes(user.Name, alsoid, postID2)
		}
	}
	ourstr := fmt.Sprintf("/likes?idlikes=%v", postID)

	http.Redirect(writer, request, ourstr, 302)
}

func AcceptDislike(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("handler AcceptDislike just started.")
	err := request.ParseForm()
	if err != nil {
		log.Println("[ERROR] in AcceptDislike")
	}
	alsoid := data.GetCookieValue(request)
	user := data.GetUserById(alsoid)

	postID := request.PostFormValue("userUSER2")
	sook := request.PostFormValue("okay2")
	postID2, _ := strconv.Atoi(postID)
	sook2, _ := strconv.Atoi(sook)
	if alsoid == sook2 {
		writer.Write([]byte("You are creator:)"))
	} else {
		if data.PrepareDislikedPosts(alsoid, postID2) {
			data.DeleteDislikes(alsoid, postID2)
		} else {
			if data.PrepareLikedPosts(alsoid, postID2) {
				data.DeleteLikes(alsoid, postID2)
			}
			data.ApplyDislikes(user.Name, alsoid, postID2)
		}
	}
	ourstr := fmt.Sprintf("/likes?idlikes=%v", postID)

	http.Redirect(writer, request, ourstr, 302)
}

// POST /thread/post
// create the post
func PostThread(writer http.ResponseWriter, request *http.Request) {
	sess, err := data.SessionCheck(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			utils.Danger(err, "Cannot parse form")
		}
		user, err := sess.User()
		if err != nil {
			utils.Danger(err, "Cannot get user from session")
		}
		body := request.PostFormValue("body")
		// fmt.Println(body)
		id := request.PostFormValue("id")
		resid, _ := strconv.Atoi(id)
		thread, err := data.ThreadById(resid)
		if err != nil {
			utils.ErrorMessage(writer, request, "Cannot read thread")
		}

		alsoid := data.GetCookieValue(request)

		ourID, err := user.CreatePost(thread, body, alsoid)
		fmt.Println(ourID)
		if err != nil {
			utils.Danger(err, "Cannot create post")
		}

		err = data.LikeOnPostCreation(alsoid, int(ourID)) // only with PostId
		err = data.DislikeOnPostCreation(alsoid, int(ourID))
		if err != nil {
			utils.Danger(err, "Cannot Create Like/Dislike on Post Creation")
		}
		url := fmt.Sprint("/thread/read?id=", id)
		http.Redirect(writer, request, url, 302)
	}
}
