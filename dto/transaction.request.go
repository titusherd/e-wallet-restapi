package dto

type TransactionListRequest struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
	Search    string `form:"s"`
	SortBy    string `form:"sortBy"`
	SortOrder string `form:"sort"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
}

type PaginationInfo struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalItems   int `json:"total_items"`
	ItemsPerPage int `json:"items_per_page"`
}
