package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		if tel == "" {
			continue
		}
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
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("[%s]>[%s] returned unexpected status: %d", n.prefixError, fname, resp.StatusCode)
		}
	}

	return nil
}

func (n *notificationClient) SendEmail(ctx context.Context, emailUsers []model.UserNotificationSend) error {
	const fname = "SendEmail"

	auth := smtp.PlainAuth("", n.email.Username, n.email.Password, n.email.Host)
	addr := fmt.Sprintf("%s:%s", n.email.Host, n.email.Port)

	displayName := "Software Dashboard Tool System Alerts"
	fromHeader := fmt.Sprintf("From: %s <%s>\r\n", displayName, n.email.Username)
	subject := "Subject: SDT System Alert\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"

	// Group recipients by message body so a single SMTP session delivers to all recipients sharing the same alert
	recipientsByMsg := make(map[string][]string)
	var msgOrder []string

	for _, u := range emailUsers {
		emailAddr := strings.TrimSpace(u.Email)
		if emailAddr == "" {
			continue
		}
		if _, exists := recipientsByMsg[u.Msg]; !exists {
			msgOrder = append(msgOrder, u.Msg)
		}
		recipientsByMsg[u.Msg] = append(recipientsByMsg[u.Msg], emailAddr)
	}

	for _, body := range msgOrder {
		recipients := recipientsByMsg[body]
		if len(recipients) == 0 {
			continue
		}
		msg := []byte(fromHeader + subject + mime + body)
		if err := smtp.SendMail(addr, auth, n.email.Username, recipients, msg); err != nil {
			return fmt.Errorf("[%s]>[%s] failed to send email batch: %w", n.prefixError, fname, err)
		}
	}

	return nil
}

func (n *notificationClient) SendLine(ctx context.Context, lineUsers []model.UserNotificationSend) error {
	const fname = "SendLine"

	for _, u := range lineUsers {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", n.prefixError, fname, err)
		}

		if u.LineUserToken == nil {
			slog.DebugContext(ctx, "User has no line token for send notify")
			continue
		}

		lineUserToken := strings.TrimSpace(*u.LineUserToken)
		if lineUserToken == "" {
			continue
		}

		if strings.TrimSpace(u.Msg) == "" {
			continue
		}

		if err := n.lineClient.SendPushMessage(lineUserToken, u.Msg); err != nil {
			return fmt.Errorf("[%s]>[%s] failed to send LINE message to user token %s: %w", n.prefixError, fname, lineUserToken, err)
		}
	}

	return nil
}
