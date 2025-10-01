package data

import (
	"fmt"
	"forum/models"
	"forum/utils"
	"time"
)

// User operations
func (dm *DatabaseManager) CreateUser(user *models.User) error {
	user.Uuid = utils.CreateUUID()
	user.CreatedAt = time.Now()

	stmt, err := dm.db.Prepare(
		"INSERT INTO users(uuid, name, email, password, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.Uuid, user.Name, user.Email, utils.Encrypt(user.Password), user.CreatedAt)
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

func (dm *DatabaseManager) DeleteUserById(user *models.User) error {
	if user.Id == 0 {
		return fmt.Errorf("invalid user ID")
	}

	_, err := dm.db.Exec("DELETE FROM users WHERE id=?", user.Id)
	return err
}

func (dm *DatabaseManager) CheckUserExists(email, name string) (bool, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR name = ?", email, name).Scan(&count)
	return count > 0, err
}

func (dm *DatabaseManager) GetUserByEmailDetailed(email string) (user models.User, err error) {
	err = dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) GetUserByID(id int) (user models.User, err error) {

	err = dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE id=?", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

func (dm *DatabaseManager) GetUserByEmail(email string) (user models.User, err error) {
	err = dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

// Update user preferred categories
func (dm *DatabaseManager) UpdateUserPreferences(userID int, category1, category2 string) error {
	_, err := dm.db.Exec("UPDATE users SET prefered_category1=?, prefered_category2=? WHERE id=?",
		category1, category2, userID)
	return err
}

func (dm *DatabaseManager) CheckOnlineUsers(considerOnline int) ([]models.User, error) {
	var users []models.User
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
		var user models.User
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

// update user information in the database
func (dm *DatabaseManager) Update(userName string, userId int) (err error) {
	_, err = dm.db.Exec("UPDATE users SET name=? WHERE id=?", userName, userId)
	return err
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

func (dm *DatabaseManager) GetUserByUUID(uuid string) (models.User, error) {
	var user models.User
	err := dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE uuid=?", uuid).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (dm *DatabaseManager) DeleteUserByID(userID int) error {
	result, err := dm.db.Exec("DELETE FROM users WHERE id=?", userID)
	if err != nil {
		return err
	}
	// LastInsertId is not meaningful for DELETE operations; it only applies to INSERT.
	// For DELETE, you can use RowsAffected to check how many rows were deleted.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", userID)
	}
	return nil
}

func (dm *DatabaseManager) DeleteUserByName(name string) error {
	result, err := dm.DoExec("DELETE FROM users WHERE name=?", name)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with name %s", name)
	}
	return nil
}

func (dm *DatabaseManager) GetUserByName(name string) (user models.User, err error) {
	err = dm.db.QueryRow("SELECT id, uuid, name, email, password, created_at, prefered_category1, prefered_category2 FROM users WHERE name=?", name).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.PreferedCategory1, &user.PreferedCategory2)
	return user, err
}

func (dm *DatabaseManager) DeleteAllUsers() error {
	_, err := dm.db.Exec("DELETE FROM users")
	return err
}

func (dm *DatabaseManager) GetAllUsers() ([]models.User, error) {
	var users []models.User
	rows, err := dm.db.Query("SELECT id, uuid, name, email, password, created_at FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (dm *DatabaseManager) GetUserLikedPosts(userID int) ([]models.Likes, error) {
	var likes []models.Likes
	rows, err := dm.db.Query("SELECT COALESCE(type, 'like') as type, user_id, post_id FROM likedposts WHERE user_id=?", userID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like models.Likes
		err = rows.Scan(&like.Type, &like.UserId, &like.PostId)
		if err != nil {
			fmt.Println("Error scanning like:", err)
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (dm *DatabaseManager) GetUserLikedThreads(userID int) ([]models.ThreadLikes, error) {
	var likes []models.ThreadLikes
	rows, err := dm.db.Query("SELECT type, user_id, thread_id FROM threadlikes WHERE user_id=?", userID)
	if err != nil {
		return likes, err
	}
	defer rows.Close()

	for rows.Next() {
		var like models.ThreadLikes
		err = rows.Scan(&like.Type, &like.UserId, &like.ThreadId)
		if err != nil {
			continue
		}
		likes = append(likes, like)
	}
	return likes, nil
}
