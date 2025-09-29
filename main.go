package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"forum/data"
	"forum/routes"
	"forum/utils"

	_ "github.com/mattn/go-sqlite3"
)

type Configuration struct {
	Address      string
	ReadTimeout  int64
	WriteTimeout int64
	Static       string
}

var config Configuration

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
	fmt.Println("Initialized with configuration:\n", config)

}

func main() {
	dbManager, err := data.NewDatabaseManager("mydb.db")
	if err != nil {
		utils.Danger("Cannot initialize database manager", err)
		return
	}
	defer dbManager.Close()

	x := os.Args
	for _, arg := range x {
		if arg == "--migrate" {
			if err := data.RunMigrations(dbManager); err != nil {
				utils.Danger("Migration error:", err)
				return
			}
			fmt.Println("Database migrations applied successfully.")
		}
	}

	data.InitAllDatabaseManagers(dbManager)

	mux := http.NewServeMux()
	files := http.FileServer(http.Dir(config.Static))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	baseChain := routes.Chain(
		routes.WithErrorRecovery(),
		routes.WithLogging(),
		routes.WithDatabaseManager(dbManager),
		routes.WithAuthentication(),
	)

	authChain := routes.Chain(
		routes.WithErrorRecovery(),
		routes.WithLogging(),
		routes.WithDatabaseManager(dbManager),
		routes.WithAuthentication(),
		routes.RequireAuth(),
	)

	dataLS := data.LoginSkin{}

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

	mux.HandleFunc("/thread/new", authChain(routes.NewThread))
	mux.HandleFunc("/thread/create", authChain(routes.CreateThread))
	mux.HandleFunc("/thread/post", authChain(routes.PostThread))

	mux.HandleFunc("/thread/read", baseChain(routes.ReadThread))
	mux.HandleFunc("/account", baseChain(routes.ReadThreadsFromAccount))
	mux.HandleFunc("/accountcheck", baseChain(routes.AccountCheck))
	mux.HandleFunc("/debug", baseChain(routes.DebugPage))

	mux.HandleFunc("/back/", baseChain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/back/thread/") {
			if strings.HasSuffix(path, "/counts") {
				routes.GetThreadCounts(w, r)
			} else if strings.HasSuffix(path, "/status") {
				routes.GetThreadVoteStatus(w, r)
			} else if strings.HasSuffix(path, "/vote") {
				routes.VoteThread(w, r)
			} else {
				utils.NotFound(w, r)
			}
		} else if strings.HasPrefix(path, "/back/") {
			if !strings.HasSuffix(path, "/back") {
				utils.NotFound(w, r)
			}
		}

	}))
	mux.HandleFunc("/api/", baseChain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/post/") {
			if strings.HasSuffix(path, "/like") {
				routes.LikePost(w, r)
			} else if strings.HasSuffix(path, "/dislike") {
				routes.DislikePost(w, r)
			} else if strings.HasSuffix(path, "/status") {
				routes.GetPostVoteStatus(w, r)
			} else {
				utils.NotFound(w, r)
			}
		} else {
			utils.NotFound(w, r)
		}
	}))

	server := &http.Server{
		Addr:           config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}

	// server.ListenAndServe()
	// fmt.Println("Server started : on 8080")

	// Setup signal channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
	}()
	fmt.Println("Server started : on", config.Address)

	// Wait for interrupt signal
	<-stop
	fmt.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v\n", err)
	} else {
		fmt.Println("Server exited properly")
	}

	// Close database connection
	dbManager.Close()
	err = dbManager.Ping()
	if err != nil {
		fmt.Println("Database is closed or unreachable:", err)
	} else {
		fmt.Println("Database connection is open.")
	}
}
