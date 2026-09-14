package service

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type logReportService struct {
	repo        model.LogReportRepository
	prefixError string
}

func NewLogReportService(repo model.LogReportRepository) model.LogReportService {
	return &logReportService{repo: repo, prefixError: "logReportService"}
}

func (s *logReportService) GetSystemLogsForExport(ctx context.Context, filter model.LogFilter) ([]model.AuditLogReport, error) {
	const fname = "GetSystemLogsForExport"
	log, err := s.repo.GetSystemLogsForExport(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return log, nil
}

func (s *logReportService) GetDeviceLogsForExport(ctx context.Context, filter model.LogFilter) ([]model.DeviceDataLogReport, error) {
	const fname = "GetDeviceLogsForExport"
	log, err := s.repo.GetDeviceLogsForExport(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return log, nil
}

func (s *logReportService) SearchSystemLogs(ctx context.Context, filter model.LogFilter) ([]model.AuditLogReport, error) {
	const fname = "SearchSystemLogs"
	log, err := s.repo.SearchSystemLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return log, nil
}

func (s *logReportService) SearchDeviceLogs(ctx context.Context, filter model.LogFilter) ([]model.DeviceDataLogReport, error) {
	const fname = "SearchDeviceLogs"
	log, err := s.repo.SearchDeviceLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return log, nil
}

// ⚡ Added: Fetch distinct menu types for the frontend dropdown
func (s *logReportService) GetAuditLogMenuTypes(ctx context.Context) ([]string, error) {
	const fname = "GetAuditLogMenuTypes"
	types, err := s.repo.GetAuditLogMenuTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return types, nil
}

func (s *logReportService) GetAuditLogEntityTypes(ctx context.Context) ([]string, error) {
	const fname = "GetAuditLogEntityTypes"
	types, err := s.repo.GetAuditLogEntityTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return types, nil
}

func (s *logReportService) CountSystemLogs(ctx context.Context, filter model.LogFilter) (int, error) {
	const fname = "CountSystemLogs"
	count, err := s.repo.CountSystemLogs(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return count, nil
}

func (s *logReportService) CountDeviceLogs(ctx context.Context, filter model.LogFilter) (int, error) {
	const fname = "CountDeviceLogs"
	count, err := s.repo.CountDeviceLogs(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return count, nil
}
