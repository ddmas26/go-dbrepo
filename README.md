# inventory

# go-dbrepo

A generic, type-safe CRUD and paginated query library for PostgreSQL, built with Go generics and reflection.

## Features

- **Generic Repository** — type-safe CRUD operations for any struct using Go 1.26 generics
- **Safe Paginated Queries** — SQL injection protected via parameterized queries and whitelist validation
- **Dynamic Filtering** — filter with `AND`/`OR` combinations, comparison operators (`=`, `>`, `<`, `>=`, `<=`)
- **Full-Text Search** — `ILIKE` search across specified columns with whitelist-validated column names
- **Ordered Pagination** — safe `ORDER BY` with whitelist-validated columns and directions
- **Count Query** — returns total matching rows for frontend pagination controls
- **Auto-Migration** — runs SQL migration files in order from `migrate/migrations/` with automatic rollback
- **Integration Tests** — 17 tests covering CRUD, pagination, edge cases, and SQL injection scenarios

## Project Structure

```
├── cmd/
│   └── main.go                    # Entry point
├── database/
│   ├── connect.go                 # DB connection (Postgres)
│   ├── domain.go                  # Domain models (User, Inventory)
│   ├── dto.go                     # Pagination request/response DTOs
│   └── repo.go                    # Generic repository implementation
├── migrate/
│   ├── migrations/                # SQL migration files (numbered folders)
│   │   ├── 0_01072026/           # Initial users table
│   │   ├── 1_04072026/           # Inventory table
│   │   ├── 2_05072026/           # Add created_at/updated_at to users
│   │   └── 3_06072026/           # (placeholder)
│   └── run/
│       └── run.go                 # Migration runner
├── tests/
│   └── repo_test.go              # Integration tests (17 test functions)
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL (or Docker — recommended for local development)

### Quick Start (Docker — Recommended)

Run PostgreSQL in a container:

```bash
docker run -d \
  --name inventory-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=inventory_db \
  -p 5432:5432 \
  postgres
```

To stop and remove later:

```bash
docker stop inventory-db
docker rm inventory-db
```

### Environment Variables

Create a `.env` file in the project root (matching the Docker setup above):

```env
HOST=localhost
PORT=5432
DB_USER=postgres
PASSWORD=your_password
DBNAME=inventory_db
```

### Clone and Run (Full Steps)

```bash
# 1. Clone the repository
git clone <repo-url>
cd go-dbrepo

# 2. Start PostgreSQL (Docker)
docker run -d \
  --name inventory-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=inventory_db \
  -p 5432:5432 \
  postgres

# 3. Run migrations (creates tables)
go run ./migrate/run/

# 4. Run the app
go run ./cmd/
```

## Usage

### Define a Model

Tag struct fields with `db` and optionally `primary`:

```go
type Inventory struct {
    ID        uuid.UUID  `db:"id" primary:"true"`
    Name      string     `db:"name"`
    Quantity  int        `db:"quantity"`
    Price     float64    `db:"price"`
    CreatedAt time.Time  `db:"created_at"`
    UpdatedAt *time.Time `db:"updated_at"`
}
```

### Create a Repository

```go
repo := database.NewGenericRepo[database.Inventory](db, "inventory")
```

### CRUD Operations

```go
// Create
created, err := repo.Create(&entity, db)

// Find All
all, err := repo.FindAll(db)

// Find By ID — returns error if not found
item, err := repo.FindById(id, db)

// Update — returns error if ID doesn't exist
updated, err := repo.Update(&entity, db)

// Delete — returns error if ID doesn't exist
err := repo.Delete(id, db)
```

### Paginated Queries

```go
req := database.PaginationRequest{
    PageIndex:      1,
    PageSize:       10,
    Filter:         "price: > 10, quantity: < 100|quantity: > 500",
    SearchBy:       []string{"name"},
    SearchValue:    "widget",
    OrderBy:        "price",
    OrderDirection: "DESC",
}

result, err := repo.FindAllPaginated(req, db)
// result.Data        — items for current page
// result.TotalCount  — total matching rows
// result.PageIndex   — current page
// result.PageSize    — items per page
// result.TotalPages  — total number of pages
```

## Filter Syntax

The filter string supports `AND` (comma) and `OR` (pipe) grouping:

```
field:value                      → field = value
field: > value                   → field > value
field: < value                   → field < value
field: >= value                  → field >= value
field: <= value                  → field <= value
field: val1|field: val2          → (field = val1 OR field = val2)  [OR group]
field: val1 , field2: val2       → field = val1 AND field2 = val2  [AND groups]
```

All field names are validated against the model's `db` tags — SQL injection via column names is prevented.

## Response Structure

`FindAllPaginated` returns a `PaginationResponse[T]`:

```go
type PaginationResponse[T any] struct {
    Data       []T   // items for the current page
    TotalCount int   // total matching rows (across all pages)
    PageIndex  int   // current page number
    PageSize   int   // items per page
    TotalPages int   // total number of pages
}
```

## Running Tests

Tests are in the `tests/` folder and require a running PostgreSQL instance.

```bash
# Set environment variables (or use .env file in project root)
export HOST=localhost PORT=5432 DB_USER=postgres PASSWORD=pass DBNAME=test_db

# Run all tests
go test ./tests/ -v -count=1

# Run a specific test
go test ./tests/ -v -run TestFindAllPaginated_WithFilter -count=1
```

Tests will **skip gracefully** if no database is available.

## Security

- **Parameterized queries** — all values are passed via `$N` placeholders
- **Whitelist validation** — column names for `OrderBy`, `Filter`, and `SearchBy` are validated against known struct fields
- **`RowsAffected` checks** — `Update` and `Delete` verify the operation actually touched a row
- **No raw string interpolation** — user input never appears directly in SQL

## Migrations

The migration runner:

1. **Creates the database if it doesn't exist** — connects to the default `postgres` database and runs `CREATE DATABASE` automatically
2. **Connects to your target database**
3. **Executes `up.sql`** files in each numbered folder sequentially
4. **Rolls back on failure** — if any migration fails, the corresponding `down.sql` is executed

```bash
go run ./migrate/run/
```

### Migration Folders

Each folder under `migrate/migrations/` represents one migration step:

```
migrate/migrations/
├── 0_01072026/           # Migration 0 — initial tables
│   ├── up.sql            # CREATE TABLE users
│   └── down.sql          # DROP TABLE users
├── 1_04072026/           # Migration 1
│   ├── up.sql            # CREATE TABLE inventory
│   └── down.sql
├── 2_05072026/           # Migration 2
│   ├── up.sql            # ALTER TABLE users ADD COLUMN
│   └── down.sql
└── 3_06072026/           # Placeholder (empty)
```

**To add a new migration:**
1. Create a folder like `4_08072026/`
2. Add `up.sql` with your changes
3. Add `down.sql` with the reverse (for rollback)

## Future Improvements

- **JOIN support** — The count query currently uses `strings.Replace(query, "*", "COUNT(*)", 1)`, which works for single-table queries. When JOINs are added, it should be refactored to build the count query independently:

```go
countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table)
if len(whereClauses) > 0 {
    countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
}
countArgs := append([]interface{}{}, args...)
```
