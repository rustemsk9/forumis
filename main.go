package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"forum/utils"

	"forum/routes"

	_ "github.com/mattn/go-sqlite3"
)

type Configuration struct {
	Address      string
	ReadTimeout  int64
	WriteTimeout int64
	Static       string
}

var (
	config Configuration
	// db            *sql.Db
	err           error
	Authenticated = false
)

func init() {
	file, err := os.Open("config.json")
	if err != nil {
		utils.Danger("Cannot open config file", err)
	}
	decoder := json.NewDecoder(file)
	config = Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		utils.Danger("Cannot get configuration from file", err)
	}
	var num int
	num++
	if num == 1 {
		fmt.Println("init(): INITIALISED NUMBER is 1")
	} else {
		fmt.Println("init(): INITIALISED AGAIN")
	}
}

func main() {

	mux := http.NewServeMux()

	// handle static assets
	files := http.FileServer(http.Dir(config.Static))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	mapper := map[string]func(http.ResponseWriter, *http.Request){
		// defined in route/index.go
		"/":    routes.Index,
		"/err": routes.Err,

		// defined in route/auth.go
		"/login":          routes.Login,
		"/logout":         routes.Logout,
		"/signup":         routes.Signup,
		"/signup_account": routes.SignupAccount,
		"/authenticate":   routes.Authenticate,

		// defined in route/thread.go
		"/thread/new":    routes.NewThread,
		"/thread/create": routes.CreateThread,
		"/thread/post":   routes.PostThread,
		"/thread/read":   routes.ReadThread,

		// account cabinet
		"/account":      routes.ReadThreadsFromAccount,
		"/accountcheck": routes.AccountCheck,

		// like handlers
		"/likes":         routes.PostLike,
		"/likes/accept":  routes.AcceptLike,
		"/likes/dislike": routes.AcceptDislike,

		// threadlikes
		"/threadLikes":         routes.ThreadLikes,
		"/threadLikes/accept":  routes.ApplyThreadLikes,
		"/threadLikes/dislike": routes.ApplyThreadDislikes,
	}

	for pattern, handler := range mapper {
		mux.HandleFunc(pattern, handler)
	}

	server := &http.Server{
		Addr:           config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()
	fmt.Println("Server started : on 8080")
}
