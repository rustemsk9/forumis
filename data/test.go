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

func TestData(t *testing.T) {
	user := User{
		Name:              "TestUser2",
		Email:             "test2@example.com",
		Password:          TestEncrypt("password123"),
		PreferedCategory1: "AI-theme",
		PreferedCategory2: "Creativity",
	}
	err := user.Create()
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	t.Logf("User created successfully! ID: %d, UUID: %s", user.Id, user.Uuid)

	retrievedUser, err := UserByEmail("test2@example.com")
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	t.Logf("User retrieved successfully! ID: %d, Name: %s, Email: %s, PrefCat1: %s, PrefCat2: %s",
		retrievedUser.Id, retrievedUser.Name, retrievedUser.Email, retrievedUser.PreferedCategory1, retrievedUser.PreferedCategory2)

	threads, err := testDm.GetAllThreadsByLikes()
	if err != nil {
		t.Fatalf("GetAllThreadsByLikes error: %v", err)
	}
	t.Logf("Found %d threads sorted by likes:", len(threads))
	for i, thread := range threads {
		t.Logf("%d. %s (Likes: %d)", i+1, thread.Topic, thread.LikesCount)
	}
}
