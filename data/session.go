package data

import (
	"errors"
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
}

// check if session is valid in the database
func (session *Session) Valid() (valid bool, err error) {
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string FROM sessions WHERE uuid=?", session.Uuid).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString)
	if err != nil {
		valid = false
		return
	}
	// fmt.Println(session.Uuid) // Removed debug print
	if session.Id != 0 {
		valid = true
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
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string FROM sessions WHERE cookie_string=?", cookieValue).
		Scan(&sess.Id, &sess.Uuid, &sess.Email, &sess.UserId, &sess.CreatedAt, &sess.CookieString)
	return
}
