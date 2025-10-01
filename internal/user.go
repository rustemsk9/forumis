package internal

import (
	"fmt"
	"forum/internal/data"
	"forum/models"
)

// user DatabaseManager instance for user operations
var userDM *data.DatabaseManager

// InitUserDM initializes the user DatabaseManager for user operations
func InitUserDM(dm *data.DatabaseManager) {
	userDM = dm
}

// create a new thread
func CreateThread(topic string, body string, alsoid int, category1 string, category2 string) (soid int64, conv models.Thread, err error) {
	soid, err = userDM.CreateThreadByUser(topic, body, alsoid, category1, category2)
	return
}

func LikeOnThreadCreation(alsoid, alsoid2 int) (err error) {
	return userDM.CreateThreadLikeOnCreation(alsoid, alsoid2)
}

func DislikeOnThreadCreation(alsoid, alsoid2 int) (err error) {
	return userDM.CreateThreadDislikeOnCreation(alsoid, alsoid2)
}

// create a new session for an existing user
func CreateSession(user models.User) (session models.Session, err error) {
	return userDM.CreateSession(&user)
}

// create a new user, save user info into the database
func CreateUser(user models.User) (err error) {
	return userDM.CreateUser(&user)
}

// delete user from database
func Delete(user models.User) (err error) {
	return userDM.DeleteUserByID(user.Id)
}

func UpdateUserPreferences(userID int, category1, category2 string) (err error) {
	err = userDM.UpdateUserPreferences(userID, category1, category2)
	if err != nil {
		fmt.Println("Error updating user preferences in UpdateUserPreferences method")
	}
	return
}

// update user information in the database
func TryUpdate(newName string, userId int) error {
	check := IfUserExist("", newName) // just to suppress unused warning
	if check {
		return fmt.Errorf("username %s already exists", newName)
	} else {

		err := userDM.Update(newName, userId)
		if err != nil {
			fmt.Println("Error updating user in TryUpdate method")
		}
		return err
	}
}

// delete all users from database
func UserDeleteAll() (err error) {
	return userDM.DeleteAllUsers()
}

// get all users in the database and returns it
func Users() (users []models.User, err error) {
	return userDM.GetAllUsers()
}

// get a single user given the email
func UserByEmail(email string) (user models.User, err error) {
	return userDM.GetUserByEmailDetailed(email)
}

// get a single user given the UUID
func UserByUUID(uuid string) (user models.User, err error) {
	return userDM.GetUserByUUID(uuid)
}

// IfUserExist is func, check user is in db
func IfUserExist(email, name string) bool {
	exists, _ := userDM.CheckUserExists(email, name)
	return exists
}

// Additional functions needed by routes/account.go
func GetUserById(userID int) models.User {
	guser, err := userDM.GetUserByID(userID)
	if err != nil {
		return models.User{}
	}
	return guser
}

func GetUserPosts(userID int) ([]models.Post, error) {
	// This function should get all posts created by a specific user
	// We need to add this method to DatabaseManager
	return userDM.GetUserCreatedPosts(userID)
}

func GetUserLikedPosts(userID int) ([]models.Likes, error) {
	return userDM.GetUserLikedPosts(userID)
}

func AccountThreads(userID int) ([]models.Thread, error) {
	// This function should get all threads created by a specific user
	// We need to add this method to DatabaseManager
	return userDM.GetUserCreatedThreads(userID)
}

// HasThreadLiked(user.Id, threads[i].Id)
func HasThreadLiked(userId int, threadId int) bool {
	return userDM.HasThreadLiked(userId, threadId)
}

func HasThreadDisliked(userId int, threadId int) bool {
	return userDM.HasThreadDisliked(userId, threadId)
}
