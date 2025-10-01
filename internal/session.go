package internal

import (
	"forum/internal/data"
	"forum/models"
)

// session DatabaseManager instance for session operations
var sessionDM *data.DatabaseManager

// InitSessionDM initializes the DatabaseManager for session operations
func InitSessionDM(dm *data.DatabaseManager) {
	sessionDM = dm
}

// delete session from database
func DeleteByUUID(Uuid string) (err error) {
	return sessionDM.DeleteSessionByUUID(Uuid)
}

// delete all sessions from database
func SessionDeleteAll() (err error) {
	return sessionDM.DeleteAllSessions()
}

// Update session with cookie string
func UpdateCookieString(Uuid string, cookieValue string) error {
	return sessionDM.UpdateSessionCookieString(Uuid, cookieValue)
}

// Get session by cookie string
func GetSessionByCookie(cookieValue string) (sess models.Session, err error) {
	return sessionDM.GetSessionByCookie(cookieValue)
}

// CheckOnlineUsers returns a list of users who have been active recently
// considerOnline: time difference in minutes to consider a user online (e.g., 5 for 5 minutes)
func CheckOnlineUsers(considerOnline int) ([]models.User, error) {
	return sessionDM.CheckOnlineUsers(considerOnline)
}

func SessionByUUID(uuid string) bool {
	// This function seems to check if a session exists with the given UUID
	// Using the DatabaseManager to check session validity
	_, valid, _ := sessionDM.ValidateSession(uuid)
	return valid
}
