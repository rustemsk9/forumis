package data

import (
	"database/sql"
	"forum/models"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseManager struct {
	db *sql.DB
}

func (dm *DatabaseManager) DoExec(query string, args ...interface{}) (sql.Result, error) {
	return dm.db.Exec(query, args...)
}

func (dm *DatabaseManager) Ping() error {
	return dm.db.Ping()
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	// Add connection string parameters to handle concurrent access better
	connStr := dbPath + "?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_timeout=5000"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for better concurrent handling
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return &DatabaseManager{db: db}, nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	return dm.db.Close()
}

// GetDB returns the database connection (for migration purposes)
func (dm *DatabaseManager) GetDB() *sql.DB {
	return dm.db
}

// Statistics methods needed by statistics.go
func (dm *DatabaseManager) GetUserCount() (int, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (dm *DatabaseManager) GetMostActiveUsers(limit int) (users []models.User, err error) {
	var user models.User
	rows, err := dm.db.Query(`
		SELECT u.id, u.uuid, u.name, u.email, u.password, u.created_at, COUNT(t.id) AS thread_count
		FROM users u
		LEFT JOIN threads t ON u.id = t.user_id
		GROUP BY u.id
		ORDER BY thread_count DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var threadCount int
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &threadCount)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// Additional methods needed by user.go
func (dm *DatabaseManager) CreateThreadLikeOnCreation(userID, threadID int) error {
	_, err := dm.db.Exec("INSERT INTO threadlikes(type, user_id, thread_id) VALUES(?, ?, ?)", "creator", userID, threadID)
	return err
}

func (dm *DatabaseManager) CreateThreadDislikeOnCreation(userID, threadID int) error {
	_, err := dm.db.Exec("INSERT INTO threaddislikes(type, user_id, thread_id) VALUES(?, ?, ?)", "creator", userID, threadID)
	return err
}

func (dm *DatabaseManager) CreatePostLikeOnCreation(userID, postID int) error {
	_, err := dm.db.Exec("INSERT INTO likedposts(type, user_id, post_id) VALUES(?, ?, ?)", "creator", userID, postID)
	return err
}

func (dm *DatabaseManager) CreatePostDislikeOnCreation(userID, postID int) error {
	_, err := dm.db.Exec("INSERT INTO dislikes(type, user_id, post_id) VALUES(?, ?, ?)", "creator", userID, postID)
	return err
}
