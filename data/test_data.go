package data

import (
	"fmt"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var testDm *DatabaseManager

func InitTestDM(dm *DatabaseManager) {
	// Initialize test data
	testDm = dm
}

func TestEncrypt(plaintext string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func TestSortThreadsByLikes() {
	threads, err := testDm.GetAllThreadsByLikes()
	if err != nil {
		fmt.Printf("GetAllThreadsByLikes error: %v\n", err)
		return
	}

	fmt.Printf("Found %d threads sorted by likes:\n", len(threads))
	for i, thread := range threads {
		fmt.Printf("%d. %s (Likes: %d)\n", i+1, thread.Topic, thread.LikesCount)
	}
}

func Testing_data(t *testing.T) {
	user := User{
		Name:              "TestUser2",
		Email:             "test2@example.com",
		Password:          Encrypt("password123"),
		PreferedCategory1: "AI-theme",
		PreferedCategory2: "Creativity",
	}
	err := user.Create()
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}
	fmt.Printf("User created successfully! ID: %d, UUID: %s\n", user.Id, user.Uuid)

	retrievedUser, err := UserByEmail("test2@example.com")
	if err != nil {
		log.Fatal("Failed to retrieve user:", err)
	}
	fmt.Printf("User retrieved successfully! ID: %d, Name: %s, Email: %s, PrefCat1: %s, PrefCat2: %s\n",
		retrievedUser.Id, retrievedUser.Name, retrievedUser.Email, retrievedUser.PreferedCategory1, retrievedUser.PreferedCategory2)

	TestSortThreadsByLikes()
}
