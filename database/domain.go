package database

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID `db:"id" primary:"true"`
	Name      string    `db:"name"`
	CreatedAt string    `db:"created_at"`
}

type Inventory struct {
	ID        uuid.UUID `db:"id" primary:"true"`
	Name      string    `db:"name"`
	Quantity  int       `db:"quantity"`
	Price     float64   `db:"price"`
	CreatedAt string    `db:"created_at"`
	UpdatedAt string    `db:"updated_at"`
}
