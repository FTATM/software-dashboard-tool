package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/FTATM/software-dashboard-tool/config"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type notificationClient struct {
	sms         config.Sms
	email       config.Email
	lineClient  model.LineClient
	httpClient  *http.Client
	prefixError string
}

func NewNotificationClient(sms config.Sms, email config.Email, lineClient model.LineClient) model.NotificationClient {
	return &notificationClient{
		sms:   sms,
		email: email,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		lineClient:  lineClient,
		prefixError: "notificationClient",
	}
}

func (n *notificationClient) SendSms(ctx context.Context, smsUser []model.UserNotificationSend) error {
	const fname = "SendSms"

	for _, u := range smsUser {
		tel := strings.TrimSpace(u.Tel)
		if after, ok := strings.CutPrefix(tel, "0"); ok {
			tel = "66" + after
		}
		payload := struct {
			Msisdn  string `json:"msisdn"`
			Message string `json:"message"`
			Sender  string `json:"sender"`
			Force   string `json:"force"`
		}{
			Msisdn:  tel,
			Message: u.Msg,
			Sender:  n.sms.Sender,
			Force:   "standard",
		}

		payloadJsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", n.prefixError, fname, err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.sms.Url, bytes.NewReader(payloadJsonData))
		if err != nil {
			return fmt.Errorf("[%s]>[%s] failed to create request: %w", n.prefixError, fname, err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(n.sms.Key, n.sms.Secret)

		resp, err := n.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("[%s]>[%s] request failed: %w", n.prefixError, fname, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("[%s]>[%s] returned unexpected status: %w", n.prefixError, fname, err)
		}
	}

	return nil
}

func (n *notificationClient) SendEmail(ctx context.Context, emailUsers []model.UserNotificationSend) error {
	const fname = "SendEmail"

	// 1. Set up SMTP Authentication
	auth := smtp.PlainAuth("", n.email.Username, n.email.Password, n.email.Host)

	// Example: "smtp.gmail.com:587"
	addr := fmt.Sprintf("%s:%s", n.email.Host, n.email.Port)

	// 2. Loop through users and send emails
	for _, u := range emailUsers {
		// Skip if the user doesn't have an email address configured
		if u.Email == "" {
			continue
		}

		// 3. Construct the email headers and body
		displayName := "Software Dashboard Tool System Alerts"

		// ⚡ Add the From header using your friendly name and your Gmail Username
		fromHeader := fmt.Sprintf("From: %s <%s>\r\n", displayName, n.email.Username)
		subject := "Subject: SDT System Alert\r\n"
		mime := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"

		body := u.Msg

		// Combine all parts into the final message payload
		msg := []byte(fromHeader + subject + mime + body)

		// 4. Send the email via SMTP
		// ⚡ Replace n.email.Sender with n.email.Username here!
		err := smtp.SendMail(addr, auth, n.email.Username, []string{u.Email}, msg)
		if err != nil {
			// Return the error to be logged by the service
			return fmt.Errorf("[%s]>[%s] failed to send email to %s: %w", n.prefixError, fname, u.Email, err)
		}
	}

	return nil
}

func (n *notificationClient) SendLine(ctx context.Context, lineUsers []model.UserNotificationSend) error {
	const fname = "SendLine"

	for _, u := range lineUsers {
		// Respect context cancellation (timeout or caller cancellation)
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", n.prefixError, fname, err)
		}

		if u.LineUserToken == nil {
			slog.DebugContext(ctx, "User has no line token for send notify")
			continue
		}

		// Skip users without a linked LINE account
		lineUserToken := strings.TrimSpace(*u.LineUserToken)
		if lineUserToken == "" {
			continue
		}

		// Skip empty messages
		if strings.TrimSpace(u.Msg) == "" {
			continue
		}

		// Send push message via LineClient
		if err := n.lineClient.SendPushMessage(lineUserToken, u.Msg); err != nil {
			return fmt.Errorf("[%s]>[%s] failed to send LINE message to user token %s: %w", n.prefixError, fname, lineUserToken, err)
		}
	}

	return nil
}
