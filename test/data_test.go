package test

import (
	"testing"

	"forum/data"
)

func TestDeleteUserByName(t *testing.T) {
	// Initialize test database manager
	dm, err := data.NewDatabaseManager("../mydb.db")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer dm.Close()

	dm.DeleteUserByID(1)
	user, err := dm.GetUserByID(1)
	t.Error("User with ID 2:", user)
	if user.Name == "" {
		t.Log("User with ID 2 deleted (if existed)", user)
	}
}

func TestData(t *testing.T) {
	// Initialize test database manager
	dm, err := data.NewDatabaseManager("../mydb.db")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer dm.Close()

	// Create a user using the real struct
	user := data.User{
		Name:              "TestUser4",
		Email:             "test4@example.com",
		Password:          data.Encrypt("Pass123!@#"),
		PreferedCategory1: "AI-theme",
		PreferedCategory2: "Creativity",
	}

	err = dm.CreateUser(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	t.Logf("User created successfully! ID: %d, UUID: %s", user.Id, user.Uuid)

	retrievedUser, err := dm.GetUserByEmail("test4@example.com")
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	t.Logf("User retrieved successfully! ID: %d, Name: %s, Email: %s, PrefCat1: %s, PrefCat2: %s",
		retrievedUser.Id, retrievedUser.Name, retrievedUser.Email, retrievedUser.PreferedCategory1, retrievedUser.PreferedCategory2)

	threadID, err := dm.CreateThread("Test Topic", "Test Body", retrievedUser.Id, 1, 1)
	if err != nil {
		t.Fatalf("Failed to create thread: %v", err)
	}
	t.Logf("Thread created successfully! ID: %d", threadID)

	threads, err := dm.GetAllThreads()
	if err != nil {
		t.Fatalf("Failed to get all threads: %v", err)
	}
	t.Logf("Found %d threads:", len(threads))
	for i, thread := range threads {
		t.Logf("%d. Topic: %s, Body: %s, Likes: %d", i+1, thread.Topic, thread.Body, thread.LikesCount)
	}
}
