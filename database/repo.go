package database

import (
	"database/sql"
	"fmt"
	"log"
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
	var activeFields []string
	var placeholders []string
	var values []interface{}

	v := reflect.ValueOf(entity).Elem()

	placeholderIndex := 1

	for i, fieldName := range r.fields {
		goFieldName := r.goFields[i]
		fieldValue := v.FieldByName(goFieldName)

		if !fieldValue.IsValid() {
			continue
		}

		value := fieldValue.Interface()

		if value == nil || value == "" {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
			continue
		}

		activeFields = append(activeFields, fieldName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderIndex))
		values = append(values, value)
		placeholderIndex++
	}

	log.Println("--------> placeholders:", placeholders)
	log.Println("--------> values:", values)
	log.Println("--------> fields", activeFields)
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		r.table,
		strings.Join(activeFields, ", "),
		strings.Join(placeholders, ", "),
	)

	log.Println("------> query:\n", query)

	_, err := db.Exec(query, values...)
	if err != nil {
		return *entity, err
	}
	return *entity, nil
}
