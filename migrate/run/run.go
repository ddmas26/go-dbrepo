package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkWithRollBack(err error, folder os.DirEntry) {
	if err != nil {
		log.Println("Error:", err)
		log.Println("Rolling back migration...")
		downfilePath := "./migrate/migrations/" + folder.Name() + "/" + "down.sql"
		downQuery, readErr := os.ReadFile(downfilePath)
		if readErr != nil {
			log.Println("Error reading down.sql:", readErr)
			os.Exit(1)
		}

		_, execErr := DB.Exec(string(downQuery))
		if execErr != nil {
			log.Println("Error executing down.sql:", execErr)
			os.Exit(1)
		}
		fmt.Printf("Rollback successful in folder: %s", folder.Name())
		os.Exit(1)
	}
}

func checkFiles(folder os.DirEntry) {
	files, err := os.ReadDir("./migrate/migrations/" + folder.Name())
	if len(files) == 0 {
		return
	}
	check(err)
	upfile := files[1]
	filePath := "./migrate/migrations/" + folder.Name() + "/" + upfile.Name()
	query, err := os.ReadFile(filePath)
	check(err)
	log.Println("Executing query:")
	log.Println(string(query))
	_, err = DB.Exec(string(query))
	if err != nil {
		checkWithRollBack(err, folder)
	}
	log.Printf("------> Executed up.sql successfully in folder: %s\n", folder.Name())

}

func checkfolders() {
	dir := "./migrate/migrations/"
	folders, err := os.ReadDir(dir)
	check(err)

	for _, folder := range folders {
		if folder.IsDir() {
			checkFiles(folder)
		}
	}
}

func createDatabaseIfNotExists(host, port, user, password, dbname string) {
	// Connect to the default 'postgres' database to create the target database
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		user, password, host, port,
	)

	tmpDB, err := sql.Open("postgres", connString)
	if err != nil {
		log.Printf("---> Warning: could not connect to create database: %v", err)
		return
	}
	defer tmpDB.Close()

	// Check if database already exists
	var exists bool
	err = tmpDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		log.Printf("---> Warning: could not check if database exists: %v", err)
		return
	}

	if exists {
		log.Printf("---> Database '%s' already exists", dbname)
		return
	}

	// CREATE DATABASE cannot run inside a transaction
	_, err = tmpDB.Exec("CREATE DATABASE " + dbname)
	if err != nil {
		log.Printf("---> Warning: could not create database '%s': %v", dbname, err)
		return
	}

	log.Printf("---> Created database '%s'", dbname)
}

func main() {
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")
	USER := os.Getenv("DB_USER")
	PASSWORD := os.Getenv("PASSWORD")
	DBNAME := os.Getenv("DBNAME")

	createDatabaseIfNotExists(HOST, PORT, USER, PASSWORD, DBNAME)

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		USER, PASSWORD, HOST, PORT, DBNAME,
	)

	log.Println("---> Connection string:", connString)
	var err error

	DB, err = sql.Open("postgres", connString)

	if err != nil {
		log.Println("---> Error connecting to the database:", err)
		return
	}
	defer DB.Close()

	err = DB.Ping()
	if err != nil {
		log.Println("---> Error pinging database:", err)
		return
	}

	log.Println("---> Connection succeeded")

	checkfolders()
}
