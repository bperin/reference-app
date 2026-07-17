package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/example/reference-app/internal/auth"
	authrepository "github.com/example/reference-app/internal/auth/repository"
	"github.com/example/reference-app/internal/config"
	"github.com/example/reference-app/internal/database"
	"github.com/example/reference-app/internal/health"
	"github.com/example/reference-app/internal/logging"
	"github.com/example/reference-app/internal/posts"
	postsrepository "github.com/example/reference-app/internal/posts/repository"
	postssqlc "github.com/example/reference-app/internal/posts/repository/sqlc"
	"github.com/example/reference-app/internal/security"
	"github.com/example/reference-app/internal/users"
	usersrepository "github.com/example/reference-app/internal/users/repository"
	userssqlc "github.com/example/reference-app/internal/users/repository/sqlc"
)

// @title Reference API
// @version 0.1.0
// @description Core reference microservice with full users and posts CRUD.
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.oauth2.password OAuth2Auth
// @tokenUrl /auth/oauth/token
// @scope.user Access the authenticated user's profile and posts.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "reference-app failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// 2. Print configuration (sensitive credentials are redacted)
	cfg.Print()

	// 3. Initialize root logger
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := logging.New(cfg.AppEnv, logLevel, os.Stdout)
	slog.SetDefault(logger)

	logger.Info("Starting API service", slog.String("env", cfg.AppEnv))

	// 4. Create PostgreSQL pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	// 5. Run automatic migrations
	logger.Info("Running migrations...")
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	logger.Info("Migrations completed successfully.")

	// 6. Initialize domain sqlc stores
	usersStore := userssqlc.New(pool)
	postsStore := postssqlc.New(pool)

	// 6. Initialize repository adapters
	usersRepository := usersrepository.NewPostgres(usersStore)
	postsRepository := postsrepository.NewPostgres(postsStore)
	refreshRepository := authrepository.NewPostgres(pool)

	// 7. Initialize application/domain services
	usersService := users.NewService(usersRepository, logger)
	postsService := posts.NewService(postsRepository, logger)
	tokenIssuer := security.NewTokenIssuer(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTokenTTL)
	authService := auth.NewService(
		usersRepository,
		refreshRepository,
		security.NewPasswordHasher(),
		tokenIssuer,
		security.NewRefreshTokenGenerator(),
		logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	// 8. Initialize HTTP handlers
	usersHandler := users.NewHandler(usersService, logger)
	postsHandler := posts.NewHandler(postsService, logger)
	authHandler := auth.NewHandler(authService, logger)

	// 10. Construct Chi router
	r := chi.NewRouter()

	// Add global middlewares
	r.Use(logging.RequestIDMiddleware)
	r.Use(logging.RequestLoggerMiddleware(logger))
	r.Use(logging.CORSMiddleware)

	// Each domain owns its route registration. The composition root supplies
	// shared infrastructure such as bearer authentication.
	r.Get("/healthz", health.Handler)
	auth.RegisterRoutes(r, authHandler)
	requireBearer := auth.RequireBearer(tokenIssuer)
	users.RegisterRoutes(r, usersHandler, requireBearer)
	posts.RegisterRoutes(r, postsHandler, requireBearer)

	// 11. Start HTTP Server
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("address", cfg.HTTPAddress))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	signalContext, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("serve HTTP: %w", err)
	case <-signalContext.Done():
	}

	logger.Info("Shutting down HTTP server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}
	logger.Info("HTTP server stopped successfully.")
	return nil
}
