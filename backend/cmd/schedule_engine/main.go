package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/DeRuina/timberjack"
	"github.com/FTATM/software-dashboard-tool/internal/app"
	"github.com/FTATM/software-dashboard-tool/internal/middleware"
)

var (
	Version   = "dev"
	CommitSHA = "none"
	BuildTime = "unknown"
)

func main() {
	logCloser := initLogger()
	defer func() {
		if logCloser != nil {
			_ = logCloser.Close()
		}
	}()

	slog.Info("Starting application initialization...")

	// 2. Listen for cancellation signals
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Initialize dependency graph
	application, err := app.InitializeScheduleEngine(rootCtx)
	if err != nil {
		slog.Error("Failed to initialize schedule engine", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Start the server in a goroutine and capture fatal runtime errors
	serverErr := make(chan error, 1)
	go func() {
		if err := application.Run(rootCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 5. Block until SIGINT/SIGTERM or unexpected server error
	select {
	case <-rootCtx.Done():
		slog.Info("Shutdown signal received, starting graceful teardown...")
	case err := <-serverErr:
		slog.Error("Server crashed unexpectedly", slog.String("error", err.Error()))
	}

	// 6. Execute graceful shutdown with a dedicated timeout context
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	application.Close(shutdownCtx)
	slog.Info("Application terminated cleanly")
}

func initLogger() io.Closer {
	logDir := app.GetEnvOrDefault("LOG_SCHEDULE_ENGINE_DIR", "log/schedule")
	slog.Info("logDir", slog.String("path", logDir))

	logMaxSize := app.GetEnvIntOrDefault("LOG_SCHEDULE_ENGINE_MAX_SIZE", 10)
	logMaxBackup := app.GetEnvIntOrDefault("LOG_SCHEDULE_ENGINE_MAX_BACKUP", 5)
	logMaxAge := app.GetEnvIntOrDefault("LOG_SCHEDULE_ENGINE_MAX_AGE", 28)

	// Safely create the directory if it doesn't exist yet
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the log folder, panic before the app starts
		panic(fmt.Sprintf("Failed to create log directory: %s | error: %s", logDir, err.Error()))
	}

	// Set up the lumberjack file writer
	logRotator := &timberjack.Logger{
		Filename:         filepath.Join(logDir, "schedule_engine_app.log"), // Uses the dynamic logDir
		MaxSize:          logMaxSize,                                       // Max size in Megabytes before rotating
		MaxBackups:       logMaxBackup,                                     // Max number of old files to keep
		MaxAge:           logMaxAge,                                        // Max number of days to keep old files
		LocalTime:        true,                                             // false will use UTC
		Compression:      "zstd",                                           // Explicitly set compression type
		RotationInterval: 24 * time.Hour,                                   // Rotate every 24 hours
		RotateAt:         []string{"00:00"},                                // Force rotation exactly at midnight
	}

	logLevelStr := os.Getenv("LOG_SCHEDULE_ENGINE_LEVEL")
	var logLevel slog.Level
	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // Default fallback for production
	}

	// Configure slog options
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Clean up the "source" file path so it's not overly long
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
				source.Function = filepath.Base(source.Function)
			}
			return a
		},
	}

	// MULTI-WRITER FIX: Write to both os.Stdout (Console) AND logRotator (File)
	multiWriter := io.MultiWriter(os.Stdout, logRotator)

	// Create the standard JSON Handler and set it directly
	jsonHandler := slog.NewJSONHandler(multiWriter, opts)

	// FIX: Use the ContextHandler from your middleware package
	ctxHandler := middleware.ContextHandler{Handler: jsonHandler}

	slog.SetDefault(slog.New(ctxHandler))
	return logRotator
}
