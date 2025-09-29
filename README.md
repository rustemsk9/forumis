#Forum Talk - 2025 - rsatimov

## How to run

Starting options
`sh
    go run .
    OR
    go run . --migrate (to start empty database)
    `
You can also build the project using:

    ```sh
    go build . && ./forum
    ```

BTW, if you want to run this project in Docker directly, you need to link this container with MySQL container, that is to
say you have been pull MySQL image and run MySQL with docker before you run this.

```
$ docker build -t forum .
$ docker run --link mysql:mysql -p 8080:8080 forum
```

## Architecture Overview

This is a Go-based forum application with a traditional web server architecture:

- `data/` package contains database models and operations (Thread, Post, User, Session)
- Direct SQL database operations using `database/sql` with prepared statements
- Session-based authentication using HTTP cookies (`_cookie`)
- Like/dislike system for both threads and posts with separate tables

## Key observed

- SQL referenced ids will autoincrement, on created Thread user_id, and User id. (for example: user_id = 2, the empty user table will start from used_id in threads). Should be manipulated correctly.
- On SQL access from app, and browsing data (not structures) will sometimes cause server to interrupts or delays. (This can be the main reason db-shm and db-wal is created OR modern sqlite3 every access case scenario)
- SQL Triggers:
  A SQLite3 trigger is a database object that automatically executes a specified SQL statement when a particular event occurs on a table (such as INSERT, UPDATE, or DELETE). Triggers are useful for enforcing business logic or maintaining data integrity.

  Example: Automatically update a `last_modified` timestamp on a `posts` table whenever a post is updated.

```sql
CREATE TRIGGER update_post_timestamp
AFTER UPDATE ON posts
FOR EACH ROW
BEGIN
    UPDATE posts SET last_modified = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
```

    This trigger runs after any update to the `posts` table and sets the `last_modified` column to the current timestamp for the affected row.

-

### Database Operations

- All database operations use prepared statements for security
- Error handling follows pattern: log error, return early with empty/nil values
- Query results use `defer rows.Close()` consistently
- Database connection accessed via `Db` in data package

Example pattern:

```go
rows, err := Db.Query("SELECT ... WHERE id=?", id)
if err != nil {
    fmt.Println("Error on select operation")
    return
}
defer rows.Close()
```
