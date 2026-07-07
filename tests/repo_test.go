package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"go_inventory/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

type TestItem struct {
	ID        uuid.UUID  `db:"id" primary:"true"`
	Name      string     `db:"name"`
	Quantity  int        `db:"quantity"`
	Price     float64    `db:"price"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

func connectTestDB(t *testing.T) *sql.DB {
	t.Helper()
	godotenv.Load("../.env")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("PASSWORD")
	dbname := os.Getenv("DBNAME")

	log.Printf("host: %s", host)

	if host == "" || port == "" || user == "" || dbname == "" {
		t.Skip("Skipping: database environment variables not set (HOST, PORT, DB_USER, PASSWORD, DBNAME)")
	}

	connStr := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("Skipping: database not reachable (%v)", err)
	}

	return db
}

func setupTestTable(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS test_items (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL DEFAULT 0,
			price DECIMAL(10,2) NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
}

func cleanTestTable(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("DELETE FROM test_items")
	if err != nil {
		t.Fatalf("Failed to clean test table: %v", err)
	}
}

func TestCreate(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "Test Widget",
		Quantity:  10,
		Price:     19.99,
		CreatedAt: time.Now(),
	}

	created, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID != item.ID {
		t.Errorf("expected ID %v, got %v", item.ID, created.ID)
	}
	if created.Name != "Test Widget" {
		t.Errorf("expected Name 'Test Widget', got '%s'", created.Name)
	}
	if created.Quantity != 10 {
		t.Errorf("expected Quantity 10, got %d", created.Quantity)
	}
	if created.Price != 19.99 {
		t.Errorf("expected Price 19.99, got %f", created.Price)
	}
}

func TestFindById(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	id := uuid.New()
	item := TestItem{
		ID:        id,
		Name:      "Find Me",
		Quantity:  5,
		Price:     9.99,
		CreatedAt: time.Now(),
	}

	_, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindById(id, db)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	if found.Name != "Find Me" {
		t.Errorf("expected Name 'Find Me', got '%s'", found.Name)
	}
	if found.Quantity != 5 {
		t.Errorf("expected Quantity 5, got %d", found.Quantity)
	}
}

func TestFindById_NotFound(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	_, err := repo.FindById(uuid.New(), db)
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestFindAll(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	items := []TestItem{
		{ID: uuid.New(), Name: "Item A", Quantity: 10, Price: 1.99, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Item B", Quantity: 20, Price: 2.99, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Item C", Quantity: 30, Price: 3.99, CreatedAt: time.Now()},
	}

	for _, item := range items {
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	all, err := repo.FindAll(db)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("expected 3 items, got %d", len(all))
	}
}

func TestUpdate(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "Original Name",
		Quantity:  10,
		Price:     19.99,
		CreatedAt: time.Now(),
	}

	_, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	item.Name = "Updated Name"
	item.Price = 29.99

	updated, err := repo.Update(&item, db)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.Price != 29.99 {
		t.Errorf("expected Price 29.99, got %f", updated.Price)
	}

	found, _ := repo.FindById(item.ID, db)
	if found.Name != "Updated Name" {
		t.Errorf("DB still has old name: '%s'", found.Name)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "Ghost",
		Quantity:  1,
		Price:     1.00,
		CreatedAt: time.Now(),
	}

	_, err := repo.Update(&item, db)
	if err == nil {
		t.Fatal("expected error for updating non-existent item, got nil")
	}
}

func TestDelete(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "To Delete",
		Quantity:  1,
		Price:     1.00,
		CreatedAt: time.Now(),
	}

	_, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.Delete(item.ID, db)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.FindById(item.ID, db)
	if err == nil {
		t.Error("expected error finding deleted item, got nil")
	}
}

func TestDelete_NotFound(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	err := repo.Delete(uuid.New(), db)
	if err == nil {
		t.Fatal("expected error for deleting non-existent item, got nil")
	}
}

func TestFindAllPaginated_NoFilter(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	for i := 0; i < 25; i++ {
		item := TestItem{
			ID:        uuid.New(),
			Name:      "Item",
			Quantity:  i + 1,
			Price:     float64(i+1) * 10,
			CreatedAt: time.Now(),
		}
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	req := database.PaginationRequest{
		PageIndex:      1,
		PageSize:       10,
		OrderBy:        "quantity",
		OrderDirection: "ASC",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if len(result.Data) != 10 {
		t.Errorf("expected 10 items on page, got %d", len(result.Data))
	}
	if result.TotalCount != 25 {
		t.Errorf("expected TotalCount 25, got %d", result.TotalCount)
	}
	if result.TotalPages != 3 {
		t.Errorf("expected TotalPages 3, got %d", result.TotalPages)
	}
	if result.PageIndex != 1 {
		t.Errorf("expected PageIndex 1, got %d", result.PageIndex)
	}
	if result.PageSize != 10 {
		t.Errorf("expected PageSize 10, got %d", result.PageSize)
	}
}

func TestFindAllPaginated_SecondPage(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	for i := 0; i < 25; i++ {
		item := TestItem{
			ID:        uuid.New(),
			Name:      "Item",
			Quantity:  i + 1,
			Price:     float64(i+1) * 10,
			CreatedAt: time.Now(),
		}
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	req := database.PaginationRequest{
		PageIndex:      2,
		PageSize:       10,
		OrderBy:        "quantity",
		OrderDirection: "ASC",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if len(result.Data) != 10 {
		t.Errorf("expected 10 items on page 2, got %d", len(result.Data))
	}
	if result.PageIndex != 2 {
		t.Errorf("expected PageIndex 2, got %d", result.PageIndex)
	}

	// First item on page 2 should have quantity 11 (since page 1 had 1-10)
	if result.Data[0].Quantity != 11 {
		t.Errorf("expected first item quantity 11, got %d", result.Data[0].Quantity)
	}
}

func TestFindAllPaginated_WithFilter(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	for i := 1; i <= 10; i++ {
		item := TestItem{
			ID:        uuid.New(),
			Name:      "Item",
			Quantity:  i * 10,
			Price:     float64(i) * 10,
			CreatedAt: time.Now(),
		}
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	req := database.PaginationRequest{
		PageIndex: 1,
		PageSize:  10,
		Filter:    "quantity: > 30",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if result.TotalCount != 7 {
		t.Errorf("expected TotalCount 7 (quantity > 30), got %d", result.TotalCount)
	}
}

func TestFindAllPaginated_WithSearch(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	items := []TestItem{
		{ID: uuid.New(), Name: "Red Widget", Quantity: 10, Price: 10.00, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Blue Widget", Quantity: 20, Price: 20.00, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Green Gadget", Quantity: 30, Price: 30.00, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Yellow Gadget", Quantity: 40, Price: 40.00, CreatedAt: time.Now()},
	}

	for _, item := range items {
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	req := database.PaginationRequest{
		PageIndex:   1,
		PageSize:    10,
		SearchBy:    []string{"name"},
		SearchValue: "Widget",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2 (search 'Widget'), got %d", result.TotalCount)
	}
}

func TestFindAllPaginated_WithFilterAndSearch(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	items := []TestItem{
		{ID: uuid.New(), Name: "Cheap Widget", Quantity: 10, Price: 5.00, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Expensive Widget", Quantity: 20, Price: 50.00, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Cheap Gadget", Quantity: 30, Price: 5.00, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Expensive Gadget", Quantity: 40, Price: 50.00, CreatedAt: time.Now()},
	}

	for _, item := range items {
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	req := database.PaginationRequest{
		PageIndex:   1,
		PageSize:    10,
		Filter:      "price: > 10",
		SearchBy:    []string{"name"},
		SearchValue: "Widget",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	// Only "Expensive Widget" matches (price > 10 AND name ILIKE '%Widget%')
	if result.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", result.TotalCount)
	}
	if len(result.Data) != 1 || result.Data[0].Name != "Expensive Widget" {
		t.Errorf("expected 'Expensive Widget', got %+v", result.Data)
	}
}

func TestFindAllPaginated_DefaultPagination(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	for i := 0; i < 5; i++ {
		item := TestItem{
			ID:        uuid.New(),
			Name:      "Item",
			Quantity:  i + 1,
			Price:     float64(i+1) * 10,
			CreatedAt: time.Now(),
		}
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// PageIndex = 0 and PageSize = 0 should use defaults (1, 10)
	req := database.PaginationRequest{
		PageIndex: 0,
		PageSize:  0,
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if result.PageIndex != 1 {
		t.Errorf("expected PageIndex default 1, got %d", result.PageIndex)
	}
	if result.PageSize != 10 {
		t.Errorf("expected PageSize default 10, got %d", result.PageSize)
	}
	if result.TotalCount != 5 {
		t.Errorf("expected TotalCount 5, got %d", result.TotalCount)
	}
	if len(result.Data) != 5 {
		t.Errorf("expected 5 items, got %d", len(result.Data))
	}
}

func TestFindAllPaginated_SearchByInjection(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "Safe Item",
		Quantity:  1,
		Price:     1.00,
		CreatedAt: time.Now(),
	}
	_, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Malicious SearchBy should be ignored by whitelist validation
	req := database.PaginationRequest{
		PageIndex:   1,
		PageSize:    10,
		SearchBy:    []string{"name; DROP TABLE test_items; --"},
		SearchValue: "test",
	}

	_, err = repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated should not error on invalid SearchBy: %v", err)
	}

	// Table should still exist — verify by finding all
	all, err := repo.FindAll(db)
	if err != nil {
		t.Fatalf("FindAll after injection attempt failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 item still in table, got %d", len(all))
	}
}

func TestFindAllPaginated_OrderByInjection(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "Safe Item",
		Quantity:  1,
		Price:     1.00,
		CreatedAt: time.Now(),
	}
	_, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Malicious OrderBy should fall back to idField
	req := database.PaginationRequest{
		PageIndex:      1,
		PageSize:       10,
		OrderBy:        "name; DROP TABLE test_items; --",
		OrderDirection: "ASC",
	}

	_, err = repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated should not error on invalid OrderBy: %v", err)
	}

	// Table should still exist
	all, err := repo.FindAll(db)
	if err != nil {
		t.Fatalf("FindAll after injection attempt failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 item still in table, got %d", len(all))
	}
}

func TestFindAllPaginated_FilterInjection(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	item := TestItem{
		ID:        uuid.New(),
		Name:      "Safe Item",
		Quantity:  1,
		Price:     1.00,
		CreatedAt: time.Now(),
	}
	_, err := repo.Create(&item, db)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Invalid filter key should be ignored by whitelist
	req := database.PaginationRequest{
		PageIndex: 1,
		PageSize:  10,
		Filter:    "nonexistent_field: > 10",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated should not error on invalid filter: %v", err)
	}

	// No filter applied — should return all items
	if result.TotalCount != 1 {
		t.Errorf("expected TotalCount 1 (filter ignored), got %d", result.TotalCount)
	}
}

func TestFindAllPaginated_EmptyResult(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	req := database.PaginationRequest{
		PageIndex: 1,
		PageSize:  10,
		Filter:    "quantity: > 99999",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected empty Data, got %d items", len(result.Data))
	}
	if result.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", result.TotalCount)
	}
	if result.TotalPages != 0 {
		t.Errorf("expected TotalPages 0, got %d", result.TotalPages)
	}
}

func TestFindAllPaginated_LastPagePartial(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()

	setupTestTable(t, db)
	defer cleanTestTable(t, db)

	repo := database.NewGenericRepo[TestItem](db, "test_items")

	for i := 0; i < 25; i++ {
		item := TestItem{
			ID:        uuid.New(),
			Name:      "Item",
			Quantity:  i + 1,
			Price:     float64(i+1) * 10,
			CreatedAt: time.Now(),
		}
		_, err := repo.Create(&item, db)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Page 3 should have only 5 items (25 total, 10 per page)
	req := database.PaginationRequest{
		PageIndex:      3,
		PageSize:       10,
		OrderBy:        "quantity",
		OrderDirection: "ASC",
	}

	result, err := repo.FindAllPaginated(req, db)
	if err != nil {
		t.Fatalf("FindAllPaginated failed: %v", err)
	}

	if len(result.Data) != 5 {
		t.Errorf("expected 5 items on last page, got %d", len(result.Data))
	}
	if result.TotalPages != 3 {
		t.Errorf("expected TotalPages 3, got %d", result.TotalPages)
	}
}
