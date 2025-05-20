package domain

// Pagination represents the structure of the pagination response
type Pagination struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
}
