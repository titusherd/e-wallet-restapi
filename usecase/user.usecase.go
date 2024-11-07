package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"main/dto"
	"main/entity"
	"main/repository"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*entity.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (string, error)
	ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) (string, error)
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type service struct {
	repo        repository.UserRepository
	jwtSecret   []byte
	jwtIssuer   string
	jwtDuration time.Duration
}

func NewService(repo repository.UserRepository, jwtSecret string, jwtIssuer string, jwtDuration time.Duration) Service {
	return &service{
		repo:        repo,
		jwtSecret:   []byte(jwtSecret),
		jwtIssuer:   jwtIssuer,
		jwtDuration: jwtDuration,
	}
}

func (s *service) Register(ctx context.Context, req dto.RegisterRequest) (*entity.User, error) {
	// Check if user exists
	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.jwtDuration).Unix(),
		"iss": s.jwtIssuer,
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *service) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	fmt.Println(user)
	if err != nil {
		return "", err
	}

	// Generate reset code
	code := make([]byte, 32)
	_, err = rand.Read(code)
	if err != nil {
		return "", err
	}
	resetCode := base64.URLEncoding.EncodeToString(code)

	err = s.repo.UpdateResetPasswordCode(ctx, req.Email, resetCode)
	if err != nil {
		return "", err
	}

	return resetCode, nil
}

func (s *service) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	if user.ResetPasswordCode == nil || *user.ResetPasswordCode != req.ResetCode {
		return errors.New("invalid reset code")
	}

	if user.ResetPasswordExpiry == nil || user.ResetPasswordExpiry.Before(time.Now()) {
		return errors.New("reset code expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, req.Email, string(hashedPassword))
}

func (s *service) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
}
