package data

// format the CreatedAt date to display nicely on the screen
func (post *Post) CreatedAtDate() string {
	return post.CreatedAt.Format("Jan/2/2006 3:04pm")
}

// get the user who wrote the post using DatabaseManager (preferred method)
func (post *Post) UserWithDB(dm *DatabaseManager) (User, error) {
	return dm.GetPostUser(post.UserId)
}

// get the number of likes for this post using DatabaseManager
func (post *Post) GetLikesCountWithDB(dm *DatabaseManager) (int, error) {
	return dm.GetPostLikesCount(post.Id)
}

// get the number of dislikes for this post using DatabaseManager
func (post *Post) GetDislikesCountWithDB(dm *DatabaseManager) (int, error) {
	return dm.GetPostDislikesCount(post.Id)
}
