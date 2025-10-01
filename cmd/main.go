package main

import (
	"context"
	"encoding/json"
	"fmt"
	"forum/internal"
	"forum/routes"
	"forum/utils"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	file, err := os.Open("config/config.json")
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
	dbManager, err := internal.ConnectDB("pkg", "mydb.db")
	if err != nil {
		utils.Danger("Cannot connect to database", err)
		return
	}
	defer dbManager.Close()

	x := os.Args
	for _, arg := range x {
		if arg == "--migrate" {
			if err := internal.RunMigrations(dbManager); err != nil {
				utils.Danger("Migration error:", err)
				return
			}
			fmt.Println("Database migrations applied successfully.")
		}
	}

	internal.InitAllDatabaseManagers(dbManager)
	mux := http.NewServeMux()
	files := http.FileServer(http.Dir(config.Static))
	routes.CompleteRoutes(mux, files, dbManager)

	server := &http.Server{
		Addr:           config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "immutable, max-age=360")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
	})
	stop := make(chan os.Signal, 1) // Setup signal channel
	signal.Notify(stop, os.Interrupt)

	go func() { // Start server in a goroutine
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
	}()
	fmt.Println("Server started on -> localhost:8080\nPress Ctrl+C to stop\n")

	<-stop // Wait for interrupt signal
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Create context with timeout for shutdown
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v\n", err)
	} else {
		fmt.Println("Server exited properly")
	}

	dbManager.Close() // Close database connection
	err = dbManager.Ping()
	if err != nil {
		fmt.Println("Database is closed or unreachable:", err)
	} else {
		fmt.Println("Database connection is open.")
	}
}
