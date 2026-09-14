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

// GatewayResponse represents the JSON structure returned by the Gateway
type GatewayResponse struct {
	Message string `json:"message"`
	Data    struct {
		DeviceId int  `json:"deviceId"`
		IsOnline bool `json:"isOnline"`
	} `json:"data"`
}

type deviceGatewayClient struct {
	baseURL        string
	httpClient     *http.Client
	internalSecret string
	prefixError    string
}

// NewDeviceGatewayClient initializes the client with a strict timeout for internal calls
func NewDeviceGatewayClient(baseURL, internalSecret string) model.DeviceGatewayClient {
	return &deviceGatewayClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 4 * time.Second, // Prevents the API from hanging if the Gateway crashes
		},
		internalSecret: internalSecret,
		prefixError:    "deviceGatewayClient",
	}
}

// GetDeviceStatus makes the HTTP GET request to the Gateway
func (c *deviceGatewayClient) GetDeviceStatus(ctx context.Context, deviceId int) (bool, error) {
	const fname = "GetDeviceStatus"
	url := fmt.Sprintf("%s/devicegateway/internal/devicestatus?deviceId=%d", c.baseURL, deviceId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)

	}

	req.Header.Set("X-Internal-Secret", c.internalSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("[%s]>[%s]: gateway returned unexpected status: %d", c.prefixError, fname, resp.StatusCode)
	}

	var gatewayData GatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gatewayData); err != nil {
		return false, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	return gatewayData.Data.IsOnline, nil
}

func (c *deviceGatewayClient) ExecuteCommand(ctx context.Context, req model.GatewayCommand) error {
	const fname = "ExecuteCommand"
	url := fmt.Sprintf("%s/devicegateway/internal/command", c.baseURL)

	// Encode the request body
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq.Header.Set("X-Internal-Secret", c.internalSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s]: gateway returned unexpected status: %d", c.prefixError, fname, resp.StatusCode)
	}

	return nil
}

func (c *deviceGatewayClient) InvalidateDeviceCache(ctx context.Context, oldDeviceName string) error {
	const fname = "InvalidateDeviceCache"
	url := fmt.Sprintf("%s/devicegateway/internal/clearcache?type=device&name=%s", c.baseURL, oldDeviceName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	req.Header.Set("X-Internal-Secret", c.internalSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s]: gateway returned unexpected status: %d", c.prefixError, fname, resp.StatusCode)
	}
	return nil
}

func (c *deviceGatewayClient) InvalidateGroupCache(ctx context.Context, oldGroupName string) error {
	const fname = "InvalidateGroupCache"
	url := fmt.Sprintf("%s/devicegateway/internal/clearcache?type=group&name=%s", c.baseURL, oldGroupName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	req.Header.Set("X-Internal-Secret", c.internalSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s]: gateway returned unexpected status: %d", c.prefixError, fname, resp.StatusCode)
	}
	return nil
}
