package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/darisadam/madabank-server/internal/api/handlers"
	"github.com/darisadam/madabank-server/internal/api/middleware"
	"github.com/darisadam/madabank-server/internal/pkg/jwt"
	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/darisadam/madabank-server/internal/pkg/metrics"
	"github.com/darisadam/madabank-server/internal/repository"
	"github.com/darisadam/madabank-server/internal/service"
)

var (
	// Version info (injected at build time)
	Version   = "dev"
	CommitSHA = "unknown"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	logger.Init(env)
	defer logger.Sync()

	// Set system info metrics
	metrics.SetSystemInfo(Version, CommitSHA, runtime.Version())

	// Connect to database
	db, err := initDB()
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		}
	}()

	logger.Info("Connected to database successfully")

	// Start metrics collector goroutine
	go collectSystemMetrics(db)

	// Initialize JWT service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required")
	}

	jwtExpiryHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if jwtExpiryHours == 0 {
		jwtExpiryHours = 24
	}
	jwtService := jwt.NewJWTService(jwtSecret, jwtExpiryHours)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, jwtService)
	accountService := service.NewAccountService(accountRepo)
	transactionService := service.NewTransactionService(transactionRepo, accountRepo, auditRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	accountHandler := handlers.NewAccountHandler(accountService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// Set Gin mode
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Metrics endpoint (Prometheus scraping)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check database connection
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "MadaBank API",
			"version":    Version,
			"commit_sha": CommitSHA,
			"go_version": runtime.Version(),
		})
	})

	// API version endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "MadaBank API",
			"version": Version,
			"status":  "operational",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// Protected routes (require authentication)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtService))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.DELETE("/profile", userHandler.DeleteAccount)
		}

		accounts := v1.Group("/accounts")
		accounts.Use(middleware.AuthMiddleware(jwtService))
		{
			accounts.POST("", accountHandler.CreateAccount)
			accounts.GET("", accountHandler.GetAccounts)
			accounts.GET("/:id", accountHandler.GetAccount)
			accounts.GET("/:id/balance", accountHandler.GetBalance)
			accounts.PATCH("/:id", accountHandler.UpdateAccount)
			accounts.DELETE("/:id", accountHandler.CloseAccount)
		}

		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthMiddleware(jwtService))
		{
			transactions.POST("/transfer", transactionHandler.Transfer)
			transactions.POST("/deposit", transactionHandler.Deposit)
			transactions.POST("/withdraw", transactionHandler.Withdraw)
			transactions.GET("/history", transactionHandler.GetHistory)
			transactions.GET("/:id", transactionHandler.GetTransaction)
		}
	}

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %s", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}

// collectSystemMetrics periodically collects system and business metrics
func collectSystemMetrics(db *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Database connection pool metrics
		stats := db.Stats()
		metrics.DBConnectionsActive.Set(float64(stats.OpenConnections))

		// Collect user metrics
		var totalUsers, activeUsers int
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&totalUsers); err != nil {
			logger.Error("Failed to collect total users metric", zap.Error(err))
		}
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND is_active = true").Scan(&activeUsers); err != nil {
			logger.Error("Failed to collect active users metric", zap.Error(err))
		}
		metrics.UpdateUserMetrics(totalUsers, activeUsers)

		// Collect account metrics
		var checkingCount, savingsCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM accounts WHERE account_type = 'checking' AND status = 'active'").Scan(&checkingCount); err != nil {
			logger.Error("Failed to collect checking accounts metric", zap.Error(err))
		}
		if err := db.QueryRow("SELECT COUNT(*) FROM accounts WHERE account_type = 'savings' AND status = 'active'").Scan(&savingsCount); err != nil {
			logger.Error("Failed to collect savings accounts metric", zap.Error(err))
		}
		metrics.UpdateAccountMetrics("checking", checkingCount)
		metrics.UpdateAccountMetrics("savings", savingsCount)
	}
}

func initDB() (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
