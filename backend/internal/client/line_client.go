package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FTATM/software-dashboard-tool/config"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type lineClient struct {
	httpClient  *http.Client
	prefixError string
	lineConfig  config.Line
}

type lineVerifyResponse struct {
	Iss              string `json:"iss"`
	Sub              string `json:"sub"`
	Aud              string `json:"aud"`
	Exp              int64  `json:"exp"`
	Name             string `json:"name"`
	Picture          string `json:"picture"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type LineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type LinePushPayload struct {
	To       string            `json:"to"`       // The Sub / Line User ID (Uxxxx...)
	Messages []LineTextMessage `json:"messages"` // Array of messages (up to 5)
}

// NewLineClient accepts both the LINE Login Channel ID and the Messaging API Channel Access Token
func NewLineClient(lineConfig config.Line) model.LineClient {
	return &lineClient{
		lineConfig:  lineConfig,
		prefixError: "lineClient",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *lineClient) VerifyIDToken(idToken string) (string, error) {
	data := url.Values{}
	data.Set("id_token", idToken)
	data.Set("client_id", c.lineConfig.LineChannelID)

	req, err := http.NewRequest("POST", "https://api.line.me/oauth2/v2.1/verify", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("%s: failed to create verify request: %w", c.prefixError, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: failed to call line verify: %w", c.prefixError, err)
	}
	defer resp.Body.Close()

	var result lineVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%s: failed to decode response: %w", c.prefixError, err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := result.ErrorDescription
		if errMsg == "" {
			errMsg = result.Error
		}
		if errMsg == "" {
			errMsg = fmt.Sprintf("status code %d", resp.StatusCode)
		}
		return "", fmt.Errorf("%s: line verification error: %s", c.prefixError, errMsg)
	}

	return result.Sub, nil
}

func (c *lineClient) SendPushMessage(lineUserID string, messageText string) error {
	payload := LinePushPayload{
		To: lineUserID,
		Messages: []LineTextMessage{
			{
				Type: "text",
				Text: messageText,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal push payload: %w", c.prefixError, err)
	}

	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%s: failed to create push request: %w", c.prefixError, err)
	}

	// Required Headers
	req.Header.Set("Content-Type", "application/json")
	// Must use the Channel Access Token here:
	req.Header.Set("Authorization", "Bearer "+c.lineConfig.ChannelAccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: failed to send push request: %w", c.prefixError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: line push api error (status %d): %s", c.prefixError, resp.StatusCode, string(resBody))
	}

	return nil
}
