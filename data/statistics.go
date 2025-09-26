package data

func UserCount() (int, error) {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func MostThreeActiveUsers(limit int) (users []User, err error) {
	rows, err := Db.Query(`
		SELECT u.id, u.uuid, u.name, u.email, u.password, u.created_at, COUNT(t.id) AS thread_count
		FROM users u
		LEFT JOIN threads t ON u.id = t.user_id
		GROUP BY u.id
		ORDER BY thread_count DESC
		LIMIT ?`, 3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var threadCount int
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &threadCount)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func TotalPostsCount() (int, error) {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func TotalThreadsCount() (int, error) {
	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM threads").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}