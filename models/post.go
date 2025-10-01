package models

func (post *Post) CreatedAtDate() string {
	return post.CreatedAt.Format("Jan/2/2006 3:04pm")
}
