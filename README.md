#Forum Talk - 2025 - rsatimov
Origin remote git - https://github.com/rustemsk9/forumis
Commits - also https://github.com/rustemsk9/forumis

## Initial info: 
    go 1.24.6
    because docker uses 1.24.6 in bullseye:linux

## How to run

Starting options

```sh
    go run cmd/main.go
    OR
    go run cmd/main.go --migrate (to start empty database)
```

You can also build the project using:

```sh
    go build -o forum cmd/main.go && ./forum

    OR

    go build -o forum cmd/main.go && ./forum --migrate
```

## How to run in Docker

    Starting options
    We have special runserver.sh (in root folder)

```sh
    ./rundocker.sh
```

##
## Project Structure

```
forumis/
├── cmd/
│   └── main.go
├── internal/ - entity .go files
|   ├──────────── data
│   │             └─── database_*.go - includes all sqlite3 commands
│   ├── thread.go 
│   ├── post.go
│   ├── user.go
│   └── session.go
├── routes/ - mux handlers
├── utils/ - project utils
├── runserver.sh - docker image build, container deploy, clear on reset.
├── layout.css - for /templates
├── go.mod
├── go.sum
└── README.md
```
## Architecture Overview

This is a Go-based forum application with a traditional web server architecture:

- `internal/data` package contains database models and operations (Thread, Post, User, Session)
- Direct SQL database operations using `database/sql` with prepared statements
- Session-based authentication using HTTP cookies (`_cookie`) and (`sessions`)
- Like/dislike system for both threads and posts with separate tables
- Middleware: ex. baseChain := Chain() , authChain := Chain()

## SQL Key observed

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

## Golang backedn key observed

- We are not allowed to use any frameworks, neither backend nor front-end, which means:
  a. cmd/main.go for will not consist of any database operations, unless .NewDatabaseManager called. Which will start daemon entity within entity folder, either /routes or /utils or any other folder will not have access to the database operations, which increases entity security.
  b. Init works once at the start, and will not affect server configuration, database management.
  c. Database Inits only data folder, for each .go file in it, with special function called "InitAllDatabaseManagers".
  d. Bootstrap was removed due to its style components for likely considered front-ent framework, which enhances overall view, so layout.css controls all styles within our application.

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
