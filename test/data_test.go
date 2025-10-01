package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"forum/internal"
	"forum/internal/data"
	"forum/models"
	"forum/utils"
)

func StartLogger() {
	file, err := os.OpenFile("casual-talk-test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestDeleteUserByName(t *testing.T) {
	// Initialize test database manager
	dm, err := data.NewDatabaseManager("../mydb.db")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer dm.Close()

	// Create a user to delete
	user := models.User{
		Name:              "DeleteMe",
		Email:             "deleteme@example.com",
		Password:          utils.Encrypt("DeletePass123"),
		PreferedCategory1: "Tech",
		PreferedCategory2: "Science",
	}
	err = dm.CreateUser(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete user by name
	err = dm.DeleteUserByName("DeleteMe")
	if err != nil {
		t.Fatalf("Failed to delete user by name: %v", err)
	}

	// Try to retrieve deleted user
	deletedUser, err := dm.GetUserByName("DeleteMe")
	if err == nil && deletedUser != (models.User{}) && deletedUser.Name != "" {
		t.Error("User was not deleted:", deletedUser)
	} else {
		t.Log("User with name 'DeleteMe' deleted (if existed)")
	}
}

func DeleteUserByName(dm *data.DatabaseManager, name string) error {

	err := dm.DeleteUserByName(name)
	if err != nil {
		fmt.Println("Error deleting user by name:", err)
	}
	return err
}

func DeleteUserById(dm *data.DatabaseManager, name string) error {
	user, err := dm.GetUserByName(name)
	if err != nil {
		fmt.Println("Error retrieving user by name:", err)
		return err
	}
	if user.Id == 0 {
		fmt.Println("User not found:", name)
		return fmt.Errorf("user not found: %s", name)
	}
	err = dm.DeleteUserById(&user)
	if err != nil {
		fmt.Println("Error deleting user by ID:", err)
	}
	return err
}

func TestUpdateUser(t *testing.T) {
	StartLogger()

	// Initialize test database manager
	dm, err := data.NewDatabaseManager("../mydb.db")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer dm.Close()
	internal.InitUserDM(dm)

	// Create a user to update
	user := models.User{
		// Id: 			  0, // sjhould be auto-assigned, but can trigger +1 id
		Name:              "UpdateUser1",
		Email:             "updateuser2@example.com",
		Password:          utils.Encrypt("UpdatePass123"),
		PreferedCategory1: "Tech",
		PreferedCategory2: "Science",
	}

	// DeleteUserById(dm, "TestUserData")
	// return
	err = dm.CreateUser(&user)

	if err != nil {

		t.Fatalf("Failed to create user: %v", err)
	}

	newEmail := "ok@example.com"
	newName := "ok"
	userTo, err := dm.GetUserByEmail("updateuser2@example.com")

	err = internal.TryUpdate(newName, userTo.Id)
	if err != nil {
		log.SetPrefix("[WARN] ")
		Warn("Failed to update user:", err)
		t.Fatalf("Failed to update user: %v", err)
	}

	// Retrieve updated user
	updatedUser, err := dm.GetUserByEmail("updateuser2@example.com")
	if err != nil {
		t.Fatalf("Failed to retrieve updated user: %v", err)
	}
	if updatedUser.Name != newName {
		t.Errorf("User update failed. Got Name: %s, Email: %s; Want Name: %s, Email: %s",
			updatedUser.Name, updatedUser.Email, newName, newEmail)
	} else {
		t.Logf("User updated successfully! ID: %d, Name: %s, Email: %s", updatedUser.Id, updatedUser.Name, updatedUser.Email)
		DeleteUserByName(dm, "ok")
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
	user := models.User{
		Name:              "TestUserData",
		Email:             "TestUserData@example.com",
		Password:          utils.Encrypt("Pass123!@#"),
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
	DeleteUserByName(dm, "TestUserData")
}
