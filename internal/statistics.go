package internal

import (
	"forum/internal/data"
	"forum/models"
)

// statistics DatabaseManager instance for statistics operations
var statsDM *data.DatabaseManager

// InitStatsDM initializes the global DatabaseManager for statistics operations
func InitStatsDM(dm *data.DatabaseManager) {
	statsDM = dm
}

func UserCount() (int, error) {
	return statsDM.GetUserCount()
}

func MostActiveUsers(limit int) (users []models.User, err error) {
	return statsDM.GetMostActiveUsers(limit)
}

func TotalPostsCount() (int, error) {
	return statsDM.GetTotalPostsCount()
}

func TotalThreadsCount() (int, error) {
	return statsDM.GetTotalThreadsCount()
}
