package main

import (
	"database/sql"
	"fmt"
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
		fmt.Println("Error:", err)
		fmt.Println("Rolling back migration...")
		downfilePath := "./migrate/migrations/" + folder.Name() + "/" + "down.sql"
		downQuery, readErr := os.ReadFile(downfilePath)
		if readErr != nil {
			fmt.Println("Error reading down.sql:", readErr)
			os.Exit(1)
		}

		_, execErr := DB.Exec(string(downQuery))
		if execErr != nil {
			fmt.Println("Error executing down.sql:", execErr)
			os.Exit(1)
		}
		fmt.Printf("Rollback successful in folder: %s", folder.Name())
		os.Exit(1)
	}
}

func checkFiles(folder os.DirEntry) {
	files, err := os.ReadDir("./migrate/migrations/" + folder.Name())
	check(err)
	upfile := files[1]
	filePath := "./migrate/migrations/" + folder.Name() + "/" + upfile.Name()
	query, err := os.ReadFile(filePath)
	check(err)
	fmt.Println("Executing query:")
	fmt.Println(string(query))
	_, err = DB.Exec(string(query))
	if err != nil {
		checkWithRollBack(err, folder)
	}
	fmt.Printf("------> Executed up.sql successfully in folder: %s\n", folder.Name())

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

func main() {
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")
	USER := os.Getenv("DB_USER")
	PASSWORD := os.Getenv("PASSWORD")
	DBNAME := os.Getenv("DBNAME")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		USER, PASSWORD, HOST, PORT, DBNAME,
	)

	fmt.Println("Connection string:", connString)
	var err error

	DB, err = sql.Open("postgres", connString)

	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer DB.Close()

	err = DB.Ping()
	if err != nil {
		fmt.Println("Error pinging database:", err)
		return
	}

	fmt.Println("Connection succeeded")

	checkfolders()
}
