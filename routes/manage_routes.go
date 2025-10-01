package routes

import (
	"forum/internal/data"
	"forum/models"
	"forum/utils"
	"net/http"
	"strings"
)

func CompleteRoutes(mux *http.ServeMux, files http.Handler, dbManager *data.DatabaseManager) {

	mux.Handle("/static/", http.StripPrefix("/static/", files))

	baseChain := Chain(
		WithErrorRecovery(),
		WithLogging(),
		WithDatabaseManager(dbManager), // if we turn off this option, 500 error occurs in auth and login, since no dbmanager in context
		WithAuthentication(),
	)

	authChain := Chain(
		WithErrorRecovery(),
		WithLogging(),
		WithDatabaseManager(dbManager),
		WithAuthentication(),
		RequireAuth(),
	) // authChain includes RequireAuth

	dataLS := models.LoginSkin{}

	mux.HandleFunc("/", baseChain(Index))
	mux.HandleFunc("/err", baseChain(Err))
	mux.HandleFunc("/login/", baseChain(func(w http.ResponseWriter, r *http.Request) {
		Login(w, r, dataLS)
	}))
	mux.HandleFunc("/signup", baseChain(func(w http.ResponseWriter, r *http.Request) {
		Signup(w, r, dataLS)
	}))
	mux.HandleFunc("/authenticate", baseChain(Authenticate))
	mux.HandleFunc("/logout", baseChain(Logout))

	mux.HandleFunc("/thread/new", authChain(NewThread))
	mux.HandleFunc("/thread/create", authChain(CreateThread))
	mux.HandleFunc("/thread/post", authChain(PostThread))
	mux.HandleFunc("/thread/read", baseChain(ReadThread))

	mux.HandleFunc("/account", baseChain(ReadThreadsFromAccount))
	mux.HandleFunc("/accountcheck", baseChain(AccountCheck))
	mux.HandleFunc("/debug", baseChain(DebugPage))

	mux.HandleFunc("/back/", baseChain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/back/thread/") {
			if strings.HasSuffix(path, "/counts") {
				GetThreadCounts(w, r)
			} else if strings.HasSuffix(path, "/status") {
				GetThreadVoteStatus(w, r)
			} else if strings.HasSuffix(path, "/vote") {
				VoteThread(w, r)
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
				LikePost(w, r)
			} else if strings.HasSuffix(path, "/dislike") {
				DislikePost(w, r)
			} else if strings.HasSuffix(path, "/status") {
				GetPostVoteStatus(w, r)
			} else {
				utils.NotFound(w, r)
			}
		} else {
			utils.NotFound(w, r)
		}
	}))
}
