package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/FTATM/software-dashboard-tool/config"
	"github.com/FTATM/software-dashboard-tool/internal/client"
	"github.com/FTATM/software-dashboard-tool/internal/handler"
	"github.com/FTATM/software-dashboard-tool/internal/middleware"
	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/FTATM/software-dashboard-tool/internal/repo"
	"github.com/FTATM/software-dashboard-tool/internal/router"
	"github.com/FTATM/software-dashboard-tool/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerScheduler struct {
	dB              *pgxpool.Pool
	server          *http.Server
	scheduleService model.ScheduleEngineService
}

func (a *ServerScheduler) Run(ctx context.Context) error {
	slog.Info(fmt.Sprintf("Server starting on %s", a.server.Addr))

	// Start the engine
	if a.scheduleService != nil {
		if err := a.scheduleService.Start(ctx); err != nil {
			slog.Error("Failed to start schedule engine", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
	return a.server.ListenAndServe()
}

func (a *ServerScheduler) Close(ctx context.Context) {
	slog.Info("Executing graceful shutdown...")

	// 1. Stop receiving new HTTP requests & drain connections
	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", slog.String("error", err.Error()))
	}

	// 2. Shut down scheduler tasks
	if a.scheduleService != nil {
		if err := a.scheduleService.Shutdown(ctx); err != nil {
			slog.Error("Schedule engine shutdown error", slog.String("error", err.Error()))
		}
	}

	// 3. Close the DB pool last
	if a.dB != nil {
		a.dB.Close()
	}

	slog.Info("Server successfully stopped", slog.String("addr", a.server.Addr))
}

func InitializeScheduleEngine(ctx context.Context) (App, error) {
	var dbConfig config.DB = &config.PostgresConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // Use the passed-in ctx as the parent
	defer cancel()

	// Initialize the connection pool
	db, err := pgxpool.New(dbCtx, dbConfig.ConnectString())
	if err != nil {
		slog.Error("Unable to initialize database pool", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := db.Ping(dbCtx); err != nil {
		slog.Error("Database is unreachable", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Database connected!")

	gatewayBaseURL := GetEnvOrDefault("DEVICE_GATEWAY_URL", "")
	internalSecret := GetEnvOrDefault("INTERNAL_API_SECRET", "")

	s3Config := config.S3{
		Url:         GetEnvOrDefault("S3_URL", ""),
		Region:      GetEnvOrDefault("S3_REGION", ""),
		AccessKey:   GetEnvOrDefault("S3_ACCESS_KEY", ""),
		SecretKey:   GetEnvOrDefault("S3_SECRET_KEY", ""),
		ImageBucket: GetEnvOrDefault("S3_IMAGE_BUCKET", ""),
	}

	// Dependency Injection
	scheduleRepo := repo.NewScheduleEngineRepository(db)
	deviceRepo := repo.NewDeviceRepository(db)
	widgetRepo := repo.NewWidgetRepository(db)
	auditLogRepo := repo.NewAuditLogRepository(db)

	gatewayClient := client.NewDeviceGatewayClient(gatewayBaseURL, internalSecret)
	s3Client := client.NewS3Client(s3Config)

	scheduleService, err := service.NewSchedulerEngineService(scheduleRepo, deviceRepo, gatewayClient, widgetRepo, s3Client, auditLogRepo)
	if err != nil {
		slog.Error("Failed to initialize service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handlers := router.RouterHandlers{
		ScheduleEngine: handler.NewScheduleEngineHandler(scheduleService),
	}

	mux := router.SetupSchedule(handlers)
	wrappedMux := middleware.LoggingApi(mux)

	serverPort := os.Getenv("SERVER_SCHEDULE_ENGINE_PORT")
	if serverPort == "" {
		serverPort = "8081"
		slog.Info(fmt.Sprintf("port fail default at %s", serverPort))
	}

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: wrappedMux,
	}

	return &ServerScheduler{
		dB:              db,
		server:          server,
		scheduleService: scheduleService,
	}, nil
}
