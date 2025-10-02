package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"forum/internal"
	"forum/models"
	"forum/utils"
)

// GET /login
// show the login page
var Error string

func Login(writer http.ResponseWriter, request *http.Request, LS models.LoginSkin) {
	if request.URL.Path == "/login/success" {
		LS = models.LoginSkin{
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
			LS = models.LoginSkin{
				Submit: "Again",
				Signup: "Signup",
				Name:   LS.Name,
				Email:  LS.Email,
				Error:  Error,
			}
		} else {
			LS = models.LoginSkin{
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
func Signup(writer http.ResponseWriter, request *http.Request, LS models.LoginSkin) {
	switch request.Method {
	case "GET":
		if Error != "" {
			LS = models.LoginSkin{
				Submit: "Submit",
				Signup: "Again",
				Name:   LS.Name,
				Email:  LS.Email,
				Error:  Error,
			}
		} else {
			LS = models.LoginSkin{
				Submit: "Submit",
				Signup: "Signup",
				Name:   LS.Name,
				Email:  LS.Email,
				Error:  "",
			}
		}
		Error = ""
		utils.GenerateHTML(writer, &LS, "public.navbar", "login.layout", "signup")
	case "POST":
		// Generate a UUID for the signup process (optional, for tracking or CSRF protection)
		// uuid := utils.GenerateUUID() // Implement this function if needed
		// You can pass uuid to SignupAccount or store it in context/session if required
		SignupAccount(writer, request)
		return
	default:
		utils.MethodNotAllowed(writer, request, "GET or POST method only")
	}
}

// POST /signup_account
// create the user account
func SignupAccount(writer http.ResponseWriter, request *http.Request) {

	if request.Method != "POST" {
		utils.MethodNotAllowed(writer, request, "POST method only")
		return
	}

	err := request.ParseForm()
	if err != nil {
		utils.BadRequest(writer, request, "Cannot parse form data")
		return
	}

	// Validate username format
	name := request.PostFormValue("name")
	email := request.PostFormValue("email")
	checkemail := strings.Split(email, ".")

	if name == "" || email == "" || request.PostFormValue("password") == "" {
		return
	}

	if len(checkemail) < 2 || len(checkemail) > 2 || strings.Contains(email, " ") {
		Error = "Wrong Email format"
		user := models.LoginSkin{
			Submit: "Try Again",
			Signup: "Signup",
			Name:   name,
			Email:  email,
			Error:  Error,
		}
		utils.GenerateHTML(writer, &user, "login.layout", "public.navbar", "signup")
		return
	}

	// Check if name contains special symbols
	for _, ch := range name {
		if !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') {
			Error = "Wrong UserName format"
			user := models.LoginSkin{
				Submit: "Try Again",
				Signup: "Signup",
				Name:   name,
				Email:  email,
				Error:  Error,
			}
			utils.GenerateHTML(writer, &user, "login.layout", "public.navbar", "signup")
			return
		}
	}

	if len(name) == 0 || (len(name) > 0 && name[0] == ' ') || len(name) > 20 || len(name) < 3 {
		user := models.LoginSkin{
			Submit: "Try Again",
			Signup: "Signup",
			Name:   name,
			Email:  email,
			Error:  "Wrong UserName format",
		}
		utils.GenerateHTML(writer, &user, "login.layout", "public.navbar", "signup")
		return
	}
	for _, ch := range name {
		if ch == ' ' || ch < 33 || ch > 121 {
			user := models.LoginSkin{
				Submit: "Try Again",
				Signup: "Signup",
				Name:   name,
				Email:  email,
				Error:  "Wrong UserName format",
			}
			utils.GenerateHTML(writer, &user, "login.layout", "public.navbar", "signup")
			return
		}
	}
	// Check if user already exists
	checkExists := internal.IfUserExist(request.PostFormValue("email"), request.PostFormValue("name"))
	if checkExists {
		Error = "This name/email already exists\nTry to signup again using different username/email"
		user := models.LoginSkin{
			Submit: "Try Again",
			Signup: "Signup",
			Name:   request.PostFormValue("name"),
			Email:  request.PostFormValue("email"),
			Error:  Error,
		}
		utils.GenerateHTML(writer, &user, "login.layout", "public.navbar", "signup")
		return
	}

	if ok := utils.PasswordMeetsCriteria(writer, request, request.PostFormValue("password")); !ok {
		Error = "Password must contain uppercase, lowercase, number, and symbol"
		user := models.LoginSkin{
			Submit: "Try Again",
			Signup: "Signup",
			Name:   request.PostFormValue("name"),
			Email:  request.PostFormValue("email"),
			Error:  Error,
		}
		utils.GenerateHTML(writer, &user, "login.layout", "public.navbar", "signup")
		return
	}

	usertoSign := models.User{
		Name:     request.PostFormValue("name"),
		Email:    request.PostFormValue("email"),
		Password: request.PostFormValue("password"),
	}

	// Use database manager to create user
	if err := internal.CreateUser(usertoSign); err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	http.Redirect(writer, request, "/login/success", 302)
}

// POST /authenticate
// authenticate the user given the email and password
func Authenticate(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		utils.MethodNotAllowed(writer, request, "POST method only")
		return
	}
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	err := request.ParseForm()
	if err != nil {
		utils.BadRequest(writer, request, "Cannot parse form data")
		return
	}
	email := request.PostFormValue("email")
	password := request.PostFormValue("password")

	if email == "" || password == "" {
		utils.BadRequest(writer, request, "Email and password are required")
		return
	}

	fmt.Printf("Login attempt for email: %s\n", email)

	// Use database manager to get user
	user, err := dbManager.GetUserByEmail(email)
	if err != nil {
		Error = "You might entered wrong email/password \n Try again"
		Login(writer, request, models.LoginSkin{})
		return
	}

	fmt.Printf("Found user: %s (ID: %d)\n", user.Name, user.Id)

	if utils.CheckPassword(user.Password, password) {
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
			utils.InternalServerError(writer, request, err)
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
		Login(writer, request, models.LoginSkin{})
	}
}

// GET /logout
// logs the user out
func Logout(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		utils.MethodNotAllowed(writer, request, "GET method only")
		return
	}
	cookie, err := request.Cookie("_cookie")

	if err != http.ErrNoCookie && cookie != nil {
		// Find session by cookie value
		session, err := internal.GetSessionByCookie(cookie.Value)
		if err == nil {
			// Delete the session
			internal.DeleteByUUID(session.Uuid)
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
