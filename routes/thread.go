package routes

import (
	"casual-talk/data"
	"casual-talk/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func AccountCheck(writer http.ResponseWriter, request *http.Request) {
	// var alsoid int
	fmt.Println("OK")
	cook, err := request.Cookie("_cookie")
	if err != nil {
		fmt.Println("Error") // or redirect
	}

	http.Redirect(writer, request, fmt.Sprintf("/account?user_id=%v", cook.Value), 302)
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
		var alsoid int
		cook, err := request.Cookie("_cookie")
		if err != nil {
			fmt.Println("Error") // or redirect
		}

		alsoid, _ = strconv.Atoi(cook.Value)
		topic := request.PostFormValue("topic")
		if _, err := user.CreateThread(topic, alsoid, selected); err != nil {
			utils.Danger(err, "Cannot create thread")
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
	// and user template define content here
	thread, err := data.AccountThreads(URLIDConv)
	if err != nil {
		fmt.Println("Error on routes ReadAccount")
		http.Redirect(writer, request, "/", 302)
		return
	}
	if thread == nil {
		http.Redirect(writer, request, "/", 302)
		return
	}

	posts, _ := data.GetUserPosts(URLIDConv)
	// likedposts, _ := data.GetUserLikedPosts(URLIDConv)
	thread[0].Cards = posts
	// thread[0].LikedPosts = likedposts
	filesFrom := []string{"layout", "private.navbar", "account", "accountbypost"}
	var files []string
	for _, file := range filesFrom {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(writer, "layout", &thread)

}

// POST /likes
func PostLike(writer http.ResponseWriter, request *http.Request) {
	// fmt.Println("---------------------")
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
		fmt.Println(body)
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
