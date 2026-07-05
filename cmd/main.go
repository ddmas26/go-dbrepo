package main

import (
	"context"
	"fmt"
	"time"

	"go_inventory/database"

	"github.com/google/uuid"
)

func testconnection() {
	repo, err := database.Connect2()
	if err != nil {
		panic(err)
	}
	database.QueryData(repo)

	repo.Close(context.Background())
}

func main() {
	db, err := database.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	user_repo := database.NewGenericRepo[database.User](db, "users")
	inventory_repo := database.NewGenericRepo[database.Inventory](db, "inventory")

	newUser := database.User{
		ID:        uuid.New(),
		Name:      "John Doe test from code 2",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	createdUser, err := user_repo.Create(&newUser, db)
	if err != nil {
		panic(err)
	}

	newInventory := database.Inventory{
		ID:        uuid.New(),
		Name:      "Sample Inventory",
		Quantity:  10,
		Price:     19.99,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	createdInventory, err := inventory_repo.Create(&newInventory, db)
	if err != nil {
		panic(err)
	}

	fmt.Println("user: ", createdUser)
	fmt.Println("inventory: ", createdInventory)
}
