package routes

import (
	"fmt"
	"net/http"

	"forum/data"
	"forum/utils"
)

// Debug route to test cookie functionality
func DebugPage(writer http.ResponseWriter, request *http.Request) {
	utils.GenerateHTML(writer, nil, "layout", "private.navbar", "debug")
}

// Debug route to test cookie values
func DebugCookieTest(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")

	// Test GetCookieValue function
	userID := data.GetCookieValue(request)

	fmt.Fprintf(writer, "<h3>Cookie Debug Information</h3>")
	fmt.Fprintf(writer, "<p><strong>GetCookieValue() result:</strong> %d</p>", userID)

	fmt.Fprintf(writer, "<p><strong>All Cookies Received:</strong></p><ul>")
	for _, cookie := range request.Cookies() {
		fmt.Fprintf(writer, "<li>%s = %s</li>", cookie.Name, cookie.Value)
	}
	fmt.Fprintf(writer, "</ul>")

	// Try to get _cookie directly
	cookie, err := request.Cookie("_cookie")
	if err != nil {
		fmt.Fprintf(writer, "<p><strong>_cookie error:</strong> %v</p>", err)
	} else {
		fmt.Fprintf(writer, "<p><strong>_cookie value:</strong> %s</p>", cookie.Value)

		// Test session lookup
		session, err := data.GetSessionByCookie(cookie.Value)
		if err != nil {
			fmt.Fprintf(writer, "<p><strong>Session lookup error:</strong> %v</p>", err)
		} else {
			fmt.Fprintf(writer, "<p><strong>Session found:</strong> UserID=%d, Email=%s</p>", session.UserId, session.Email)
		}
	}
}
