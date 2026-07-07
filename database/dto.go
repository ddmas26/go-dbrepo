package database

type PaginationRequest struct {
	PageIndex      int
	PageSize       int
	Filter         string
	SearchValue    string
	SearchBy       []string
	OrderBy        string
	OrderDirection string
}

type PaginationResponse[T any] struct {
	Data       []T
	TotalCount int
	PageIndex  int
	PageSize   int
	TotalPages int
}
