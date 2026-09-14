package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type scheduleEngineClient struct {
	baseURL        string
	httpClient     *http.Client
	internalSecret string
	prefixError    string
}

// NewSchedulerClient creates a new HTTP client with a strict 5-second timeout
func NewScheduleClient(baseURL, internalSecret string) model.ScheduleEngineClient {
	return &scheduleEngineClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		internalSecret: internalSecret,
		prefixError:    "scheduleEngineClient",
	}
}

func (c *scheduleEngineClient) SyncSchedule(ctx context.Context, scheduleID string) error {
	const fname = "SyncSchedule"
	url := fmt.Sprintf("%s/scheduleengine/internal/sync", c.baseURL)

	body := model.SyncJobReq{
		ScheduleId: scheduleID,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq.Header.Set("X-Internal-Secret", c.internalSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] scheduler service unreachable: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s] scheduler service returned non-200 status: %d", c.prefixError, fname, resp.StatusCode)
	}

	return nil
}

func (c *scheduleEngineClient) UnsyncSchedule(ctx context.Context, scheduleID string) error {
	const fname = "UnsyncSchedule"
	url := fmt.Sprintf("%s/scheduleengine/internal/unsync/%s", c.baseURL, scheduleID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq.Header.Set("X-Internal-Secret", c.internalSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] scheduler service unreachable: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s] scheduler service returned non-200 status: %d", c.prefixError, fname, resp.StatusCode)
	}

	return nil
}
