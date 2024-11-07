package repository

import (
	"context"
	"database/sql"
	"fmt"
	"main/dto"
	"main/entity"
	"strings"
	"time"
)

type TransactionRepository interface {
	ListTransactions(ctx context.Context, userID int, req dto.TransactionListRequest) ([]entity.Transaction, int, error)
}

type transactionRepoImpl struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepoImpl{db: db}
}

func (r *transactionRepoImpl) ListTransactions(ctx context.Context, userID int, req dto.TransactionListRequest) ([]entity.Transaction, int, error) {
	// Build the base query
	baseQuery := `
        SELECT 
            t.id, t.from_wallet_id, t.to_wallet_id, t.amount, 
            t.description, t.source_of_fund_id, t.transaction_type, 
            t.created_at,
            fw.wallet_number as from_wallet_number,
            tw.wallet_number as to_wallet_number,
            u.username as recipient_name
        FROM transactions t
        LEFT JOIN wallets fw ON t.from_wallet_id = fw.id
        JOIN wallets tw ON t.to_wallet_id = tw.id
        JOIN users u ON tw.user_id = u.id
        WHERE (fw.user_id = $1 OR tw.user_id = $1)
    `

	// Build the count query
	countQuery := `
        SELECT COUNT(*) 
        FROM transactions t
        LEFT JOIN wallets fw ON t.from_wallet_id = fw.id
        JOIN wallets tw ON t.to_wallet_id = tw.id
        WHERE (fw.user_id = $1 OR tw.user_id = $1)
    `

	params := []interface{}{userID}
	paramCount := 1

	// Add search condition
	if req.Search != "" {
		paramCount++
		baseQuery += fmt.Sprintf(" AND LOWER(t.description) LIKE LOWER($%d)", paramCount)
		countQuery += fmt.Sprintf(" AND LOWER(t.description) LIKE LOWER($%d)", paramCount)
		params = append(params, "%"+req.Search+"%")
	}

	// Add date range filter
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			paramCount++
			baseQuery += fmt.Sprintf(" AND t.created_at >= $%d", paramCount)
			countQuery += fmt.Sprintf(" AND t.created_at >= $%d", paramCount)
			params = append(params, startDate)
		}
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			endDate = endDate.Add(24 * time.Hour) // Include the entire end date
			paramCount++
			baseQuery += fmt.Sprintf(" AND t.created_at < $%d", paramCount)
			countQuery += fmt.Sprintf(" AND t.created_at < $%d", paramCount)
			params = append(params, endDate)
		}
	}

	// Add sorting
	switch req.SortBy {
	case "date":
		baseQuery += " ORDER BY t.created_at"
	case "amount":
		baseQuery += " ORDER BY t.amount"
	case "recipient":
		baseQuery += " ORDER BY recipient_name"
	default:
		baseQuery += " ORDER BY t.created_at" // Default sort by date
	}

	if strings.ToLower(req.SortOrder) == "asc" {
		baseQuery += " ASC"
	} else {
		baseQuery += " DESC" // Default to DESC
	}

	// Add pagination
	offset := (req.Page - 1) * req.Limit
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", req.Limit, offset)

	// Get total count
	var totalItems int
	err := r.db.QueryRowContext(ctx, countQuery, params...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	// Execute the main query
	rows, err := r.db.QueryContext(ctx, baseQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []entity.Transaction
	for rows.Next() {
		var t entity.Transaction
		var fromWalletID, fromWalletNumber sql.NullString
		err := rows.Scan(
			&t.ID, &fromWalletID, &t.ToWalletID, &t.Amount,
			&t.Description, &t.SourceOfFundID, &t.TransactionType,
			&t.CreatedAt, &fromWalletNumber, &t.ToWalletNumber,
			&t.RecipientName,
		)
		if err != nil {
			return nil, 0, err
		}
		if fromWalletNumber.Valid {
			t.FromWalletNumber = fromWalletNumber.String
		}
		transactions = append(transactions, t)
	}

	return transactions, totalItems, nil
}
