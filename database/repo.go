package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type GenericRepo[T any] struct {
	db       *sql.DB
	table    string
	fields   []string
	goFields []string
	idField  string
}

type Repo[T any] interface {
	Create(entity *T, db *sql.DB) (T, error)
	Update(entity *T, db *sql.DB) (T, error)
	FindById(id uuid.UUID, db *sql.DB) (T, error)
	FindAll(db *sql.DB) ([]T, error)
	Delete(id uuid.UUID, db *sql.DB) error
}

func NewGenericRepo[T any](db *sql.DB, tableName string) *GenericRepo[T] {
	var entity T
	t := reflect.TypeOf(entity)
	fields := make([]string, 0)
	goFields := make([]string, 0)

	var idField string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("db"); tag != "" {
			fields = append(fields, tag)

			if field.Tag.Get("primary") == "true" {
				idField = tag
			}
		}
		goFields = append(goFields, field.Name)
	}
	return &GenericRepo[T]{
		db:       db,
		table:    tableName,
		fields:   fields,
		goFields: goFields,
		idField:  idField,
	}
}

func (r *GenericRepo[T]) Create(entity *T, db *sql.DB) (T, error) {
	placeholders := make([]string, len(r.fields))
	values := make([]interface{}, len(r.fields))

	v := reflect.ValueOf(entity).Elem()

	for i, _ := range r.fields {

		placeholders[i] = fmt.Sprintf("$%d", i+1)

		value := v.FieldByName(r.goFields[i]).Interface()
		if value == nil {
			values[i] = nil
		} else {
			values[i] = value
		}
	}

	fmt.Println("placeholders:", placeholders)
	fmt.Println("values:", values)
	fmt.Println("fields", r.fields)

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		r.table,
		strings.Join(r.fields, ", "),
		strings.Join(placeholders, ", "),
	)

	fmt.Println(query)

	_, err := db.Exec(query, values...)
	if err != nil {
		return *entity, err
	}
	return *entity, nil
}
