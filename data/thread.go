package data

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Thread struct {
	Id         int
	Uuid       string
	Topic      string
	UserId     int
	CreatedAt  time.Time
	Category   string
	Cards      []Post
	LikedPosts []Post
	UserHere   int
}

type LikeProperties struct {
	Li []Likes
	Di []Dislikes
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

func GetCookieValue(request *http.Request) int {
	var alsoid int
	cook, err := request.Cookie("_cookie")
	if err != nil {
		fmt.Println("Error") // or redirect
	}
	cooPart := strings.Split(cook.Value, "&")
	alsoid, _ = strconv.Atoi(cooPart[0])
	return alsoid
}

func GetLikes(postID int) (Li []Likes) {
	rows, err := Db.Query("SELECT * FROM likedposts WHERE post_id=?", postID)
	defer rows.Close()
	if err != nil {
		return
	}

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
	defer rows.Close()
	if err != nil {
		return
	}
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
	return
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
	fmt.Println("done apply dislike")
	return
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

	return
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
	fmt.Println("done delete dislike")
	return
}

// format the CreateAt date to display nicely on the screen
func (thread *Thread) CreatedAtDate() string {
	return thread.CreatedAt.Format("Jan/2/2006 3:04pm")
}

// get the number of posts in a thread
func (thread *Thread) NumReplies() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM posts where thread_id=?", thread.Id)
	defer rows.Close()
	if err != nil {
		return
	}
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
		threads = append(threads, th)
	}
	return
}

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
	if !rows.Next() {
		return nil, err
	}

	for rows.Next() {
		post := 0
		if err = rows.Scan(&post); err != nil {
			return
		}
		likedPosts = append(likedPosts, post)
	}
	return
}
func GetFromLikedDB(allIds []int) (posts []Post, err error) {
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
		fmt.Println("Error on GetFromLikedDB")
		return nil, err
	}
	if !rows.Next() {
		fmt.Println("couldnt find any liked post")
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
