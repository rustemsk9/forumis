package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        int
	Uuid      string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

var Db *sql.DB

func init() {
	var err error
	Db, err = sql.Open("sqlite3", "../mydb.db")
	if err != nil {
		log.Fatal(err)
	}
}

// hash plaintext with SHA-1
func Encrypt(plaintext string) string {
    hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    return string(hash)
}

func (user *User) Create() (err error) {
	// Create UUID
	user.Uuid = uuid.Must(uuid.NewV4()).String()
	user.CreatedAt = time.Now()
	
	// Insert and get the ID
	result, err := Db.Exec("INSERT INTO users (uuid, name, email, password, created_at) VALUES (?, ?, ?, ?, ?)", 
		user.Uuid, user.Name, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		return err
	}
	
	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	user.Id = int(id)
	return nil
}

func UserByEmail(email string) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return
}

func main() {
	// Create a test user
	user := User{
		Name:     "TestUser2",
		Email:    "test2@example.com",
		Password: Encrypt("password123"),
	}
	
	err := user.Create()
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}
	
	fmt.Printf("User created successfully! ID: %d, UUID: %s\n", user.Id, user.Uuid)
	
	// Try to retrieve the user
	retrievedUser, err := UserByEmail("test2@example.com")
	if err != nil {
		log.Fatal("Failed to retrieve user:", err)
	}
	
	fmt.Printf("User retrieved successfully! ID: %d, Name: %s, Email: %s\n", 
		retrievedUser.Id, retrievedUser.Name, retrievedUser.Email)
}