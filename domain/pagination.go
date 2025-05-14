package domain

type Pagination struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
}
