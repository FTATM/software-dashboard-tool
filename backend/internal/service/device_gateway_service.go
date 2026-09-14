package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type deviceGatewayService struct {
	repo        model.DeviceGatewayRepository
	prefixError string

	// Batching state
	inputCh      chan model.DeviceData
	batchSize    int
	timeout      time.Duration
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	notifService model.NotificationService

	deviceNameCache  sync.Map
	deviceGroupCache sync.Map
}

func NewDeviceGatewayService(repo model.DeviceGatewayRepository, batchSize int, timeout time.Duration, notifService model.NotificationService) model.DeviceGatewayService {
	return &deviceGatewayService{
		repo:         repo,
		prefixError:  "deviceGatewayService",
		inputCh:      make(chan model.DeviceData, batchSize),
		batchSize:    batchSize,
		timeout:      timeout,
		notifService: notifService,
	}
}

func (s *deviceGatewayService) Start(parentCtx context.Context) error {
	const fname = "Start"

	s.ctx, s.cancel = context.WithCancel(parentCtx)
	s.wg.Add(1)

	slog.InfoContext(s.ctx, fmt.Sprintf("[%s]>[%s] Starting Device Gateway Batcher...", s.prefixError, fname))

	go func() {
		defer s.wg.Done()
		batch := make([]model.DeviceData, 0, s.batchSize)
		ticker := time.NewTicker(s.timeout)
		defer ticker.Stop()

		flush := func() {
			if len(batch) > 0 {
				// Use a new timeout context so the DB write doesn't block forever
				dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer dbCancel()

				if err := s.repo.BulkUpsertDeviceData(dbCtx, batch); err != nil {
					slog.ErrorContext(s.ctx, "Failed to flush batch to database", slog.String("error", err.Error()))
				} else {
					slog.InfoContext(s.ctx, "Flushed batch to database", slog.Int("count", len(batch)))
				}
				s.notifService.AddDeviceDataAlert(batch)
				// Re-slice to zero to reuse memory
				batch = batch[:0]
			}
		}

		for {
			select {
			case <-s.ctx.Done():
				slog.InfoContext(s.ctx, "Shutting down batcher, flushing remaining data...")
				flush()
				return

			case data := <-s.inputCh:
				batch = append(batch, data)
				if len(batch) >= s.batchSize {
					flush()
					ticker.Reset(s.timeout) // Reset timer to avoid immediate double-flush
				}

			case <-ticker.C:
				flush()
			}
		}
	}()

	return nil
}

func (s *deviceGatewayService) Add(data model.DeviceData) {
	data.ReceivedAt = time.Now()
	select {
	case s.inputCh <- data:
		// Success! The channel had room and accepted the data.
	default:
		// The channel is completely full (Database is likely lagging).
		// Drop the packet and log a warning so the network threads don't freeze!
		slog.Warn("Gateway batcher channel is full! Dropping telemetry to prevent network block",
			slog.Int("deviceId", data.DeviceId),
		)
	}
}

func (s *deviceGatewayService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait() // Wait for the final DB flush to complete before exiting
	slog.Info("Device Gateway Batcher stopped safely")
}

func (s *deviceGatewayService) UpdateDeviceLastSeen(ctx context.Context, deviceId int) error {
	const fname = "UpdateDeviceLastSeen"
	err := s.repo.UpdateLastSeen(ctx, deviceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return nil
}
