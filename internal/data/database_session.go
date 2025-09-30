package data

import (
	"fmt"
	"forum/models"
	"forum/utils"
	"log"
	"time"
)

// Session operations
func (dm *DatabaseManager) CreateSession(user *models.User) (models.Session, error) {
	// Delete existing sessions for this user
	dm.db.Exec("DELETE from sessions where user_id=?", user.Id)

	// Create the session UUID
	sessionUUID := utils.CreateUUID()

	// Create cookie string value in format "userId&sessionUUID"
	cookieString := fmt.Sprintf("%d&%s", user.Id, sessionUUID)

	// Calculate current time as hour*100 + minute
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	// Insert session
	stmt, err := dm.db.Prepare(
		"INSERT INTO sessions(uuid, email, user_id, created_at, cookie_string, active_last) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return models.Session{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sessionUUID, user.Email, user.Id, time.Now(), cookieString, currentTime)
	if err != nil {
		return models.Session{}, err
	}

	return models.Session{
		Uuid:         sessionUUID,
		Email:        user.Email,
		UserId:       user.Id,
		CreatedAt:    time.Now(),
		CookieString: cookieString,
		ActiveLast:   currentTime,
	}, nil
}

func (dm *DatabaseManager) ValidateSession(sessionUUID string) (models.Session, bool, error) {
	var session models.Session

	// Calculate current time as hour*100 + minute
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	err := dm.db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE uuid=?", sessionUUID).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	if err != nil {
		return session, false, err
	}

	if session.Id != 0 {
		// Only update active_last if it's been more than 5 minutes since last update
		// This reduces database writes and prevents locking issues
		timeDiff := currentTime - session.ActiveLast
		if timeDiff < 0 {
			// Handle day rollover (current time is next day)
			timeDiff = (2400 + currentTime) - session.ActiveLast
		}

		// Update if more than 5 minutes have passed (500 = 5 hours in our time format, 5 = 5 minutes)
		if timeDiff >= 5 {
			_, err = dm.db.Exec("UPDATE sessions SET active_last = ? WHERE uuid = ?", currentTime, sessionUUID)
			if err != nil {
				// Don't fail validation if we can't update timestamp, just log it
				log.Printf("Warning: Could not update active_last for session %s: %v\n", sessionUUID, err)
			} else {
				session.ActiveLast = currentTime
			}
		}
		return session, true, nil
	}

	return session, false, nil
}

func (dm *DatabaseManager) GetSessionByCookie(cookieValue string) (models.Session, error) {
	var session models.Session
	err := dm.db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE cookie_string=?", cookieValue).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	return session, err
}

func (dm *DatabaseManager) DeleteSession(sessionUUID string) error {
	_, err := dm.db.Exec("DELETE FROM sessions WHERE uuid=?", sessionUUID)
	return err
}

// Session management methods needed by session.go
func (dm *DatabaseManager) UpdateSessionCookieString(uuid, cookieValue string) error {
	_, err := dm.db.Exec("UPDATE sessions SET cookie_string=? WHERE uuid=?", cookieValue, uuid)
	return err
}

func (dm *DatabaseManager) DeleteSessionByUUID(uuid string) error {
	_, err := dm.db.Exec("DELETE FROM sessions WHERE uuid=?", uuid)
	return err
}

func (dm *DatabaseManager) DeleteAllSessions() error {
	_, err := dm.db.Exec("DELETE FROM sessions")
	return err
}

func (dm *DatabaseManager) GetSessionUser(userID int) (models.User, error) {
	var user models.User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", userID).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) GetUserDislikedThreads(userID int) ([]models.ThreadDislikes, error) {
	var dislikes []models.ThreadDislikes
	rows, err := dm.db.Query("SELECT type, user_id, thread_id FROM threaddislikes WHERE user_id=?", userID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike models.ThreadDislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.ThreadId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}

func (dm *DatabaseManager) GetUserDislikedPosts(userID int) ([]models.Dislikes, error) {
	var dislikes []models.Dislikes
	rows, err := dm.db.Query("SELECT type, user_id, post_id FROM dislikes WHERE user_id=?", userID)
	if err != nil {
		return dislikes, err
	}
	defer rows.Close()

	for rows.Next() {
		var dislike models.Dislikes
		err = rows.Scan(&dislike.Type, &dislike.UserId, &dislike.PostId)
		if err != nil {
			continue
		}
		dislikes = append(dislikes, dislike)
	}
	return dislikes, nil
}
