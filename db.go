package main

import (
	"database/sql"
	_ "embed"

	_ "github.com/glebarez/go-sqlite"
)

var (
	db *sql.DB
)

//go:embed migrations/new.sql
var createDBQuery string

type Post struct {
	ID      int
	Title   string
	Content string
}

func openDB(path string) {
	// connect
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		Fail("Failed to open SQLite DB! %s: %v", path, err)
	}

	// create db tables
	_, err = db.Exec(createDBQuery)
	if err != nil {
		Fail("Failed to create DB tables! %v", err)
	}
}

// calls transaction, if transaction returns a non-nil error the transaction is rolled back. otherwise the transaction is committed
func Transaction(transaction func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			// we panic'd ??? rollback and rethrow
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = transaction(tx)
	return
}

// ==================== Posts ====================

func addPost(title, content string) {
	_, err := db.Exec("INSERT INTO Posts (title, content) VALUES (?, ?)", title, content)
	if err != nil {
		Fail("Failed to add post! %v", err)
	}
}

// get the most recent <limit> posts
func getPosts(limit int) []Post {
	rows, err := db.Query("SELECT id, title, content FROM Posts ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		Fail("Failed to get posts! %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content)
		if err != nil {
			Fail("Failed to scan post! %v", err)
		}
		posts = append(posts, post)
	}
	return posts
}
