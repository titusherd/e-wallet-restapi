// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	auth "main/handler"
	"main/middleware"
	"main/repository"
	"main/usecase"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	ServerPort  string
	JWTSecret   string
	JWTIssuer   string
	JWTDuration time.Duration
}

func loadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	config := &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "Kamji^50"),
		DBName:      getEnv("DB_NAME", "e-wallet_db"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "=-0=-0"),
		JWTIssuer:   getEnv("JWT_ISSUER", "ewallet-api"),
		JWTDuration: 24 * time.Hour,
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	return logger
}

func setupDatabase(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func setupRouter(logger *logrus.Logger, authHandler *auth.UserHandler, authMiddleware gin.HandlerFunc) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestLogger(logger))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/forgot-password", authHandler.ForgotPassword)
	router.POST("/reset-password", authHandler.ResetPassword)

	// Protected routes
	api := router.Group("/api")
	api.Use(authMiddleware)
	//{
	//	// User routes
	//	api.GET("/profile", getUserProfile) //
	//	api.PUT("/profile", updateProfile)  //
	//
	//	// Wallet routes
	//	wallet := api.Group("/wallet")
	//	{
	//		wallet.GET("", getWalletDetails)        //
	//		wallet.POST("/topup", topUpWallet)      //
	//		wallet.POST("/transfer", transferMoney) //
	//	}
	//
	//	// Transaction routes
	transactions := api.Group("/transactions")
	{
		transactions.GET("")
		//transactions.GET("/:id", getTransaction)
	}
	//
	//	// Game routes
	//	game := api.Group("/game")
	//	{
	//		game.GET("/attempts", getGameAttempts) // TODO: Implement this
	//		game.POST("/play", playGame)           // TODO: Implement this
	//	}
	//}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func requestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request
		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration.String(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("Request processed")
	}
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := setupLogger()

	// Setup database
	db, err := setupDatabase(config)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	authRepo := repository.NewUserRepository(db)
	//transactionRepo := repository.NewTransactionRepository(db)

	// Initialize services
	authService := usecase.NewService(
		authRepo,
		config.JWTSecret,
		config.JWTIssuer,
		config.JWTDuration,
	)

	//transactionService := usecase.NewTransactionService(
	//	transactionRepo,
	//)
	// TODO: Initialize other services

	// Initialize handlers
	authHandler := auth.NewUserHandler(authService)
	//txHandler := auth.NewTransactionHandler(transactionService)

	// TODO: Initialize other handlers

	// Setup router
	router := setupRouter(logger, authHandler, middleware.AuthMiddleware(authService))

	// Start server
	serverAddr := fmt.Sprintf(":%s", config.ServerPort)
	logger.Infof("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
