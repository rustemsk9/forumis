package data

import (
	"fmt"
	"net/http"
	"time"
)

// Global DatabaseManager instance for user operations
var userDM *DatabaseManager

// InitUserDM initializes the global DatabaseManager for user operations
func InitUserDM(dm *DatabaseManager) {
	userDM = dm
}

type User struct {
	Id                int
	Uuid              string
	Name              string
	Email             string
	Password          string
	Role              string
	CreatedAt         time.Time
	PreferedCategory1 string
	PreferedCategory2 string
	// LikedPosts []Likes
}

// create a new thread
func (user *User) CreateThread(topic string, body string, alsoid int, category1 string, category2 string) (soid int64, conv Thread, err error) {
	soid, err = userDM.CreateThreadByUser(topic, body, alsoid, category1, category2)
	return
}

// create a new post to a thread
func (user *User) CreatePost(conv Thread, body string, alsoid int) (soid int64, err error) { // (post Post, err error) {
	soid, err = userDM.CreatePostByUser(body, alsoid, conv.Id)
	return
}

func LikeOnThreadCreation(alsoid, alsoid2 int) (err error) {
	return userDM.CreateThreadLikeOnCreation(alsoid, alsoid2)
}

func DislikeOnThreadCreation(alsoid, alsoid2 int) (err error) {
	return userDM.CreateThreadDislikeOnCreation(alsoid, alsoid2)
}

// when post created , creates instance of like and dislike
// but creator of this post cannot like or dislike it
func LikeOnPostCreation(alsoid, alsoid2 int) (err error) {
	return userDM.CreatePostLikeOnCreation(alsoid, alsoid2)
}

// when post created , creates instance of like and dislike
// but creator of this post cannot like or dislike it
func DislikeOnPostCreation(alsoid, alsoid2 int) (err error) {
	return userDM.CreatePostDislikeOnCreation(alsoid, alsoid2)
}

// create a new session for an existing user
func (user *User) CreateSession() (session Session, err error) {
	return userDM.CreateSession(user)
}

// get the session for an existing user
func (user *User) Session() (session Session, err error) {
	// This is a basic implementation - you may need to adjust based on how sessions are retrieved by user ID
	rows, err := userDM.db.Query("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE user_id=?", user.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	}
	return
}

// create a new user, save user info into the database
func (user *User) Create() (err error) {
	return userDM.CreateUser(user)
}

// delete user from database
func (user *User) Delete() (err error) {
	return userDM.DeleteUserByID(user.Id)
}

// update user information in the database
func (user *User) Update() (err error) {
	_, err = userDM.db.Exec("UPDATE users SET name=?, email=? WHERE id=?", user.Name, user.Email, user.Id)
	return
}

// delete all users from database
func UserDeleteAll() (err error) {
	return userDM.DeleteAllUsers()
}

func CurrentUser(request *http.Request) (name string, err error) {
	// get from DB the user from the session
	session, err := SessionCheck(nil, request)
	if err != nil {
		return
	}
	user, err := session.User()
	if err != nil {
		fmt.Println("Error session.User() in CurrentUser")
		return
	}
	name = user.Name
	return
}

// get all users in the database and returns it
func Users() (users []User, err error) {
	return userDM.GetAllUsers()
}

// true if user liked this post
func PrepareLikedPosts(userID, postID int) bool {
	likes, err := userDM.GetUserLikedPosts(userID)
	if err != nil {
		fmt.Println("Error on PrepareLikedPosts")
		return false
	}

	for _, like := range likes {
		if like.PostId == postID {
			return true
		}
	}
	return false
}

func PrepareThreadLikedPosts(userID, threadid int) bool {
	likes, err := userDM.GetUserLikedThreads(userID)
	if err != nil {
		fmt.Println("Error on PrepareThreadLikedPosts")
		return false
	}

	for _, like := range likes {
		if like.ThreadId == threadid {
			return true
		}
	}
	return false
}

func PrepareThreadDislikedPosts(userID, threadid int) bool {
	dislikes, err := userDM.GetUserDislikedThreads(userID)
	if err != nil {
		fmt.Println("Error on PrepareThreadDislikedPosts")
		return false
	}

	for _, dislike := range dislikes {
		if dislike.ThreadId == threadid {
			return true
		}
	}
	return false
}

// true if user disliked this post
func PrepareDislikedPosts(userID, postID int) bool {
	dislikes, err := userDM.GetUserDislikedPosts(userID)
	if err != nil {
		fmt.Println("Error on PrepareDislikedPosts")
		return false
	}

	for _, dislike := range dislikes {
		if dislike.PostId == postID {
			return true
		}
	}
	return false
}

// get a single user given the email
func UserByEmail(email string) (user User, err error) {
	return userDM.GetUserByEmailDetailed(email)
}

// get a single user given the UUID
func UserByUUID(uuid string) (user User, err error) {
	return userDM.GetUserByUUID(uuid)
}

func SessionByUUID(uuid string) bool {
	// This function seems to check if a session exists with the given UUID
	// Using the DatabaseManager to check session validity
	_, valid, _ := userDM.ValidateSession(uuid)
	return valid
}

// IfUserExist is func, check user is in db
func IfUserExist(email, name string) bool {
	exists, _ := userDM.CheckUserExists(email, name)
	return exists
}

// Additional functions needed by routes/account.go
func GetUserById(userID int) User {
	user, err := userDM.GetUserByID(userID)
	if err != nil {
		return User{}
	}
	return user
}

func GetUserPosts(userID int) ([]Post, error) {
	// This function should get all posts created by a specific user
	// We need to add this method to DatabaseManager
	return userDM.GetUserCreatedPosts(userID)
}

func GetUserLikedPosts(userID int) ([]Likes, error) {
	return userDM.GetUserLikedPosts(userID)
}

func AccountThreads(userID int) ([]Thread, error) {
	// This function should get all threads created by a specific user
	// We need to add this method to DatabaseManager
	return userDM.GetUserCreatedThreads(userID)
}

func GetLikesPostsFromDB(likes []Likes) ([]Post, error) {
	// This function should get full post objects from the likes
	var posts []Post
	for _, like := range likes {
		post, err := userDM.GetPostByID(like.PostId)
		if err != nil {
			continue // Skip posts that can't be found
		}

		// Populate User information for the post
		user, err := userDM.GetPostUser(post.UserId)
		if err != nil {
			// If user loading fails, create a placeholder
			post.User = "Unknown User"
		} else {
			post.User = user.Name
		}

		// Note: CreatedAtDate formatting will be done by the template method

		posts = append(posts, post)
	}
	return posts, nil
}

// Additional functions needed by API routes
func ThreadById(threadID int) (Thread, error) {
	return userDM.GetThreadByID(threadID)
}

func FilterThreadsByCategories(category1, category2 string) ([]Thread, error) {
	// We need to add this method to DatabaseManager
	return userDM.GetThreadsByCategories(category1, category2)
}
