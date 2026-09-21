package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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

type ServerApi struct {
	DB                *pgxpool.Pool
	Server            *http.Server
	DeviceStartPublic func(ctx context.Context)
}

func (a *ServerApi) Run(ctx context.Context) error {
	slog.Info(fmt.Sprintf("Server starting on %s", a.Server.Addr))

	go a.DeviceStartPublic(ctx)

	return a.Server.ListenAndServe()
}

func (a *ServerApi) Close(ctx context.Context) {
	slog.Info("Executing graceful shutdown...")

	// 1. Shut down the HTTP server FIRST
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			slog.Error("API HTTP server shutdown error", slog.String("error", err.Error()))
		}
	}

	// 2. Close the database ONLY AFTER the server has stopped processing requests
	if a.DB != nil {
		a.DB.Close()
	}

	slog.Info(fmt.Sprintf("Server successfully stopped on %s", a.Server.Addr))
}

func InitializeApi(ctx context.Context) (App, error) {

	var dbConfig config.DB = &config.PostgresConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	// Create a context for the connection timeout
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	// JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtKey := []byte(jwtSecret)

	scheduleEngineURL := os.Getenv("SCHEDULE_ENGINE_URL")
	if scheduleEngineURL == "" {
		slog.Error("SCHEDULE_ENGINE_URL is not set in environment")
		os.Exit(1)
	}

	deviceGatewayURL := os.Getenv("DEVICE_GATEWAY_URL")
	if deviceGatewayURL == "" {
		slog.Error("DEVICE_GATEWAY_URL is not set in environment")
		os.Exit(1)
	}

	internalSecret := os.Getenv("INTERNAL_API_SECRET")
	if internalSecret == "" {
		slog.Error("INTERNAL_API_SECRET is empty")
		os.Exit(1)
	}

	cooldownNotifSend, err := strconv.Atoi(os.Getenv("COOLDOWN_NOTIF_SEND"))
	if err != nil {
		slog.Error("COOLDOWN_NOTIF_SEND is Invalide")
		os.Exit(1)
	}

	s3Config := config.S3{
		Url:         os.Getenv("S3_URL"),
		Region:      os.Getenv("S3_REGION"),
		AccessKey:   os.Getenv("S3_ACCESS_KEY"),
		SecretKey:   os.Getenv("S3_SECRET_KEY"),
		ImageBucket: os.Getenv("S3_IMAGE_BUCKET"),
	}

	httpsConfig := GetEnvBoolOrDefault("HTTPS", false)

	// Dependency Injection
	//? DB
	txManager := repo.NewTxManager(db)
	widgetRepo := repo.NewWidgetRepository(db)
	widgetTypeRepo := repo.NewWidgetTypeRepository(db)
	canvasRepo := repo.NewCanvasRepository(db)
	userRepo := repo.NewUserRepository(db)
	deviceRepo := repo.NewDeviceRepository(db)
	auditLogRepo := repo.NewAuditLogRepository(db)
	roleRepo := repo.NewRoleRepository(db)
	scheduleRepo := repo.NewScheduleRepository(db)
	logReportRepo := repo.NewLogReportRepository(db)
	notificationRepo := repo.NewNotificationRepository(db)

	//? client
	s3Client := client.NewS3Client(s3Config)
	scheduleClient := client.NewScheduleClient(scheduleEngineURL, internalSecret)
	deviceGatewayClient := client.NewDeviceGatewayClient(deviceGatewayURL, internalSecret)
	lineClient := client.NewLineClient(config.Line{
		LineChannelID:      GetEnvOrDefault("LINE_CHANNEL_ID", ""),
		ChannelAccessToken: GetEnvOrDefault("LINE_CHANNEL_ACCESS_TOKEN", ""),
	})

	//? service
	roleCache := service.NewRoleCache()
	widgetService := service.NewWidgetService(txManager, widgetRepo, widgetTypeRepo, canvasRepo, auditLogRepo)
	canvasService := service.NewCanvasService(txManager, widgetRepo, canvasRepo, auditLogRepo)
	widgetTypeService := service.NewWidgetTypeService(txManager, widgetTypeRepo)
	userService := service.NewUserService(txManager, userRepo, jwtKey, roleRepo, auditLogRepo, roleCache)
	deviceService := service.NewDeviceService(txManager, deviceRepo, auditLogRepo)
	roleService := service.NewRoleService(txManager, roleRepo, auditLogRepo, roleCache)
	scheduleService := service.NewScheduleService(txManager, scheduleRepo, auditLogRepo)
	logReportService := service.NewLogReportService(logReportRepo)
	notificationService := service.NewNotificationService(txManager, notificationRepo, auditLogRepo, nil, make(chan []model.DeviceData), cooldownNotifSend)

	handlers := router.RouterHandlers{
		Widget:       handler.NewWidgetHandler(widgetService, roleService),
		Canvas:       handler.NewCanvasHandler(canvasService, roleService),
		WidgetType:   handler.NewWidgetTypeHandler(widgetTypeService),
		User:         handler.NewUserHandler(userService, roleService, httpsConfig, lineClient),
		Device:       handler.NewDeviceHandler(deviceService, roleService, deviceGatewayClient, auditLogRepo),
		Role:         handler.NewRoleHandler(roleService),
		Schedule:     handler.NewScheduleHandler(scheduleService, roleService, scheduleClient),
		LogReport:    handler.NewLogReportHandler(logReportService),
		Notification: handler.NewNotificationHandler(notificationService, roleService),
		S3File:       handler.NewS3FileHandler(s3Client),
	}

	// Initialize Router (using the struct bundle from earlier)
	mux := router.SetupApi(handlers)

	// Chain Middleware
	// AuthMiddleware is global check method for open public route
	protectedMux := middleware.AuthApi(jwtKey, mux)
	wrappedMux := middleware.LoggingApi(protectedMux)

	// Configure HTTP Server

	serverPort := os.Getenv("SERVER_API_PORT")
	if serverPort == "" {
		serverPort = "8080"
		slog.Info(fmt.Sprintf("port fail default at %s", serverPort))
	}
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: wrappedMux,
	}

	count, err := userRepo.UserCount(ctx)
	if err != nil {
		slog.Error("Error at count user", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if count == 0 {
		adminUser := os.Getenv("INITIAL_ADMIN_USERNAME")
		adminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
		createUserReq := model.CreateUser{
			FirstName: "system",
			LastName:  "admin",
			Username:  adminUser,
			Password:  adminPassword,
			Active:    true,
			RoleId:    1,
		}
		user, err := userService.CreateUser(ctx, &createUserReq, 0)
		if err != nil {
			slog.Error("Error at init admin user", slog.String("error", err.Error()))
			os.Exit(1)
		}

		slog.Info(fmt.Sprintf("init admin user: %s  %s", user.FirstName, user.LastName))
	}

	return &ServerApi{
		DB:                db,
		Server:            server,
		DeviceStartPublic: deviceService.StartPublic,
	}, nil
}
