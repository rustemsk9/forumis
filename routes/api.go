package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"forum/data"
	"forum/utils"
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
	// dbManager := GetDatabaseManager(request)
	// if dbManager == nil {
	// 	utils.InternalServerError(writer, request, fmt.Errorf("database not available"))
	// 	return
	// }

	// Extract thread ID from URL path
	var thread *data.Thread
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/counts
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Get likes and dislikes count using DatabaseManager
	likesCount, err := thread.GetThreadLikesCount(threadId)
	if err != nil {
		utils.InternalServerError(writer, request, err)
		return
	}

	dislikesCount, err := thread.GetThreadDislikesCount(threadId)
	if err != nil {
		utils.InternalServerError(writer, request, err)
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
		utils.BadRequest(writer, request, "Method not allowed")
		return
	}

	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/like
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}

	// Apply the like using smart function
	err = dbManager.SmartApplyThreadLike(user.Id, threadId)
	if err != nil {
		utils.InternalServerError(writer, request, err)
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
		utils.BadRequest(writer, request, "Method not allowed")
		return
	}

	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/dislike
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}

	// Apply the dislike using smart function
	err = dbManager.SmartApplyThreadDislike(user.Id, threadId)
	if err != nil {
		utils.InternalServerError(writer, request, err)
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
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	threadIdStr := parts[3] // /api/thread/{id}/status
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		utils.NotFound(writer, request)
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
		utils.BadRequest(writer, request, "Method not allowed")
		return
	}

	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Extract thread ID from URL path
	path := request.URL.Path

	parts := strings.Split(path, "/")
	threadIdStr := parts[3] // /api/thread/{id}/vote
	threadId, err := strconv.Atoi(threadIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid thread ID format")
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	// Check if thread exists
	_, err = dbManager.GetThreadByID(threadId)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}

	// Parse form to get vote type
	err = request.ParseForm()
	if err != nil {
		utils.BadRequest(writer, request, "Cannot parse form data")
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
		utils.InternalServerError(writer, request, err)
		return
	}

	// Check if this is an AJAX request by looking for Accept header or Fetch API
	acceptHeader := request.Header.Get("Accept")
	isAjax := strings.Contains(acceptHeader, "application/json") ||
		request.Header.Get("X-Requested-With") == "XMLHttpRequest"

	if isAjax {
		// Return JSON response for AJAX
		likesCount, _ := dbManager.GetThreadLikesCount(threadId)
		dislikesCount, _ := dbManager.GetThreadDislikesCount(threadId)
		userLiked := dbManager.HasThreadLiked(user.Id, threadId)
		userDisliked := dbManager.HasThreadDisliked(user.Id, threadId)

		response := ThreadVoteStatus{
			Likes:        likesCount,
			Dislikes:     dislikesCount,
			UserLiked:    userLiked,
			UserDisliked: userDisliked,
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Traditional form submission - redirect back
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
		utils.BadRequest(writer, request, "Method not allowed")
		return
	}

	// Extract post ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	postIdStr := parts[3] // /api/post/{id}/like
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid post ID format")
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	// Get database manager
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Verify the post exists
	_, err = dbManager.GetPostByID(postId)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}

	// Apply the like using smart function
	err = dbManager.SmartApplyPostLike(user.Id, postId)
	if err != nil {
		utils.InternalServerError(writer, request, err)
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
		utils.BadRequest(writer, request, "Method not allowed")
		return
	}

	// Extract post ID from URL path
	path := request.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	postIdStr := parts[3] // /api/post/{id}/dislike
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid post ID format")
		return
	}

	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}

	// Get database manager
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Verify the post exists
	_, err = dbManager.GetPostByID(postId)
	if err != nil {
		utils.NotFound(writer, request)
		return
	}

	// Apply the dislike using smart function
	err = dbManager.SmartApplyPostDislike(user.Id, postId)
	if err != nil {
		utils.InternalServerError(writer, request, err)
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
		utils.BadRequest(writer, request, "Invalid URL format")
		return
	}

	postIdStr := parts[3] // /api/post/{id}/status
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		utils.BadRequest(writer, request, "Invalid post ID format")
		return
	}

	// Get current user (may be nil for unauthenticated users)
	user := GetCurrentUser(request)

	// Get database manager
	dbManager := GetDatabaseManager(request)
	if dbManager == nil {
		utils.InternalServerError(writer, request, fmt.Errorf("database connection unavailable"))
		return
	}

	// Verify the post exists
	_, err = dbManager.GetPostByID(postId)
	if err != nil {
		utils.NotFound(writer, request)
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

func ApiStatus(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		utils.BadRequest(writer, request, "Method not allowed")
		return
	}
	// Get current user from middleware
	user := GetCurrentUser(request)
	if user == nil {
		utils.Unauthorized(writer, request, "Authentication required")
		return
	}
}
