package repository

import (
	"context"
	"database/sql"
	"errors"
	"main/entity"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateResetPasswordCode(ctx context.Context, email, code string) error
	UpdatePassword(ctx context.Context, email, passwordHash string) error
}

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(ctx context.Context, user *entity.User) error {
	query := `
        INSERT INTO users (username, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}

	// Create wallet and game attempts for new user
	walletQuery := `
        INSERT INTO wallets (wallet_number, user_id, balance)
        VALUES (generate_wallet_number(), $1, 0)`

	gameQuery := `
        INSERT INTO game_attempts (user_id, attempts)
        VALUES ($1, 0)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, walletQuery, user.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, gameQuery, user.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *userRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	query := `
        SELECT id, username, email, password_hash, 
               reset_password_code, reset_password_code_expiry,
               created_at, updated_at
        FROM users
        WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.ResetPasswordCode,
		&user.ResetPasswordExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepositoryImpl) UpdateResetPasswordCode(ctx context.Context, email, code string) error {
	query := `
        UPDATE users 
        SET reset_password_code = $1,
            reset_password_code_expiry = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE email = $3`

	expiry := time.Now().Add(15 * time.Minute)
	result, err := r.db.ExecContext(ctx, query, code, expiry, email)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	query := `
        UPDATE users 
        SET password_hash = $1,
            reset_password_code = NULL,
            reset_password_code_expiry = NULL,
            updated_at = CURRENT_TIMESTAMP
        WHERE email = $2`

	result, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}
