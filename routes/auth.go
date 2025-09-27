package routes

import (
	"fmt"
	"net/http"
	"time"

	"forum/utils"

	"forum/data"
)

// GET /login
// show the login page
var Error string

func Login(writer http.ResponseWriter, request *http.Request, LS data.LoginSkin) {
	if request.URL.Path == "/login/success" {
		LS = data.LoginSkin{
			Submit: "Submit",
			Signup: "Signup",
			Name:   LS.Name,
			Email:  LS.Email,
			Error:  "Signup successful! Please log in.",
		}
		utils.GenerateHTML(writer, &LS, "login.layout", "public.navbar", "login")
		Error = ""
		return
	} else {
	if Error != "" {
		LS = data.LoginSkin{
			Submit: "Again",
			Signup: "Signup",
			Name:   LS.Name,
			Email:  LS.Email,
			Error:  Error,
		}
	} else {
		LS = data.LoginSkin{
			Submit: "Submit",
			Signup: "Signup",
			Name:   LS.Name,
			Email:  LS.Email,
			Error:  "",
		}
	}
	
	utils.GenerateHTML(writer, &LS, "login.layout", "public.navbar", "login")
	Error = ""
	}
}

// GET /signup
// show the signup page
func Signup(writer http.ResponseWriter, request *http.Request, LS data.LoginSkin) {
	// var LS LoginSkin
	if Error != "" {
		LS = data.LoginSkin{
			Submit: "Submit",
			Signup: "Again",
			Name:   LS.Name,
			Email:  LS.Email,
			Error:  Error,
		}
	} else {
		LS = data.LoginSkin{
			Submit: "Submit",
			Signup: "Signup",
			Name:   LS.Name,
			Email:  LS.Email,
			Error:  "",
		}
	}
	Error = ""
	utils.GenerateHTML(writer, &LS, "login.layout", "public.navbar", "signup")
}

// POST /signup_account
// create the user account
func SignupAccount(writer http.ResponseWriter, request *http.Request) {
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.ErrorMessage(writer, request, "Database not available")
		return
	}

	err := request.ParseForm()
	if err != nil {
		utils.Danger(err, "Cannot parse form")
	}
	
	// Check if user already exists
	checkExists := data.IfUserExist(request.PostFormValue("email"), request.PostFormValue("name"))
	if checkExists {
		Error = "This name/email already exists\nTry to signup again using different username/email"
		user := data.LoginSkin{
			Submit: "Try Again",
			Signup: "Signup",
			Name:   request.PostFormValue("name"),
			Email:  request.PostFormValue("email"),
			Error:  Error,
		}
		Signup(writer, request, user)
		return
	}

	usertoSign := data.User{
		Name:     request.PostFormValue("name"),
		Email:    request.PostFormValue("email"),
		Password: request.PostFormValue("password"),
	}
	
	// Use database manager to create user
	if err := dbManager.CreateUser(&usertoSign); err != nil {
		utils.Danger(err, "Cannot create user")
		Error = "Internal server error\nPlease try again later"
		user := data.LoginSkin{
			Submit: "Try Again",
			Signup: "Signup",
			Name:   request.PostFormValue("name"),
			Email:  request.PostFormValue("email"),
			Error:  Error,
		}
		Signup(writer, request, user)
		return
	}
	
	http.Redirect(writer, request, "/login/success", 302)
}

// POST /authenticate
// authenticate the user given the email and password
func Authenticate(writer http.ResponseWriter, request *http.Request) {
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.ErrorMessage(writer, request, "Database not available")
		return
	}

	switch request.Method {
	case "GET":
		// Handle GET if needed
	case "POST":
		err := request.ParseForm()
		if err != nil {
			utils.Danger(err, "Cannot parse form")
		}
		email := request.PostFormValue("email")
		password := request.PostFormValue("password")

		fmt.Printf("Login attempt for email: %s\n", email)

		// Use database manager to get user
		user, err := dbManager.GetUserByEmail(email)
		if err != nil {
			fmt.Printf("User not found for email: %s, error: %v\n", email, err)
			Error = "You might entered wrong email/password \n Try again"
			Login(writer, request, data.LoginSkin{})
			return
		}

		fmt.Printf("Found user: %s (ID: %d)\n", user.Name, user.Id)

		if data.CheckPassword(user.Password, password) {
			fmt.Println("Password correct, checking existing session...")

			// Check if user already has a session from middleware
			if IsAuthenticated(request) {
				fmt.Println("Existing valid session found, redirecting to home")
				http.Redirect(writer, request, "/", 302)
				return
			}

			fmt.Println("Creating new session...")
			// Create session using database manager
			session, err := dbManager.CreateSession(&user)
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
			Error = "You might entered wrong email/password \n Try again"
			Login(writer, request, data.LoginSkin{})
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
