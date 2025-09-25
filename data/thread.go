package data

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Thread struct {
	Id            int
	Uuid          string
	Topic         string
	UserId        int
	CreatedAt     time.Time
	Category      string
	Cards         []Post
	LikedPosts    []Post
	UserHere      int
	LikesCount    int
	DislikesCount int
}

type LikeProperties struct {
	Li []Likes
	Di []Dislikes
}

type ThreadLikeProperties struct {
	Li []ThreadLikes
	Di []ThreadDislikes
}

type Likes struct {
	Type          string
	UserId        int
	PostId        int
	LengthOfLikes int
	UserLiked     bool
}

type Dislikes struct {
	Type             string
	UserId           int
	PostId           int
	LengthOfDislikes int
	UserDisliked     bool
}

type ThreadLikes struct {
	Type          string
	UserId        int
	ThreadId      int
	LengthOfLikes int
	UserDisliked  bool
}

type ThreadDislikes struct {
	Type             string
	UserId           int
	ThreadId         int
	LengthOfDislikes int
	UserDisliked     bool
}

func GetCookieValue(request *http.Request) int {
	// Debug: Print all cookies
	fmt.Printf("DEBUG: All cookies for request: ")
	for _, cookie := range request.Cookies() {
		fmt.Printf("[%s=%s] ", cookie.Name, cookie.Value)
	}
	fmt.Println()

	// Get the _cookie from request
	cook, err := request.Cookie("_cookie")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			fmt.Println("Cookie '_cookie' not found")
		} else {
			fmt.Println("Error on Get Cookie:", err)
		}
		return -1
	}

	fmt.Printf("Found cookie: %s = %s\n", cook.Name, cook.Value)

	// Check if session exists in database with this cookie string
	session, err := GetSessionByCookie(cook.Value)
	if err != nil {
		fmt.Printf("No valid session found for cookie: %s\n", cook.Value)
		return -1
	}

	if session.UserId == 0 {
		fmt.Println("Session found but no valid user ID")
		return -1
	}

	fmt.Printf("Authenticated user ID %d via cookie string\n", session.UserId)
	return session.UserId
}

func GetThreadLikes(findid int) (THLI []ThreadLikes) {
	rows, err := Db.Query("SELECT * FROM threadlikes WHERE thread_id=?", findid)
	if err != nil {
		fmt.Println("Error on select GetThreadLikes")
		return
	}
	defer rows.Close()

	// likesAll := []Likes{}
	var likeLength int
	for rows.Next() {
		thLikes := ThreadLikes{}
		likeLength++
		if err = rows.Scan(&thLikes.Type, &thLikes.ThreadId, &thLikes.UserId); err != nil {
			return
		}
		THLI = append(THLI, thLikes)
	}
	for i := range THLI {
		THLI[i].LengthOfLikes = likeLength - 1
	}
	return
}

