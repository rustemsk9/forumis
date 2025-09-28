package data

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Global DatabaseManager instance for session operations
var sessionDM *DatabaseManager

// InitSessionDM initializes the global DatabaseManager for session operations
func InitSessionDM(dm *DatabaseManager) {
	sessionDM = dm
}

// check if session is valid in the database
func (session *Session) Valid() (valid bool, err error) {
	// Calculate current time as hour*100 + minute (e.g., 14:30 = 1430)
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	err = sessionDM.db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE uuid=?", session.Uuid).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	if err != nil {
		valid = false
		return
	}

	// fmt.Println(session.Uuid) // Removed debug print
	if session.Id != 0 {
		valid = true
		// Update active_last with current time
		_, err = sessionDM.db.Exec("UPDATE sessions SET active_last = ? WHERE uuid = ?", currentTime, session.Uuid)
		if err != nil {
			// Don't fail validation if update fails, just log it
			fmt.Printf("Warning: Could not update active_last for session %s: %v\n", session.Uuid, err)
		}
	}
	return
}

// delete session from database
func (session *Session) DeleteByUUID() (err error) {
	return sessionDM.DeleteSessionByUUID(session.Uuid)
}

// get the user from the session
func (session *Session) User() (user User, err error) {
	return sessionDM.GetSessionUser(session.UserId)
}

// delete all sessions from database
func SessionDeleteAll() (err error) {
	return sessionDM.DeleteAllSessions()
}

// checks if the user is logged in and has a session, if not err is not nil
func SessionCheck(writer http.ResponseWriter, request *http.Request) (sess Session, err error) {
	cookie, err := request.Cookie("_cookie")
	if err == nil {
		cooPart := strings.Split(cookie.Value, "&")
		if len(cooPart) >= 2 {
			sess = Session{Uuid: cooPart[1]} // Use the UUID part, not the user ID
			if ok, _ := sess.Valid(); !ok {
				err = errors.New("invalid session")
			}
		} else {
			err = errors.New("invalid cookie format")
		}
	}
	return
}

// Update session with cookie string
func (session *Session) UpdateCookieString(cookieValue string) error {
	return sessionDM.UpdateSessionCookieString(session.Uuid, cookieValue)
}

// Get session by cookie string
func GetSessionByCookie(cookieValue string) (sess Session, err error) {
	return sessionDM.GetSessionByCookie(cookieValue)
}

// CheckOnlineUsers returns a list of users who have been active recently
// considerOnline: time difference in minutes to consider a user online (e.g., 5 for 5 minutes)
func CheckOnlineUsers(considerOnline int) ([]User, error) {
	return sessionDM.CheckOnlineUsers(considerOnline)
}
