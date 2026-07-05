package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

func Connect2() (*pgx.Conn, error) {

	connString := os.Getenv("DB_CONNECTION")
	conn, err := pgx.Connect(context.Background(), connString)
	return conn, err
}

func QueryData(conn *pgx.Conn) {
	rows, err := conn.Query(context.Background(), "SELECT id, name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User ID: %s, Name: %s\n", id, name)
	}
}
