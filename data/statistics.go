package data

// statistics DatabaseManager instance for statistics operations
var statsDM *DatabaseManager

// InitStatsDM initializes the global DatabaseManager for statistics operations
func InitStatsDM(dm *DatabaseManager) {
	statsDM = dm
}

func UserCount() (int, error) {
	return statsDM.GetUserCount()
}

func MostActiveUsers(limit int) (users []User, err error) {
	return statsDM.GetMostActiveUsers(limit)
}

func TotalPostsCount() (int, error) {
	return statsDM.GetTotalPostsCount()
}

func TotalThreadsCount() (int, error) {
	return statsDM.GetTotalThreadsCount()
}
