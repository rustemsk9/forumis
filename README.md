# Forum Application - AI Coding Instructions

## Architecture Overview
This is a Go-based forum application with a traditional web server architecture:
- `data/` package contains database models and operations (Thread, Post, User, Session)
- Direct SQL database operations using `database/sql` with prepared statements
- Session-based authentication using HTTP cookies (`_cookie`)
- Like/dislike system for both threads and posts with separate tables

## Key Patterns & Conventions

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

## How to run

Clone this project under `$GOPATH/src`, then create a database (it's ok whatever you want to named) and create some tables 
by following [`data/setup.sql`](./data/setup.sql). After that, configure your MySql username, password and database name 
in [`data/sql.json`](./data/sql.json).

A MySql driver is required

```
$ go get github.com/go-sql-driver/mysql
```

then you can run it
```
$ go run main.go 
```

BTW, if you want to run this project in Docker directly, you need to link this container with MySQL container, that is to 
say you have been pull MySQL image and run MySQL with docker before you run this.

```
$ docker build -t forum .
$ docker run --link mysql:mysql -p 8080:8080 forum
```
