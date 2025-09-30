package data

// session DatabaseManager instance for session operations
var sessionDM *DatabaseManager

// InitSessionDM initializes the DatabaseManager for session operations
func InitSessionDM(dm *DatabaseManager) {
	sessionDM = dm
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
