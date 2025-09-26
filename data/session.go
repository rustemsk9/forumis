package data

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Session struct {
	Id           int
	Uuid         string
	Email        string
	UserId       int
	CreatedAt    time.Time
	CookieString string
	ActiveLast   int
}

// check if session is valid in the database
func (session *Session) Valid() (valid bool, err error) {
	// Calculate current time as hour*100 + minute (e.g., 14:30 = 1430)
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()
	
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE uuid=?", session.Uuid).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	if err != nil {
		valid = false
		return
	}
	
	// fmt.Println(session.Uuid) // Removed debug print
	if session.Id != 0 {
		valid = true
		// Update active_last with current time
		_, err = Db.Exec("UPDATE sessions SET active_last = ? WHERE uuid = ?", currentTime, session.Uuid)
		if err != nil {
			// Don't fail validation if update fails, just log it
			fmt.Printf("Warning: Could not update active_last for session %s: %v\n", session.Uuid, err)
		}
	}
	return
}

// delete session from database
func (session *Session) DeleteByUUID() (err error) {
	stmt, err := Db.Prepare("DELETE FROM sessions WHERE uuid=?")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(session.Uuid)
	return
}

// get the user from the session
func (session *Session) User() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", session.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return
}

// delete all sessions from database
func SessionDeleteAll() (err error) {
	_, err = Db.Exec("DELETE FROM sessions")
	return
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
	stmt, err := Db.Prepare("UPDATE sessions SET cookie_string=? WHERE uuid=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cookieValue, session.Uuid)
	return err
}

// Get session by cookie string
func GetSessionByCookie(cookieValue string) (sess Session, err error) {
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE cookie_string=?", cookieValue).
		Scan(&sess.Id, &sess.Uuid, &sess.Email, &sess.UserId, &sess.CreatedAt, &sess.CookieString, &sess.ActiveLast)
	return
}

// CheckOnlineUsers returns a list of users who have been active recently
// considerOnline: time difference in minutes to consider a user online (e.g., 5 for 5 minutes)
func CheckOnlineUsers(considerOnline int) ([]User, error) {
	var users []User
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()
	
	// Query to get active sessions with user information
	query := `
		SELECT DISTINCT u.id, u.uuid, u.name, u.email, u.created_at, s.active_last
		FROM users u 
		INNER JOIN sessions s ON u.id = s.user_id 
		WHERE s.active_last > 0`
	
	rows, err := Db.Query(query)
	if err != nil {
		return users, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var user User
		var activeLast int
		
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &activeLast)
		if err != nil {
			continue
		}
		
		// Calculate time difference in minutes
		// Handle hour rollover (e.g., from 23:59 to 00:05)
		var timeDiff int
		if currentTime >= activeLast {
			// Same day
			hourDiff := (currentTime/100) - (activeLast/100)
			minuteDiff := (currentTime%100) - (activeLast%100)
			timeDiff = hourDiff*60 + minuteDiff
		} else {
			// Hour rollover (next day)
			hourDiff := (24 + currentTime/100) - (activeLast/100)
			minuteDiff := (currentTime%100) - (activeLast%100)
			timeDiff = hourDiff*60 + minuteDiff
		}
		
		// Only include users active within the specified time frame
		if timeDiff <= considerOnline {
			users = append(users, user)
		}
	}
	
	return users, nil
}
