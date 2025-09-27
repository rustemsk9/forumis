package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"forum/data"
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

var config        Configuration
	

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
		fmt.Println("[INIT]: INITIALISED NUMBER is 1")
	} else {
		fmt.Println("[INIT]: INITIALISED AGAIN")
	}
}

func main() {
	// Initialize database manager
	dbManager, err := data.NewDatabaseManager("mydb.db")
	if err != nil {
		utils.Danger("Cannot initialize database manager", err)
		return
	}
	defer dbManager.Close()

	mux := http.NewServeMux()
	files := http.FileServer(http.Dir(config.Static))
	mux.Handle("/static/", http.StripPrefix("/static/", files))
	
	// Create base middleware chain that applies to all routes
	baseChain := routes.Chain(
		routes.WithLogging(),
		routes.WithDatabaseManager(dbManager),
		routes.WithAuthentication(),
	)
	
	// Create authenticated routes chain
	authChain := routes.Chain(
		routes.WithLogging(),
		routes.WithDatabaseManager(dbManager),
		routes.WithAuthentication(),
		routes.RequireAuth(),
	)

	dataLS := data.LoginSkin{}
	
	// Public routes (no authentication required)
	mux.HandleFunc("/", baseChain(routes.Index))
	mux.HandleFunc("/err", baseChain(routes.Err))
	mux.HandleFunc("/login/", baseChain(func(w http.ResponseWriter, r *http.Request) {
		routes.Login(w, r, dataLS)
	}))
	mux.HandleFunc("/signup", baseChain(func(w http.ResponseWriter, r *http.Request) {
		routes.Signup(w, r, dataLS)
	}))
	mux.HandleFunc("/signup_account", baseChain(routes.SignupAccount))
	mux.HandleFunc("/authenticate", baseChain(routes.Authenticate))
	mux.HandleFunc("/logout", baseChain(routes.Logout))

	// Protected routes (require authentication)
	mux.HandleFunc("/thread/new", authChain(routes.NewThread))
	mux.HandleFunc("/thread/create", authChain(routes.CreateThread))
	mux.HandleFunc("/thread/post", authChain(routes.PostThread))
	
	// Semi-protected routes (different behavior based on auth status)
	mux.HandleFunc("/thread/read", baseChain(routes.ReadThread))
	mux.HandleFunc("/account", baseChain(routes.ReadThreadsFromAccount))
	mux.HandleFunc("/accountcheck", baseChain(routes.AccountCheck))
	mux.HandleFunc("/debug", baseChain(routes.DebugPage))

	// Add API routes with a custom handler for pattern matching
	// API routes with middleware
	mux.HandleFunc("/api/", baseChain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Handle thread API endpoints
		if strings.HasPrefix(path, "/api/thread/") {
			if strings.HasSuffix(path, "/counts") {
				routes.GetThreadCounts(w, r)
			} else if strings.HasSuffix(path, "/like") {
				routes.LikeThread(w, r)
			} else if strings.HasSuffix(path, "/dislike") {
				routes.DislikeThread(w, r)
			} else if strings.HasSuffix(path, "/status") {
				routes.GetThreadVoteStatus(w, r)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		} else if strings.HasPrefix(path, "/api/post/") {
			// Handle post API endpoints
			if strings.HasSuffix(path, "/like") {
				routes.LikePost(w, r)
			} else if strings.HasSuffix(path, "/dislike") {
				routes.DislikePost(w, r)
			} else if strings.HasSuffix(path, "/status") {
				routes.GetPostVoteStatus(w, r)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))

	server := &http.Server{
		Addr:           config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	server.ListenAndServe()
	fmt.Println("Server started : on 8080")
}
