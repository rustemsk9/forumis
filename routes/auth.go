package routes

import (
	"casual-talk/data"
	"casual-talk/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// GET /login
// show the login page
func Login(writer http.ResponseWriter, request *http.Request) {
	t := utils.ParseTemplateFiles("login.layout", "public.navbar", "login")
	t.Execute(writer, nil)
}

// GET /signup
// show the signup page
func Signup(writer http.ResponseWriter, request *http.Request) {
	utils.GenerateHTML(writer, nil, "login.layout", "public.navbar", "signup")
}

// POST /signup_account
// create the user account
func SignupAccount(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		utils.Danger(err, "Cannot parse form")
	}
	user := data.User{
		Name:     request.PostFormValue("name"),
		Email:    request.PostFormValue("email"),
		Password: request.PostFormValue("password"),
	}
	if err := user.Create(); err != nil {
		utils.Danger(err, "Cannot create user")
	}
	http.Redirect(writer, request, "/login", 302)
}

// POST /authenticate
// authenticate the user given the email and password
func Authenticate(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	user, err := data.UserByEmail(request.PostFormValue("email"))
	if err != nil {
		// http.Redirect()
		// utils.Danger(err, "Cannot find user")
		// http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.Password == data.Encrypt(request.PostFormValue("password")) {
		_, err := user.CreateSession()
		if err != nil {
			utils.Danger(err, "Cannot find user")
		}
		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    fmt.Sprintf("%d", user.Id),
			HttpOnly: true,
		}
		// main.Authenticate = true
		http.SetCookie(writer, &cookie)
		http.Redirect(writer, request, "/", 302)
	} else {
		http.Redirect(writer, request, "/login", 302)
	}
}

// GET /logout
// logs the user out
func Logout(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("_cookie")
	ourID, _ := strconv.Atoi(cookie.Value)
	if err != http.ErrNoCookie {
		session := data.Session{UserId: ourID}
		session.DeleteByUUID()
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0)
		http.SetCookie(writer, cookie)
		fmt.Println("LOL1")
	} else {
		utils.Warn(err, "Failed to get cookie")
		fmt.Println("LOL2")
	}
	http.Redirect(writer, request, "/", 302)
}
