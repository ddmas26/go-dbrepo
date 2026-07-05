package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func buildConnectionString() string {
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")
	USER := os.Getenv("DB_USER")
	PASSWORD := os.Getenv("PASSWORD")
	DBNAME := os.Getenv("DBNAME")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		USER, PASSWORD, HOST, PORT, DBNAME,
	)

	return connString
}

func Connect() (*sql.DB, error) {
	connString := buildConnectionString()
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
}
