package models

import (
	"time"
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

type ThreadCounts struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
}

type ThreadVoteStatus struct {
	Likes        int  `json:"likes"`
	Dislikes     int  `json:"dislikes"`
	UserLiked    bool `json:"userLiked"`
	UserDisliked bool `json:"userDisliked"`
}

type LoginSkin struct {
	Submit string
	Signup string
	Name   string
	Email  string
	Error  string
}

type Thread struct {
	Id               int
	Uuid             string
	Topic            string
	Body             string
	UserId           int
	User             string
	Email            string
	CreatedAt        time.Time
	CreatedAtDate    string
	NumReplies       int
	Len              int
	Cards            []Post
	LikedPosts       []Post
	UserLikedThreads []Thread
	LengthOfPosts    int
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
