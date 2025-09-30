package data

import (
	"fmt"
	"time"
)

// User operations
func (dm *DatabaseManager) CreateUser(user *User) error {
	user.Uuid = createUUID()
	user.CreatedAt = time.Now()

	stmt, err := dm.db.Prepare(
		"INSERT INTO users(uuid, name, email, password, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.Uuid, user.Name, user.Email, Encrypt(user.Password), user.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.Id = int(id)
	return nil
}

func (dm *DatabaseManager) CheckUserExists(email, name string) (bool, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR name = ?", email, name).Scan(&count)
	return count > 0, err
}

func (dm *DatabaseManager) GetUserByEmailDetailed(email string) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) GetUserByID(id int) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE id=?", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

func (dm *DatabaseManager) GetUserByEmail(email string) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

// Update user preferred categories
func (dm *DatabaseManager) UpdateUserPreferences(userID int, category1, category2 string) error {
	_, err := dm.db.Exec("UPDATE users SET prefered_category1=?, prefered_category2=? WHERE id=?",
		category1, category2, userID)
	return err
}

func (dm *DatabaseManager) CheckOnlineUsers(considerOnline int) ([]User, error) {
	var users []User
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	query := `
		SELECT DISTINCT u.id, u.uuid, u.name, u.email, u.created_at, s.active_last
		FROM users u 
		INNER JOIN sessions s ON u.id = s.user_id 
		WHERE s.active_last > 0`

	rows, err := dm.db.Query(query)
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
		var timeDiff int
		if currentTime >= activeLast {
			hourDiff := (currentTime / 100) - (activeLast / 100)
			minuteDiff := (currentTime % 100) - (activeLast % 100)
			timeDiff = hourDiff*60 + minuteDiff
		} else {
			hourDiff := (24 + currentTime/100) - (activeLast / 100)
			minuteDiff := (currentTime % 100) - (activeLast % 100)
			timeDiff = hourDiff*60 + minuteDiff
		}

		if timeDiff <= considerOnline {
			users = append(users, user)
		}
	}

	return users, nil
}

// Alternative method names for compatibility with thread.go
func (dm *DatabaseManager) HasUserLikedThread(userID, threadID int) int {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threadlikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	if err != nil {
		return count
	}
	return 0
}

func (dm *DatabaseManager) HasUserDislikedThread(userID, threadID int) int {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM threaddislikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	if err != nil {
		return count
	}
	return 0
}

func (dm *DatabaseManager) GetUserByUUID(uuid string) (User, error) {
	var user User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE uuid=?", uuid).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) DeleteUserByID(userID int) error {
	_, err := dm.db.Exec("DELETE FROM users WHERE id=?", userID)
	return err
}

func (dm *DatabaseManager) DeleteAllUsers() error {
	_, err := dm.db.Exec("DELETE FROM users")
	return err
}

func (dm *DatabaseManager) GetAllUsers() ([]User, error) {
	var users []User
	rows, err := dm.db.Query("SELECT id, uuid, name, email, password, created_at FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (dm *DatabaseManager) GetUserLikedPosts(userID int) ([]Likes, error) {
	var likes []Likes
	rows, err := dm.db.Query("SELECT COALESCE(type, 'like') as type, user_id, post_id FROM likedposts WHERE user_id=?", userID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like Likes
		err = rows.Scan(&like.Type, &like.UserId, &like.PostId)
		if err != nil {
			fmt.Println("Error scanning like:", err)
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetUserLikedThreads(userID int) ([]ThreadLikes, error) {
	var likes []ThreadLikes
	rows, err := dm.db.Query("SELECT type, user_id, thread_id FROM threadlikes WHERE user_id=?", userID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like ThreadLikes
		err = rows.Scan(&like.Type, &like.UserId, &like.ThreadId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}
