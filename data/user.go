package data

import (
	"fmt"
	"net/http"
	"time"
)

type User struct {
	Id        int
	Uuid      string
	Name      string
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
	// LikedPosts []Likes
}

// create a new thread
func (user *User) CreateThread(topic string, body string, alsoid int, category1 string, category2 string) (soid int64, conv Thread, err error) {
	res, err := Db.Exec("INSERT INTO threads(uuid, topic, body, user_id, created_at, category1, category2) VALUES(?, ?, ?, ?, ?, ?, ?)", createUUID(), topic, body, alsoid, time.Now(), category1, category2)
	if err != nil {
		panic(err)
	}
	soid, _ = res.LastInsertId()
	return
}

// create a new post to a thread
func (user *User) CreatePost(conv Thread, body string, alsoid int) (soid int64, err error) { // (post Post, err error) {
	res, err := Db.Exec("INSERT INTO posts(uuid, body, user_id, thread_id, created_at) VALUES(?, ?, ?, ?, ?)", createUUID(), body, alsoid, conv.Id, time.Now())
	if err != nil {
		panic(err)
	}
	soid, _ = res.LastInsertId()
	return
}

func LikeOnThreadCreation(alsoid, alsoid2 int) (err error) {
	stmt, err := Db.Prepare("INSERT INTO threadlikes(type, user_id, thread_id) VALUES(?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	li := ThreadLikes{}
	err = stmt.QueryRow("creator", alsoid, alsoid2).Scan(&li.Type, &li.UserId, &li.ThreadId)
	return
}

func DislikeOnThreadCreation(alsoid, alsoid2 int) (err error) {
	stmt, err := Db.Prepare("INSERT INTO threaddislikes(type, user_id, thread_id) VALUES(?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	li := ThreadDislikes{}
	err = stmt.QueryRow("creator", alsoid, alsoid2).Scan(&li.Type, &li.UserId, &li.ThreadId)
	return
}

// when post created , creates instance of like and dislike
// but creator of this post cannot like or dislike it
func LikeOnPostCreation(alsoid, alsoid2 int) (err error) {
	stmt, err := Db.Prepare("INSERT INTO likedposts(type, user_id, post_id) VALUES(?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	li := Likes{}
	err = stmt.QueryRow("creator", alsoid, alsoid2).Scan(&li.Type, &li.UserId, &li.PostId)
	return
}

// when post created , creates instance of like and dislike
// but creator of this post cannot like or dislike it
func DislikeOnPostCreation(alsoid, alsoid2 int) (err error) {
	stmt, err := Db.Prepare("INSERT INTO dislikes(type, user_id, post_id) VALUES(?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	di := Dislikes{}
	err = stmt.QueryRow("creator", alsoid, alsoid2).Scan(&di.Type, &di.UserId, &di.PostId)
	return
}

// create a new session for an existing user
func (user *User) CreateSession() (session Session, err error) {
	// Delete existing sessions for this user
	Db.Exec("DELETE from sessions where user_id=?", user.Id)

	// Create the session UUID
	sessionUUID := createUUID()

	// Create cookie string value in format "userId&sessionUUID"
	cookieString := fmt.Sprintf("%d&%s", user.Id, sessionUUID)

	// Insert session with cookie_string field
	stmt, err := Db.Prepare(
		"INSERT INTO sessions(uuid, email, user_id, created_at, cookie_string, active_last) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	// Calculate current time as hour*100 + minute
	now := time.Now()
	currentTime := now.Hour()*100 + now.Minute()

	// Execute the INSERT statement
	_, err = stmt.Exec(sessionUUID, user.Email, user.Id, time.Now(), cookieString, currentTime)
	if err != nil {
		return
	}

	// Now retrieve the created session
	session = Session{
		Uuid:         sessionUUID,
		Email:        user.Email,
		UserId:       user.Id,
		CreatedAt:    time.Now(),
		CookieString: cookieString,
	}

	return
}

// get the session for an existing user
func (user *User) Session() (session Session, err error) {
	session = Session{}
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at, cookie_string, active_last FROM sessions WHERE user_id=?", user.Id).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.CookieString, &session.ActiveLast)
	return
}

// create a new user, save user info into the database
func (user *User) Create() (err error) {
	stmt, err := Db.Prepare(
		"INSERT INTO users(uuid, name, email, password, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(createUUID(), user.Name, user.Email, Encrypt(user.Password), time.Now()).
		Scan(&user.Id, &user.Uuid, &user.CreatedAt)
	if err != nil {
		fmt.Println("Error during user creation:", err)
		return err
	}
	return
}

// delete user from database
func (user *User) Delete() (err error) {
	stmt, err := Db.Prepare("DELETE FROM users WHERE id=?")
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.Id)
	return
}

// update user information in the database
func (user *User) Update() (err error) {
	statement := "UPDATE users SET name=?, email=? where id=?"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.Name, user.Email, user.Id)
	return
}

// delete all users from database
func UserDeleteAll() (err error) {
	_, err = Db.Exec("DELETE FROM users")
	return
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
	rows, err := Db.Query("SELECT id, uuid, name, email, password, created_at FROM users")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		user := User{}
		if err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt); err != nil {
			return
		}
		users = append(users, user)
	}
	return
}

// true if user disliked this post
func PrepareLikedPosts(userID, postID int) bool { //(userLikes []Likes) {

	rows, err := Db.Query("SELECT type, user_id, post_id FROM likedposts WHERE user_id=?", userID)
	if err != nil {
		fmt.Println("Error on PrepareLikedPosts")
		return false
	}
	defer rows.Close()
	for rows.Next() {
		user := Likes{}
		if err = rows.Scan(&user.Type, &user.UserId, &user.PostId); err != nil {
			return false
		}

		if user.PostId == postID {
			return true
		}
		// userLikes = append(userLikes, user)
	}
	return false
}

func PrepareThreadLikedPosts(userID, threadid int) bool { //(userLikes []Likes) {

	rows, err := Db.Query("SELECT type, user_id, thread_id FROM threadlikes WHERE user_id=?", userID)
	if err != nil {
		fmt.Println("Error on PrepareLikedPosts")
		return false
	}
	defer rows.Close()
	for rows.Next() {
		user := ThreadLikes{}
		if err = rows.Scan(&user.Type, &user.UserId, &user.ThreadId); err != nil {
			fmt.Println("BASE on PrepareThreadLikedPosts didnt work")
		}

		if user.ThreadId == threadid {
			return true
		}
		// userLikes = append(userLikes, user)
	}
	return false
}

func PrepareThreadDislikedPosts(userID, threadid int) bool { //(userLikes []Likes) {

	rows, err := Db.Query("SELECT type, user_id, thread_id FROM threaddislikes WHERE user_id=?", userID)
	if err != nil {
		fmt.Println("Error on PrepareLikedPosts")
		return false
	}
	defer rows.Close()
	for rows.Next() {
		user := ThreadDislikes{}
		if err = rows.Scan(&user.Type, &user.UserId, &user.ThreadId); err != nil {
			fmt.Println("BASE on PrepareThreadDislikedPosts didnt work")
		}

		if user.ThreadId == threadid {
			return true
		}
		// userLikes = append(userLikes, user)
	}
	return false
}

// true if user disliked this post
func PrepareDislikedPosts(userID, postID int) bool { // (userLikes []Likes) {
	rows, err := Db.Query("SELECT type, user_id, post_id FROM dislikes WHERE user_id=?", userID)
	if err != nil {
		fmt.Println("Error on PrepareLikedPosts")
		return false
	}
	defer rows.Close()
	for rows.Next() {
		user := Dislikes{}
		if err = rows.Scan(&user.Type, &user.UserId, &user.PostId); err != nil {
			return false
		}
		if user.PostId == postID {
			return true
		}
		// userLikes = append(userLikes, user)
	}
	return false
}

// get a single user given the email
func UserByEmail(email string) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE email=?", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return
}

// get a single user given the UUID
func UserByUUID(uuid string) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at FROM users WHERE uuid=?", uuid).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return
}

func SessionByUUID(uuid string) bool {
	// user = User{}
	rows, _ := Db.Query("SELECT id, uuid, name, email, password, created_at FROM sessions WHERE uuid = '" + uuid + "'")
	uuid2 := ""
	rows.Scan(&uuid2)
	return uuid != ""
}

// IfUserExist is func, check user is in db
func IfUserExist(email, name string) bool {
	rows, _ := Db.Query("select uuid from users where email = '" + email + "' or name = '" + name + "'")
	defer rows.Close()
	rows.Next()
	uuid := ""
	rows.Scan(&uuid)

	return uuid != ""
}
