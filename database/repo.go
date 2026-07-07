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
	FindAllPaginated(req PaginationRequest, db *sql.DB) ([]T, error)
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

func (r *GenericRepo[T]) FindAll(db *sql.DB) ([]T, error) {
	query := fmt.Sprintf("SELECT * FROM %s", r.table)

	log.Println("query: ", query)

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []T

	for rows.Next() {
		var entity T

		v := reflect.ValueOf(&entity).Elem()

		values := make([]interface{}, len(r.fields))
		for i := range r.fields {
			values[i] = v.Field(i).Addr().Interface()
		}

		err := rows.Scan(values...)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		results = append(results, entity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return results, nil

}

func (r *GenericRepo[T]) FindAllPaginated(req PaginationRequest, db *sql.DB) (PaginationResponse[T], error) {
	var whereClauses []string
	var args []interface{}

	if req.Filter != "" {
		filterClauses, filterArgs := buildFilter(req.Filter, r.fields, len(args)+1)
		if len(filterClauses) > 0 {
			whereClauses = append(whereClauses, filterClauses...)
			args = append(args, filterArgs...)
		}
	}

	if req.SearchValue != "" && len(req.SearchBy) != 0 {
		var validSearchBy []string
		for _, sb := range req.SearchBy {
			for _, field := range r.fields {
				if strings.EqualFold(field, sb) {
					validSearchBy = append(validSearchBy, field)
					break
				}
			}
		}

		if len(validSearchBy) > 0 {
			searchClause, searchArg := buildSearch(req.SearchValue, validSearchBy, len(args)+1)
			if searchClause != "" {
				whereClauses = append(whereClauses, searchClause)
				args = append(args, searchArg)
			}
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", r.table)
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	orderBy := r.idField
	if req.OrderBy != "" {
		for _, field := range r.fields {
			if strings.EqualFold(field, req.OrderBy) {
				orderBy = field
				break
			}
		}
	}

	countQuery := strings.Replace(query, "*", "COUNT(*)", 1)
	countArgs := append([]interface{}{}, args...)

	query += fmt.Sprintf(" ORDER BY %s", orderBy)

	orderDir := "ASC"
	switch strings.ToUpper(req.OrderDirection) {
	case "ASC", "DESC":
		orderDir = strings.ToUpper(req.OrderDirection)
	}
	query += " " + orderDir

	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	offset := (req.PageIndex - 1) * req.PageSize

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	args = append(args, req.PageSize, offset)

	log.Println("Final Query:", query)
	log.Println("Count Query:", countQuery)
	log.Println("Arguments:  ", args)

	rows, err := db.Query(query, args...)
	if err != nil {
		return PaginationResponse[T]{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var count int
	countErr := db.QueryRow(countQuery, countArgs...).Scan(&count)

	if countErr != nil {
		return PaginationResponse[T]{}, fmt.Errorf("count query failed: %w", countErr)
	}

	log.Println("Counter now is: ", count)

	var results []T
	for rows.Next() {
		var entity T
		v := reflect.ValueOf(&entity).Elem()
		values := make([]interface{}, len(r.fields))
		for i := range r.fields {
			values[i] = v.Field(i).Addr().Interface()
		}
		if err := rows.Scan(values...); err != nil {
			return PaginationResponse[T]{}, fmt.Errorf("scan failed: %w", err)
		}
		results = append(results, entity)
	}
	if err = rows.Err(); err != nil {
		return PaginationResponse[T]{}, fmt.Errorf("rows iteration failed: %w", err)
	}

	totalPages := (count + req.PageSize - 1) / req.PageSize

	return PaginationResponse[T]{
		Data:       results,
		TotalCount: count,
		PageIndex:  req.PageIndex,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func buildFilter(filter string, fields []string, startingPlaceholderIndex int) ([]string, []interface{}) {
	if len(strings.TrimSpace(filter)) == 0 {
		return nil, nil
	}

	var andClauses []string
	var args []interface{}
	placeholderIndex := startingPlaceholderIndex

	andGroups := strings.Split(filter, ",")

	for _, group := range andGroups {
		orPairs := strings.Split(group, "|")
		var orClauses []string

		for _, pair := range orPairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				continue
			}

			filterKey := strings.TrimSpace(parts[0])
			filterVal := strings.TrimSpace(parts[1])

			operator := "="
			cleanVal := filterVal

			if strings.HasPrefix(filterVal, ">=") {
				operator = ">="
				cleanVal = strings.TrimPrefix(filterVal, ">=")
			} else if strings.HasPrefix(filterVal, "<=") {
				operator = "<="
				cleanVal = strings.TrimPrefix(filterVal, "<=")
			} else if strings.HasPrefix(filterVal, ">") {
				operator = ">"
				cleanVal = strings.TrimPrefix(filterVal, ">")
			} else if strings.HasPrefix(filterVal, "<") {
				operator = "<"
				cleanVal = strings.TrimPrefix(filterVal, "<")
			}

			cleanVal = strings.TrimSpace(cleanVal)
			if cleanVal == "" {
				continue
			}

			for _, dbField := range fields {
				if strings.EqualFold(dbField, filterKey) {
					orClauses = append(orClauses, fmt.Sprintf("%s %s $%d", dbField, operator, placeholderIndex))
					args = append(args, cleanVal)
					placeholderIndex++
					break
				}
			}
		}

		if len(orClauses) > 0 {
			if len(orClauses) == 1 {
				andClauses = append(andClauses, orClauses[0])
			} else {
				andClauses = append(andClauses, "("+strings.Join(orClauses, " OR ")+")")
			}
		}
	}

	return andClauses, args
}

func buildSearch(searchValue string, searchableFields []string, placeholderIndex int) (string, interface{}) {
	cleanSearch := strings.TrimSpace(searchValue)
	if cleanSearch == "" || len(searchableFields) == 0 {
		return "", nil
	}

	var matchClauses []string

	for _, dbField := range searchableFields {
		matchClauses = append(matchClauses, fmt.Sprintf("%s ILIKE $%d", dbField, placeholderIndex))
	}
	searchClause := "(" + strings.Join(matchClauses, " OR ") + ")"
	wrappedValue := "%" + cleanSearch + "%"

	return searchClause, wrappedValue
}