func GetThreadDislikes(findid int) (THLI []ThreadDislikes) {
	rows, err := Db.Query("SELECT * FROM threaddislikes WHERE thread_id=?", findid)
	if err != nil {
		fmt.Println("Error on select GetThreadDislikes")
		return
	}
	defer rows.Close()

	// likesAll := []Likes{}
	var likeLength int
	for rows.Next() {
		thLikes := ThreadDislikes{}
		likeLength++
		if err = rows.Scan(&thLikes.Type, &thLikes.ThreadId, &thLikes.UserId); err != nil {
			return
		}
		THLI = append(THLI, thLikes)
	}
	for i := range THLI {
		THLI[i].LengthOfDislikes = likeLength - 1
	}
	return
}
func ApplyThreadLike(stateLike string, userID int, threadId int) {
	fmt.Println("Apply thread likes in database proccess")
	var li ThreadLikes
	stmt, err := Db.Prepare("INSERT INTO threadlikes(type, user_id, thread_id) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(stateLike, userID, threadId).
		Scan(&li.Type, &li.UserId, &li.ThreadId)
	if err != nil {
		fmt.Println("Error on ApplyThreadLike:", err)
	}
	// can apply only if threadId is right // TODO check sequence

}

func ApplyThreadDislike(stateLike string, userID int, threadId int) {
	fmt.Println("Apply thread dislikes in database proccess")
	var li ThreadDislikes
	stmt, err := Db.Prepare("INSERT INTO threaddislikes(type, user_id, thread_id) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	// can apply only if threadId is right // TODO check sequence
	err = stmt.QueryRow(stateLike, userID, threadId).
		Scan(&li.Type, &li.UserId, &li.ThreadId)
	if err != nil {
		fmt.Println("Error on ApplyThreadDislike:", err)
	}
}

// Check if user has already liked a thread
func HasUserLikedThread(userID int, threadID int) bool {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM threadlikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Check if user has already disliked a thread
func HasUserDislikedThread(userID int, threadID int) bool {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM threaddislikes WHERE user_id=? AND thread_id=?", userID, threadID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Remove user's like from a thread
func RemoveThreadLike(userID int, threadID int) error {
	stmt, err := Db.Prepare("DELETE FROM threadlikes WHERE user_id=? AND thread_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID, threadID)
	return err
}

// Remove user's dislike from a thread
func RemoveThreadDislike(userID int, threadID int) error {
	stmt, err := Db.Prepare("DELETE FROM threaddislikes WHERE user_id=? AND thread_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID, threadID)
	return err
}

// Smart like function - handles vote switching
func SmartApplyThreadLike(userID int, threadID int) error {
	// Check if user already liked this thread
	if HasUserLikedThread(userID, threadID) {
		// User already liked, so remove the like (toggle off)
		return RemoveThreadLike(userID, threadID)
	}

	// Check if user disliked this thread, if so remove the dislike first
	if HasUserDislikedThread(userID, threadID) {
		err := RemoveThreadDislike(userID, threadID)
		if err != nil {
			return err
		}
	}

	// Add the like
	ApplyThreadLike("like", userID, threadID)
	return nil
}

// Smart dislike function - handles vote switching
func SmartApplyThreadDislike(userID int, threadID int) error {
	// Check if user already disliked this thread
	if HasUserDislikedThread(userID, threadID) {
		// User already disliked, so remove the dislike (toggle off)
		return RemoveThreadDislike(userID, threadID)
	}

	// Check if user liked this thread, if so remove the like first
	if HasUserLikedThread(userID, threadID) {
		err := RemoveThreadLike(userID, threadID)
		if err != nil {
			return err
		}
	}

	// Add the dislike
	ApplyThreadDislike("dislike", userID, threadID)
	return nil
}

// }
func GetLikes(postID int) (Li []Likes) {
	rows, err := Db.Query("SELECT * FROM likedposts WHERE post_id=?", postID)
	if err != nil {
		return
	}
	defer rows.Close()

	// likesAll := []Likes{}
	var likeLength int
	for rows.Next() {
		likes := Likes{}
		likeLength++
		if err = rows.Scan(&likes.Type, &likes.PostId, &likes.UserId); err != nil {
			return
		}
		Li = append(Li, likes)
	}
	for i := range Li {
		Li[i].LengthOfLikes = likeLength - 1
	}
	return
}

func GetDislikes(postID int) (Di []Dislikes) {
	rows, err := Db.Query("SELECT * FROM dislikes WHERE post_id=?", postID)
	if err != nil {
		return
	}
	defer rows.Close()
	var dislikeLength int
	for rows.Next() {
		dislikes := Dislikes{}
		dislikeLength++
		if err = rows.Scan(&dislikes.Type, &dislikes.PostId, &dislikes.UserId); err != nil {
			return
		}
		Di = append(Di, dislikes)
	}

	for i := range Di {
		Di[i].LengthOfDislikes = dislikeLength - 1
	}

	return
}

func ApplyLikes(stateLike string, userID int, postID int) {
	fmt.Println("---------------------")
	fmt.Println("in database proccess")
	var li Likes
	stmt, err := Db.Prepare("INSERT INTO likedposts(type, user_id, post_id) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(stateLike, userID, postID).
		Scan(&li.Type, &li.UserId, &li.PostId)
	if err != nil {
		fmt.Println("Error on ApplyLikes:", err)
	}
}

func ApplyDislikes(stateLike string, userID int, postID int) {
	fmt.Println("---------------------")
	fmt.Println("in database proccess")
	var di Dislikes
	stmt, err := Db.Prepare("INSERT INTO dislikes(type, user_id, post_id) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(stateLike, userID, postID).
		Scan(&di.Type, &di.UserId, &di.PostId)
	if err != nil {
		fmt.Println("Error on ApplyDislikes:", err)
		return
	}
}

func DeleteLikes(alsoid, postID int) {
	// row, err := Db.Query
	stmt, err := Db.Prepare("DELETE FROM likedposts WHERE user_id=? AND post_id=?")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(alsoid, postID)
	if err != nil {
		fmt.Println("Error on DeleteLikes:", err)
	}
}

func DeleteDislikes(alsoid, postID int) {
	// row, err := Db.Query
	stmt, err := Db.Prepare("DELETE FROM dislikes WHERE user_id=? AND post_id=?")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(alsoid, postID)
	if err != nil {
		fmt.Println("Error on DeleteDislikes:", err)
	}
}

func DeleteThreadDislikes(alsoid, threadid int) {
	// row, err := Db.Query
	stmt, err := Db.Prepare("DELETE FROM threaddislikes WHERE user_id=? AND thread_id=?")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(alsoid, threadid)
	if err != nil {
		fmt.Println("Error on DeleteThreadDislikes:", err)
	}
}

func DeleteThreadLikes(alsoid, threadid int) {
	// row, err := Db.Query
	stmt, err := Db.Prepare("DELETE FROM threadlikes WHERE user_id=? AND thread_id=?")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(alsoid, threadid)
	if err != nil {
		fmt.Println("Error on DeleteThreadLikes:", err)
	}
}

// format the CreateAt date to display nicely on the screen
func (thread *Thread) CreatedAtDate() string {
	return thread.CreatedAt.Format("Jan/2/2006 3:04pm")
}

// get the number of posts in a thread
func (thread *Thread) NumReplies() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM posts where thread_id=?", thread.Id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	return
}

// get posts to a thread
func (thread *Thread) Posts() (posts []Post, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE thread_id=?", thread.Id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := Post{}
		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt); err != nil {
			return
		}
		posts = append(posts, post)
	}
	return
}

// get the user who started this thread
func (thread *Thread) User() (user User) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", thread.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return
}

// get the number of likes for this thread
func (thread *Thread) GetLikesCount() int {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM threadlikes WHERE thread_id=?", thread.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// get the number of dislikes for this thread
func (thread *Thread) GetDislikesCount() int {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM threaddislikes WHERE thread_id=?", thread.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func GetUserById(passID int) (user User) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id=?", passID).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return
}

// get all threads in the database and returns it
func Threads() (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, topic, user_id, created_at, category FROM threads ORDER BY created_at DESC")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		th := Thread{}
		if err = rows.Scan(&th.Id, &th.Uuid, &th.Topic, &th.UserId, &th.CreatedAt, &th.Category); err != nil {
			return
		}
		// Populate likes and dislikes counts
		th.LikesCount = th.GetLikesCount()
		th.DislikesCount = th.GetDislikesCount()
		threads = append(threads, th)
	}
	return
}

// show /account threads created by user
func AccountThreads(alsoid int) (threads []Thread, err error) {
	fmt.Println("Account Threads started DB search")
	rows, err := Db.Query("SELECT id, uuid, topic, user_id, created_at, category FROM threads WHERE user_id=?", alsoid)
	if err != nil {
		fmt.Println("Error on AccountThreads")
		return
	}

	defer rows.Close()
	// if !rows.Next() {
	// 	fmt.Println("bool worked on empty")
	// 	return nil, err
	// }
	for rows.Next() {
		th := Thread{}
		if err = rows.Scan(&th.Id, &th.Uuid, &th.Topic, &th.UserId, &th.CreatedAt, &th.Category); err != nil {
			return
		}
		threads = append(threads, th)
	}
	if len(threads) == 0 {
		return nil, err
	}
	return
}

// shows posts written by user
func GetUserPosts(alsoid int) (posts []Post, err error) {
	fmt.Println("GetUserPosts started DB search")
	rows, err := Db.Query("SELECT Id, uuid, body, user_id, thread_id, created_at FROM posts WHERE user_id=?", alsoid)
	if err != nil {
		fmt.Println("Error on GetUserPosts")
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := Post{}
		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt); err != nil {
			return
		}
		posts = append(posts, post)
	}
	return
}

// shows liked posts in account profile
func GetUserLikedPosts(alsoid int) (likedPosts []int, err error) {
	fmt.Println("GetUserLikedPosts started DB search")
	rows, err := Db.Query("SELECT post_id FROM likedposts WHERE user_id=?", alsoid)
	if err != nil {
		fmt.Println("Error on GetUserLikedPosts")
		return
	}
	defer rows.Close()
	// if !rows.Next() {
	// 	return nil, err
	// }

	for rows.Next() {
		post := 0
		if err = rows.Scan(&post); err != nil {
			return
		}
		likedPosts = append(likedPosts, post)
	}
	return
}
func GetLikesPostsFromDB(allIds []int) (posts []Post, err error) {
	str := "SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE "
	for i, g := range allIds {
		if i == len(allIds)-1 {
			str += fmt.Sprintf("id=%d;", g)
			break
		}
		str += fmt.Sprintf("id=%d OR ", g)
	}
	// fmt.Println(str)
	rows, err := Db.Query(str)
	// fmt.Println(rows)
	if err != nil {
		fmt.Println("Error on GetLikesPostsFromDB")
		return nil, err
	}
	for rows.Next() {
		post := Post{}

		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt); err != nil {
			fmt.Println("error on Scan")
			return
		}

		posts = append(posts, post)
	}

	return
}

// get a thread by the UUID
func ThreadById(id int) (conv Thread, err error) {
	conv = Thread{}
	err = Db.QueryRow("SELECT id, uuid, topic, user_id, created_at FROM threads WHERE id=?", id).
		Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt)

	return
}
