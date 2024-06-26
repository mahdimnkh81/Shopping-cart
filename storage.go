package main

import (
	"database/sql"
	_ "database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func OpenDataBase() error {
	var err error
	connStr := "user=postgres dbname=basket password=20667 sslmode=disable"
	DB, err = sql.Open("postgres", connStr)

	if err != nil {
		return err
	}

	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        username TEXT PRIMARY KEY,
        password TEXT NOT NULL
    );`

	_, err = DB.Exec(createUsersTable)
	if err != nil {
		fmt.Printf("Error creating users table: %s\n\n", err)
		return err
	}

	createBasketsTable := `
    CREATE TABLE IF NOT EXISTS baskets (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        data JSONB NOT NULL,
        state INTEGER NOT NULL,
		FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
    );`

	_, err = DB.Exec(createBasketsTable)
	if err != nil {
		fmt.Printf("Error creating baskets table: %s\n", err)
		return err
	}

	return nil
}
func CloseDatabase() error {
	return DB.Close()
}
