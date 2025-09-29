package data

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"forum/utils"

	_ "github.com/mattn/go-sqlite3"
)

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

type LoginSkin struct {
	Submit string
	Signup string
	Name   string
	Email  string
	Error  string
}

type Session struct {
	Id           int
	Uuid         string
	Email        string
	UserId       int
	CreatedAt    time.Time
	CookieString string
	ActiveLast   int
}

type Thread struct {
	Id            int
	Uuid          string
	Topic         string
	Body          string
	UserId        int
	User          string
	Email         string
	CreatedAt     time.Time
	CreatedAtDate string
	NumReplies    int
	Len           int
	//
	Cards            []Post
	LikedPosts       []Post
	UserLikedThreads []Thread // Threads liked by the user
	LengthOfPosts    int
	ThreadCategory   int
	LikesCount       int
	DislikesCount    int
	UserLiked        bool
	UserDisliked     bool
	Category1        string
	Category2        string
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

type Post struct {
	Id            int
	Uuid          string
	Body          string
	UserId        int
	ThreadId      int
	CreatedAt     time.Time
	FormattedDate string // formatted creation date for template access
	User          string // User information for template access
}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func createUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		utils.Danger("Cannot generate UUID", err)
	}

	// 0x40 is reserved variant from RFC 4122

	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}

// hash plaintext with SHA-1, changed to bcrypt as in forum instructions.
func Encrypt(plaintext string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// InitAllDatabaseManagers initializes all global DatabaseManager instances
func InitAllDatabaseManagers(dm *DatabaseManager) {
	InitSessionDM(dm)
	InitStatsDM(dm)
	InitUserDM(dm)
	InitThreadDM(dm)
}

// RunMigrations ensures that all required tables are present.

func RunMigrations(db *DatabaseManager) error {
	stmts := []string{
		`DROP TABLE IF EXISTS dislikes;`,
		`DROP TABLE IF EXISTS likedposts;`,
		`DROP TABLE IF EXISTS threaddislikes;`,
		`DROP TABLE IF EXISTS threadlikes;`,
		`DROP TABLE IF EXISTS posts;`,
		`DROP TABLE IF EXISTS threads;`,
		`DROP TABLE IF EXISTS sessions;`,
		`DROP TABLE IF EXISTS users;`,

		`CREATE TABLE users (
		  id         INTEGER PRIMARY KEY AUTOINCREMENT,
		  uuid       varchar(64) not null unique,
		  name       varchar(64),
		  email      varchar(64) not null unique,
		  password   varchar(128) not null,
		  created_at timestamp not null,
		  prefered_category1 varchar(255) default '',
		  prefered_category2 varchar(255) default ''
		);`,

		`CREATE TABLE sessions (
		  id            INTEGER PRIMARY KEY AUTOINCREMENT,
		  uuid          varchar(64) not null unique,
		  email         varchar(64),
		  user_id       integer references users(id),
		  created_at    timestamp not null,
		  cookie_string varchar(255),
		  active_last   integer default 0
		);`,

		`CREATE TABLE threads (
		  id         INTEGER PRIMARY KEY AUTOINCREMENT,
		  uuid       varchar(64) not null unique,
		  topic      text,
		  body       text,
		  user_id    integer references users(id),
		  created_at timestamp not null,
		  category1  varchar(255) default '',
		  category2  varchar(255) default ''
		);`,

		`CREATE TABLE posts (
		  id         INTEGER PRIMARY KEY AUTOINCREMENT,
		  uuid       varchar(64) not null unique,
		  body       text,
		  user_id    integer references users(id),
		  thread_id  integer references threads(id),
		  created_at timestamp not null
		);`,

		`CREATE TABLE threadlikes (
		  id        INTEGER PRIMARY KEY AUTOINCREMENT,
		  type      varchar(50),
		  user_id   integer references users(id),
		  thread_id integer references threads(id)
		);`,

		`CREATE TABLE threaddislikes (
		  id        INTEGER PRIMARY KEY AUTOINCREMENT,
		  type      varchar(50),
		  user_id   integer references users(id),
		  thread_id integer references threads(id)
		);`,

		`CREATE TABLE likedposts (
		  id      INTEGER PRIMARY KEY AUTOINCREMENT,
		  type    varchar(50),
		  user_id integer references users(id),
		  post_id integer references posts(id)
		);`,

		`CREATE TABLE dislikes (
		  id      INTEGER PRIMARY KEY AUTOINCREMENT,
		  type    varchar(50),
		  user_id integer references users(id),
		  post_id integer references posts(id)
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.DoExec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}
