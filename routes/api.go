package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"forum/data"
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

	// Get the thread to access its methods
	thread, err := data.ThreadById(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	counts := ThreadCounts{
		Likes:    thread.GetLikesCount(),
		Dislikes: thread.GetDislikesCount(),
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

	// Get user ID from cookie
	userId := data.GetCookieValue(request)
	if userId <= 0 {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the thread
	thread, err := data.ThreadById(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Apply the like using smart function
	err = data.SmartApplyThreadLike(userId, threadId)
	if err != nil {
		http.Error(writer, "Failed to process like", http.StatusInternalServerError)
		return
	}

	// Return updated counts with vote status
	status := ThreadVoteStatus{
		Likes:        thread.GetLikesCount(),
		Dislikes:     thread.GetDislikesCount(),
		UserLiked:    data.HasUserLikedThread(userId, threadId),
		UserDisliked: data.HasUserDislikedThread(userId, threadId),
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

	// Get user ID from cookie
	userId := data.GetCookieValue(request)
	if userId <= 0 {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the thread
	thread, err := data.ThreadById(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Apply the dislike using smart function
	err = data.SmartApplyThreadDislike(userId, threadId)
	if err != nil {
		http.Error(writer, "Failed to process dislike", http.StatusInternalServerError)
		return
	}

	// Return updated counts with vote status
	status := ThreadVoteStatus{
		Likes:        thread.GetLikesCount(),
		Dislikes:     thread.GetDislikesCount(),
		UserLiked:    data.HasUserLikedThread(userId, threadId),
		UserDisliked: data.HasUserDislikedThread(userId, threadId),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}

// GET /api/thread/{id}/status
func GetThreadVoteStatus(writer http.ResponseWriter, request *http.Request) {
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

	// Get user ID from cookie
	userId := data.GetCookieValue(request)

	// Get the thread to access its methods
	thread, err := data.ThreadById(threadId)
	if err != nil {
		http.Error(writer, "Thread not found", http.StatusNotFound)
		return
	}

	// Return vote status (even for unauthenticated users, just without personal vote info)
	status := ThreadVoteStatus{
		Likes:        thread.GetLikesCount(),
		Dislikes:     thread.GetDislikesCount(),
		UserLiked:    userId > 0 && data.HasUserLikedThread(userId, threadId),
		UserDisliked: userId > 0 && data.HasUserDislikedThread(userId, threadId),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
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

	// Get user ID from cookie
	userId := data.GetCookieValue(request)
	if userId <= 0 {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the post
	post, err := data.PostById(postId)
	if err != nil {
		http.Error(writer, "Post not found", http.StatusNotFound)
		return
	}

	// Apply the like using smart function
	err = data.SmartApplyPostLike(userId, postId)
	if err != nil {
		http.Error(writer, "Failed to process like", http.StatusInternalServerError)
		return
	}

	// Return updated counts with vote status
	status := ThreadVoteStatus{
		Likes:        post.GetLikesCount(),
		Dislikes:     post.GetDislikesCount(),
		UserLiked:    data.HasUserLikedPost(userId, postId),
		UserDisliked: data.HasUserDislikedPost(userId, postId),
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

	// Get user ID from cookie
	userId := data.GetCookieValue(request)
	if userId <= 0 {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the post
	post, err := data.PostById(postId)
	if err != nil {
		http.Error(writer, "Post not found", http.StatusNotFound)
		return
	}

	// Apply the dislike using smart function
	err = data.SmartApplyPostDislike(userId, postId)
	if err != nil {
		http.Error(writer, "Failed to process dislike", http.StatusInternalServerError)
		return
	}

	// Return updated counts with vote status
	status := ThreadVoteStatus{
		Likes:        post.GetLikesCount(),
		Dislikes:     post.GetDislikesCount(),
		UserLiked:    data.HasUserLikedPost(userId, postId),
		UserDisliked: data.HasUserDislikedPost(userId, postId),
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

	// Get user ID from cookie
	userId := data.GetCookieValue(request)

	// Get the post to access its methods
	post, err := data.PostById(postId)
	if err != nil {
		http.Error(writer, "Post not found", http.StatusNotFound)
		return
	}

	// Return vote status (even for unauthenticated users, just without personal vote info)
	status := ThreadVoteStatus{
		Likes:        post.GetLikesCount(),
		Dislikes:     post.GetDislikesCount(),
		UserLiked:    userId > 0 && data.HasUserLikedPost(userId, postId),
		UserDisliked: userId > 0 && data.HasUserDislikedPost(userId, postId),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(status)
}
