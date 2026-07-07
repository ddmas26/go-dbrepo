package main

import (
	"fmt"
	"go_inventory/database"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// user_repo := database.NewGenericRepo[database.User](db, "users")
	inventory_repo := database.NewGenericRepo[database.Inventory](db, "inventory")

	// newUser := database.User{
	// 	ID:        uuid.New(),
	// 	Name:      "John Doe test from code 2",
	// 	CreatedAt: time.Now().Format(time.RFC3339),
	// }

	// createdUser, err := user_repo.Create(&newUser, db)
	// if err != nil {
	// 	panic(err)
	// }

	// newInventory := database.Inventory{
	// 	ID:        uuid.New(),
	// 	Name:      "Sample Inventory",
	// 	Quantity:  10,
	// 	Price:     19.99,
	// 	CreatedAt: time.Now().Format(time.RFC3339),
	// }

	// createdInventory, err := inventory_repo.Create(&newInventory, db)

	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("user: ", createdUser)
	// fmt.Println("inventory: ", createdInventory)

	// inventoryList, err := inventory_repo.FindAll(db)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(inventoryList)

	req := database.PaginationRequest{
		// Filter: "quantity: 10",
		SearchBy:       []string{"name"},
		SearchValue:    "wireless",
		OrderBy:        "price",
		OrderDirection: "DESC",
		PageIndex:      1,
		PageSize:       10,
	}
	inventoryList2, err := inventory_repo.FindAllPaginated(req, db)
	if err != nil {
		panic(err)
	}
	fmt.Println(inventoryList2)
}
