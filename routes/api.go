package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

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

// GET /api/thread/{id}/counts
func GetThreadCounts(writer http.ResponseWriter, request *http.Request) {
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/counts
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		http.Error(writer, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Get likes and dislikes count using DatabaseManager
	likesCount, err := dbManager.GetThreadLikesCount(threadId)
	if err != nil {
		http.Error(writer, "Failed to get likes count", http.StatusInternalServerError)
		return
	}

	dislikesCount, err := dbManager.GetThreadDislikesCount(threadId)
	if err != nil {
		http.Error(writer, "Failed to get dislikes count", http.StatusInternalServerError)
		return
	}

	counts := ThreadCounts{
		Likes:    likesCount,
		Dislikes: dislikesCount,
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(counts)
}

// POST /api/thread/{id}/like
func LikeThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/like
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		http.Error(writer, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Apply the like using smart function
	err = dbManager.SmartApplyThreadLike(user.Id, threadId)
	if err != nil {
		http.Error(writer, "Failed to process like", http.StatusInternalServerError)
		return
	}

	// Get updated counts
	likesCount, _ := dbManager.GetThreadLikesCount(threadId)
	dislikesCount, _ := dbManager.GetThreadDislikesCount(threadId)

	// Return updated counts with vote status
	status := ThreadVoteStatus{
		Likes:        likesCount,
		Dislikes:     dislikesCount,
		UserLiked:    dbManager.HasThreadLiked(user.Id, threadId),
		UserDisliked: dbManager.HasThreadDisliked(user.Id, threadId),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}

// POST /api/thread/{id}/dislike
func DislikeThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/dislike
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		http.Error(writer, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Apply the dislike using smart function
	err = dbManager.SmartApplyThreadDislike(user.Id, threadId)
	if err != nil {
		http.Error(writer, "Failed to process dislike", http.StatusInternalServerError)
		return
	}

	// Get updated counts
	likesCount, _ := dbManager.GetThreadLikesCount(threadId)
	dislikesCount, _ := dbManager.GetThreadDislikesCount(threadId)

	// Return updated counts with vote status
	status := ThreadVoteStatus{
		Likes:        likesCount,
		Dislikes:     dislikesCount,
		UserLiked:    dbManager.HasThreadLiked(user.Id, threadId),
		UserDisliked: dbManager.HasThreadDisliked(user.Id, threadId),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}

// GET /api/thread/{id}/status
func GetThreadVoteStatus(writer http.ResponseWriter, request *http.Request) {
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/status
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		http.Error(writer, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Get current user (may be nil for unauthenticated users)
	user := GetCurrentUser(request)

	// Get counts
	likesCount, _ := dbManager.GetThreadLikesCount(threadId)
	dislikesCount, _ := dbManager.GetThreadDislikesCount(threadId)

	// Return vote status (even for unauthenticated users, just without personal vote info)
	status := ThreadVoteStatus{
		Likes:        likesCount,
		Dislikes:     dislikesCount,
		UserLiked:    user != nil && dbManager.HasThreadLiked(user.Id, threadId),
		UserDisliked: user != nil && dbManager.HasThreadDisliked(user.Id, threadId),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}

// POST /api/thread/{id}/vote - Generic vote endpoint
func VoteThread(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path

	parts := strings.Split(path, "/")
	threadIdStr := parts[3] // /api/thread/{id}/vote
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		http.Error(writer, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		http.Redirect(writer, request, "/login", http.StatusSeeOther)
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Parse form to get vote type
	err = request.ParseForm()
	if err != nil {
		http.Error(writer, "Invalid form data", http.StatusBadRequest)
		return
	}

	voteType := request.FormValue("vote_type")

	// Apply the appropriate vote
	switch voteType {
	case "like":
		err = dbManager.SmartApplyThreadLike(user.Id, threadId)
	case "dislike":
		err = dbManager.SmartApplyThreadDislike(user.Id, threadId)
	default:
		http.Redirect(writer, request, request.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	if err != nil {
		http.Error(writer, "Failed to process vote", http.StatusInternalServerError)
		return
	}

	// Redirect back to where the user came from or to the thread page
	redirectTo := request.Header.Get("Referer")
	if redirectTo == "" {
		redirectTo = "/#thread-" + strconv.Itoa(threadId-3) + "#thread-" + strconv.Itoa(threadId+3)
	} else if !strings.Contains(redirectTo, "#thread-") {
		// If anchor is missing, add it
		redirectTo += "#thread-" + strconv.Itoa(threadId)
	}
	http.Redirect(writer, request, redirectTo, http.StatusSeeOther)
}

// POST /api/post/{id}/like
func LikePost(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	postIdStr := parts[3] // /api/post/{id}/like
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(writer, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get database manager
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Verify the post exists
	_, err = dbManager.GetPostByID(postId)
	if err != nil {
		http.Error(writer, "Post not found", http.StatusNotFound)
		return
	}

	// Apply the like using smart function
	err = dbManager.SmartApplyPostLike(user.Id, postId)
	if err != nil {
		http.Error(writer, "Failed to process like", http.StatusInternalServerError)
		return
	}

	// Return updated counts with vote status
	likes, _ := dbManager.GetPostLikesCount(postId)
	dislikes, _ := dbManager.GetPostDislikesCount(postId)
	userLiked, _ := dbManager.HasUserLikedPost(user.Id, postId)
	userDisliked, _ := dbManager.HasUserDislikedPost(user.Id, postId)

	status := ThreadVoteStatus{
		Likes:        likes,
		Dislikes:     dislikes,
		UserLiked:    userLiked,
		UserDisliked: userDisliked,
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}

// POST /api/post/{id}/dislike
func DislikePost(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	postIdStr := parts[3] // /api/post/{id}/dislike
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(writer, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get database manager
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Verify the post exists
	_, err = dbManager.GetPostByID(postId)
	if err != nil {
		http.Error(writer, "Post not found", http.StatusNotFound)
		return
	}

	// Apply the dislike using smart function
	err = dbManager.SmartApplyPostDislike(user.Id, postId)
	if err != nil {
		http.Error(writer, "Failed to process dislike", http.StatusInternalServerError)
		return
	}

	// Return updated counts with vote status
	likes, _ := dbManager.GetPostLikesCount(postId)
	dislikes, _ := dbManager.GetPostDislikesCount(postId)
	userLiked, _ := dbManager.HasUserLikedPost(user.Id, postId)
	userDisliked, _ := dbManager.HasUserDislikedPost(user.Id, postId)

	status := ThreadVoteStatus{
		Likes:        likes,
		Dislikes:     dislikes,
		UserLiked:    userLiked,
		UserDisliked: userDisliked,
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}

// GET /api/post/{id}/status
func GetPostVoteStatus(writer http.ResponseWriter, request *http.Request) {
	// Extract post ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	postIdStr := parts[3] // /api/post/{id}/status
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(writer, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get current user (may be nil for unauthenticated users)
	user := GetCurrentUser(request)

	// Get database manager
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		http.Error(writer, "Database not available", http.StatusInternalServerError)
		return
	}

	// Verify the post exists
	_, err = dbManager.GetPostByID(postId)
	if err != nil {
		http.Error(writer, "Post not found", http.StatusNotFound)
		return
	}

	// Get counts and user vote status
	likes, _ := dbManager.GetPostLikesCount(postId)
	dislikes, _ := dbManager.GetPostDislikesCount(postId)

	var userLiked, userDisliked bool
	if user != nil {
		userLiked, _ = dbManager.HasUserLikedPost(user.Id, postId)
		userDisliked, _ = dbManager.HasUserDislikedPost(user.Id, postId)
	}

	// Return vote status (even for unauthenticated users, just without personal vote info)
	status := ThreadVoteStatus{
		Likes:        likes,
		Dislikes:     dislikes,
		UserLiked:    userLiked,
		UserDisliked: userDisliked,
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}
