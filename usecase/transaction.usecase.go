package usecase

import (
	"context"
	"main/dto"
	"main/repository"
	"math"
)

type TransactionService interface {
	ListTransactions(ctx context.Context, userID int, req dto.TransactionListRequest) (*dto.TransactionListResponse, error)
}

type transactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) ListTransactions(ctx context.Context, userID int, req dto.TransactionListRequest) (*dto.TransactionListResponse, error) {
	transactions, totalItems, err := s.repo.ListTransactions(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(req.Limit)))

	return &dto.TransactionListResponse{
		Transactions: transactions,
		Pagination: dto.PaginationInfo{
			CurrentPage:  req.Page,
			TotalPages:   totalPages,
			TotalItems:   totalItems,
			ItemsPerPage: req.Limit,
		},
	}, nil
}
