package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var logger *log.Logger

func init() {
	file, err := os.OpenFile("casual-talk.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func CreateUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		Danger("Cannot generate UUID", err)
	}

	// 0x40 is reserved variant from RFC 4122

	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}

// hash plaintext with SHA-1, changed to bcrypt as in forum instructions.
func Encrypt(plaintext string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func PasswordMeetsCriteria(writer http.ResponseWriter, request *http.Request, password string) bool {
	// Check password strength using regex
	var (
		upperCase = regexp.MustCompile(`[A-Z]`)
		lowerCase = regexp.MustCompile(`[a-z]`)
		number    = regexp.MustCompile(`[0-9]`)
		symbol    = regexp.MustCompile(`[\W_]`)
	)

	if !upperCase.MatchString(password) ||
		!lowerCase.MatchString(password) ||
		!number.MatchString(password) ||
		!symbol.MatchString(password) {
		return false
	}
	return true
}

// parse HTML templates
// pass in a list of file names, and get a template
func ParseTemplateFiles(filenames ...string) (t *template.Template) {
	var files []string
	t = template.New("layout.html")
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}
	t = template.Must(t.ParseFiles(files...))
	return
}

func GenerateHTML(writer http.ResponseWriter, data interface{}, fn ...string) {

	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) // Marks the string as safe HTML (no escaping)
		},
		// Add more functions here if needed, e.g., "upper": strings.ToUpper
	}
	// Create a new template with functions
	tmpl := template.New("").Funcs(funcMap)

	var files []string
	for _, file := range fn {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}
	// Always include cookie-consent and lidi templates
	files = append(files, "templates/cookie-consent.html")
	// files = append(files, "templates/lidi.html")
	template := template.Must(tmpl.ParseFiles(files...))
	err := template.ExecuteTemplate(writer, "layout", data)
	if err != nil {
		Danger("Failed to execute template:", err)
	}
}

// convenience function to redirect to the error message page
func ErrorMessage(writer http.ResponseWriter, request *http.Request, msg string) {
	url := []string{"/err?msg=", msg}
	http.Redirect(writer, request, strings.Join(url, ""), 302)
}

// Handle 400 Bad Request errors with custom message
func BadRequest(writer http.ResponseWriter, request *http.Request, message string) {
	writer.WriteHeader(http.StatusBadRequest)
	if isAPIRequest(request) {
		writeJSONError(writer, http.StatusBadRequest, message)
		return
	}
	GenerateHTML(writer, map[string]interface{}{
		"Title":   "Bad Request",
		"Message": message,
		"Code":    400,
	}, "layout", "public.navbar", "error")
}

// Handle 404 Not Found errors
func NotFound(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	if isAPIRequest(request) {
		writeJSONError(writer, http.StatusNotFound, "Resource not found")
		return
	}
	GenerateHTML(writer, map[string]interface{}{
		"Title":   "Page Not Found",
		"Message": "The page you're looking for doesn't exist.",
		"Code":    404,
	}, "layout", "public.navbar", "error")
}

// Handle 500 Internal Server Error
func InternalServerError(writer http.ResponseWriter, request *http.Request, err error) {
	Danger("Internal server error:", err)
	writer.WriteHeader(http.StatusInternalServerError)
	if isAPIRequest(request) {
		writeJSONError(writer, http.StatusInternalServerError, "Internal server error")
		return
	}
	GenerateHTML(writer, map[string]interface{}{
		"Title":   "Server Error",
		"Message": "Something went wrong on our end. Please try again later.",
		"Code":    500,
	}, "layout", "public.navbar", "error")
}

// Handle 405 Method Not Allowed errors
func MethodNotAllowed(writer http.ResponseWriter, request *http.Request, message string) {
	writer.WriteHeader(http.StatusMethodNotAllowed)
	if isAPIRequest(request) {
		writeJSONError(writer, http.StatusMethodNotAllowed, message)
		return
	}
	GenerateHTML(writer, map[string]interface{}{
		"Title":   "Method Not Allowed",
		"Message": message,
		"Code":    405,
	}, "layout", "public.navbar", "error")
}

// Handle 401 Unauthorized errors
func Unauthorized(writer http.ResponseWriter, request *http.Request, message string) {
	writer.WriteHeader(http.StatusUnauthorized)
	if isAPIRequest(request) {
		writeJSONError(writer, http.StatusUnauthorized, "Authentication required")
		return
	}
	GenerateHTML(writer, map[string]interface{}{
		"Title":   "Unauthorized",
		"Message": message,
		"Code":    401,
	}, "layout", "public.navbar", "error")
	http.Redirect(writer, request, "/login", http.StatusSeeOther)
}

// Handle 403 Forbidden errors
func Forbidden(writer http.ResponseWriter, request *http.Request, message string) {
	writer.WriteHeader(http.StatusForbidden)
	if isAPIRequest(request) {
		writeJSONError(writer, http.StatusForbidden, message)
		return
	}
	GenerateHTML(writer, map[string]interface{}{
		"Title":   "Access Forbidden",
		"Message": message,
		"Code":    403,
	}, "layout", "public.navbar", "error")
}

// Check if the request is an API request
func isAPIRequest(request *http.Request) bool {
	return strings.HasPrefix(request.URL.Path, "/api/")
}

// Write JSON error response for API requests
func writeJSONError(writer http.ResponseWriter, code int, message string) {
	writer.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	json.NewEncoder(writer).Encode(response)
}

func FileExists(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}
	return true
}

// log
func Info(args ...interface{}) {
	logger.SetPrefix("[INFO] ")
	logger.Println(args...)
}

func Danger(args ...interface{}) {
	logger.SetPrefix("[ERROR] ")
	logger.Println(args...)
}

func Warn(args ...interface{}) {
	logger.SetPrefix("[WARN] ")
	logger.Println(args...)
}
