package routes

import (
	"casual-talk/data"
	"casual-talk/utils"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GET /login
// show the login page
func Login(writer http.ResponseWriter, request *http.Request) {
	t := utils.ParseTemplateFiles("login.layout", "public.navbar", "login")

	// switch str {
	// case "":
	str := "Welcome to Forum"
	// case "errormsg":
	// 	str = "You entered either wrong login or wrong password"
	// }

	t.Execute(writer, str)
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
		http.Redirect(writer, request, "/login", 302)
		// http.Redirect()
		// utils.Danger(err, "Cannot find user")
		// http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.Password == data.Encrypt(request.PostFormValue("password")) {
		_, err := user.CreateSession()
		sess, err := user.Session()
		if err != nil {
			utils.Danger(err, "Cannot find user")
		}
		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    fmt.Sprintf("%d&%v", user.Id, sess.Uuid),
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
	cooPart := strings.Split(cookie.Value, "&")
	fmt.Println(cooPart[1])
	if err != http.ErrNoCookie {
		session := data.Session{Uuid: cooPart[1]}
		session.DeleteByUUID()
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0)
		http.SetCookie(writer, cookie)
	} else {
		utils.Warn(err, "Failed to get cookie")
	}
	http.Redirect(writer, request, "/", 302)
}
