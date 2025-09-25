package routes

import (
	"fmt"
	"net/http"
	"time"

	"forum/utils"

	"forum/data"
)

type LoginSkin struct {
	Submit string
	Signup string

	ErrorStr string
}

// GET /login
// show the login page
var ErrorStrC1 string

func Login(writer http.ResponseWriter, request *http.Request) {
	// t := utils.ParseTemplateFiles("login.layout", "public.navbar", "login")
	var LS LoginSkin
	if ErrorStrC1 != "" {
		LS = LoginSkin{
			"Submit",
			"Signup",
			ErrorStrC1,
		}
	} else {
		LS = LoginSkin{
			"Submit",
			"Signup",
			"",
		}
	}
	utils.GenerateHTML(writer, &LS, "login.layout", "public.navbar", "login")
	ErrorStrC1 = ""

	// t.ExecuteTemplate(writer, LoginSkin)
}

// GET /signup
// show the signup page
func Signup(writer http.ResponseWriter, request *http.Request) {
	var LS LoginSkin
	if ErrorStrC1 != "" {
		LS = LoginSkin{
			"Submit",
			"Signup",
			ErrorStrC1,
		}
	} else {
		LS = LoginSkin{
			"Submit",
			"Signup",
			"",
		}
	}
	ErrorStrC1 = ""
	utils.GenerateHTML(writer, &LS, "login.layout", "public.navbar", "signup")
}

// POST /signup_account
// create the user account
func SignupAccount(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm() // err
	checkExists := data.IfUserExist(request.PostFormValue("email"), request.PostFormValue("name"))
	if checkExists {
		// fmt.Println("lol exists")
		ErrorStrC1 = "This name/email already exists\nTry to signup again using different username/email"
		Signup(writer, request)
		return
	}

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
	switch request.Method {
	case "GET":

	case "POST":
		request.ParseForm()
		email := request.PostFormValue("email")
		password := request.PostFormValue("password")

		fmt.Printf("Login attempt for email: %s\n", email)

		user, err := data.UserByEmail(email)
		if err != nil {
			fmt.Printf("User not found for email: %s, error: %v\n", email, err)
			ErrorStrC1 = "You might entered wrong email/password \n Try again"
			Login(writer, request)
			return
		}

		fmt.Printf("Found user: %s (ID: %d)\n", user.Name, user.Id)

		if user.Password == data.Encrypt(password) {
			fmt.Println("Password correct, checking existing session...")

			_, err = data.SessionCheck(writer, request)
			if err == nil {
				fmt.Println("Existing valid session found, redirecting to home")
				http.Redirect(writer, request, "/", 302)
				return
			}

			fmt.Println("Creating new session...")
			// Create session and get the session object
			session, err := user.CreateSession()
			if err != nil {
				fmt.Printf("Failed to create session: %v\n", err)
				utils.Danger(err, "Cannot create session")
				return
			}

			fmt.Printf("Session created successfully. CookieString: %s\n", session.CookieString)

			// Use the cookie string from the session
			cookie := http.Cookie{
				Name:     "_cookie",
				Value:    session.CookieString,
				HttpOnly: false, // Allow JavaScript access for debugging
				Expires:  time.Now().Add(24 * time.Hour),
				Path:     "/",
				SameSite: http.SameSiteLaxMode, // Allow same-site requests
				Secure:   false,                // Allow over HTTP for local development
			}
			fmt.Printf("Setting cookie: %s = %s\n", cookie.Name, cookie.Value) // Debug
			http.SetCookie(writer, &cookie)
			http.Redirect(writer, request, "/", 302)
		} else {
			fmt.Printf("Password incorrect for user: %s\n", user.Name)
			ErrorStrC1 = "You might entered wrong email/password \n Try again"
			Login(writer, request)
		}
	}
}

// GET /logout
// logs the user out
func Logout(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("_cookie")

	if err != http.ErrNoCookie && cookie != nil {
		// Find session by cookie value
		session, err := data.GetSessionByCookie(cookie.Value)
		if err == nil {
			// Delete the session
			session.DeleteByUUID()
		}

		// Invalidate the cookie
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0)
		http.SetCookie(writer, cookie)
	} else {
		utils.Warn(err, "Failed to get cookie")
	}
	http.Redirect(writer, request, "/", 302)
}
