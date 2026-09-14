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
	"github.com/FTATM/software-dashboard-tool/internal/listener"
	"github.com/FTATM/software-dashboard-tool/internal/middleware"
	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/FTATM/software-dashboard-tool/internal/repo"
	"github.com/FTATM/software-dashboard-tool/internal/router"
	"github.com/FTATM/software-dashboard-tool/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerDeviceGateway struct {
	dB                 *pgxpool.Pool
	server             *http.Server
	gatewayService     model.DeviceGatewayService
	sessionService     model.SessionManagerService
	deviceCacheService model.DeviceCacheService
	notifiService      model.NotificationService
}

func (a *ServerDeviceGateway) Run(ctx context.Context) error {
	slog.Info("Device Gateway starting listeners...")
	gatewayPort := os.Getenv("DEVICE_GATEWAY_PORT")
	if gatewayPort == "" {
		gatewayPort = "8090"
	}

	if err := a.gatewayService.Start(ctx); err != nil {
		slog.Error("Failed to start device gateway service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go a.notifiService.StartDeviceRuleAlert()

	// ⚡ NEW: Start the cache sweeper in the background
	a.deviceCacheService.StartSweeper(ctx)
	a.sessionService.StartSweeper(ctx)

	// ⚡ INJECTION: Pass the new cacheService into the listeners!
	go listener.StartTCPServer(ctx, ":"+gatewayPort, a.gatewayService, a.sessionService, a.deviceCacheService)
	go listener.StartUDPServer(ctx, ":"+gatewayPort, a.gatewayService, a.sessionService, a.deviceCacheService)

	brokerURL := os.Getenv("MQTT_BROKER_URL")
	if brokerURL != "" {
		go listener.StartMQTTClient(ctx, brokerURL, a.gatewayService, a.sessionService, a.deviceCacheService)
	}

	slog.Info(fmt.Sprintf("Gateway HTTP Server starting on %s", a.server.Addr))
	return a.server.ListenAndServe()
}

func (a *ServerDeviceGateway) Close(ctx context.Context) {
	slog.Info("Executing graceful shutdown for Device Gateway...")

	// Shut down HTTP Server
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			slog.Error("Gateway HTTP server shutdown error", slog.String("error", err.Error()))
		}
	}

	// Shut down the in-memory batcher engine (flushes remaining records to DB)
	if a.gatewayService != nil {
		a.gatewayService.Stop()
	}

	// Shut down PostgreSQL connection pool
	if a.dB != nil {
		a.dB.Close()
	}

	slog.Info(fmt.Sprintf("Device Gateway successfully stopped on %s", a.server.Addr))
}

func InitializeDeviceGateway(ctx context.Context) (App, error) {
	// 1. Initialize Database Config EXACTLY like schedule engine
	var dbConfig config.DB = &config.PostgresConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

	slog.Info("Device Gateway Database connected!")

	cooldownNotifSend := GetEnvIntOrDefault("COOLDOWN_NOTIF_SEND", 3)
	qBuffer := GetEnvIntOrDefault("DEVICE_GATEWAY_QUEUE_BUFFER", 300)

	qTimeout, err := time.ParseDuration(os.Getenv("DEVICE_GATEWAY_QUEUE_TIMEOUT_SECOND") + "s")
	if err != nil {
		slog.Error("QUEUE TIMEOUT env is not correct", slog.String("error", err.Error()))
		os.Exit(1)
	}

	txManager := repo.NewTxManager(db)
	gatewayRepo := repo.NewDeviceGatewayRepository(db)
	notificationRepo := repo.NewNotificationRepository(db)
	auditLogRepo := repo.NewAuditLogRepository(db)

	notificationClient := client.NewNotificationClient(
		config.Sms{
			Url:    GetEnvOrDefault("SMS_API_URL", ""),
			Key:    GetEnvOrDefault("SMS_API_KEY", ""),
			Secret: GetEnvOrDefault("SMS_API_SECRET", ""),
			Sender: GetEnvOrDefault("SMS_SENDER", ""),
		},
		config.Email{
			Host:       GetEnvOrDefault("MAIL_HOST", ""),
			Port:       GetEnvOrDefault("MAIL_PORT", ""),
			Encryption: GetEnvOrDefault("MAIL_ENCRYPTION", ""),
			Username:   GetEnvOrDefault("MAIL_USERNAME", ""),
			Password:   GetEnvOrDefault("MAIL_PASSWORD", ""),
		},
		config.Line{
			Token: GetEnvOrDefault("LINE_NOTIFY_TOKEN", ""),
		},
	)

	cacheService := service.NewCacheService(gatewayRepo)
	notificationService := service.NewNotificationService(txManager, notificationRepo, auditLogRepo, notificationClient, make(chan []model.DeviceData, qBuffer), cooldownNotifSend)
	gatewayService := service.NewDeviceGatewayService(gatewayRepo, qBuffer, qTimeout, notificationService)
	sessionService := service.NewSessionManagerService(gatewayRepo, cacheService)

	handlers := router.RouterHandlers{
		DeviceGateway: handler.NewDeviceGatewayHandler(gatewayService, sessionService, cacheService),
	}

	mux := router.SetupDeviceGateway(handlers)
	wrappedMux := middleware.LoggingApi(mux)

	// 5. HTTP Server Setup
	serverPort := os.Getenv("DEVICE_GATEWAY_HTTP_PORT")
	if serverPort == "" {
		serverPort = "8091"
		slog.Info(fmt.Sprintf("Gateway HTTP port failed, defaulting to %s", serverPort))
	}

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: wrappedMux,
	}

	return &ServerDeviceGateway{
		dB:                 db,
		server:             server,
		gatewayService:     gatewayService,
		sessionService:     sessionService,
		deviceCacheService: cacheService,
		notifiService:      notificationService,
	}, nil
}
