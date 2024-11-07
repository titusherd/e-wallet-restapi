package dto

import "main/entity"

type TransactionListResponse struct {
	Transactions []entity.Transaction `json:"transactions"`
	Pagination   PaginationInfo       `json:"pagination"`
}
