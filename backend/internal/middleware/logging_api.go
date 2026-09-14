package middleware

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"uuid"
)

const MaxBodyLogSize = 4096 // 4KB Limit

// responseRecorder captures the status code and a limited response body
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	captured   int
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	if rec.captured < MaxBodyLogSize {
		chunk := b
		if rec.captured+len(chunk) > MaxBodyLogSize {
			chunk = chunk[:MaxBodyLogSize-rec.captured]
		}
		rec.body.Write(chunk)
		rec.captured += len(chunk)
	}
	return rec.ResponseWriter.Write(b)
}

func (rec *responseRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}

func LoggingApi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// ==========================================
		// 0. GENERATE & INJECT REQUEST ID
		// ==========================================
		// Generate the ID
		reqID := uuid.New()

		// Inject it into a new context
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)

		// Attach the new context to the HTTP request
		r = r.WithContext(ctx)

		// ==========================================
		// 1. CAPTURE REQUEST BODY (Safely)
		// ==========================================
		var reqBodyLog string
		isJSON := strings.Contains(r.Header.Get("Content-Type"), "application/json")

		if isJSON && r.Body != nil {
			bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, MaxBodyLogSize))
			reqBodyLog = string(bodyBytes)

			if len(bodyBytes) == MaxBodyLogSize {
				reqBodyLog += " ... [TRUNCATED]"
			}

			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyBytes), r.Body))
		}

		// Log the incoming request using the NEW context (r.Context())
		logAttrs := []any{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		}
		if reqBodyLog != "" {
			logAttrs = append(logAttrs, slog.String("request_body", reqBodyLog))
		}
		slog.InfoContext(r.Context(), "Incoming request", logAttrs...)

		// ==========================================
		// 2. PROCESS REQUEST
		// ==========================================
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(recorder, r)

		// ==========================================
		// 3. CAPTURE & LOG RESPONSE BODY
		// ==========================================
		var respBodyLog string
		isRespJSON := strings.Contains(recorder.Header().Get("Content-Type"), "application/json")

		if isRespJSON {
			respBodyLog = recorder.body.String()
			if recorder.captured >= MaxBodyLogSize {
				respBodyLog += " ... [TRUNCATED]"
			}
		}

		respAttrs := []any{
			slog.Int("status", recorder.statusCode),
			slog.String("duration", time.Since(start).String()),
		}
		if respBodyLog != "" {
			respAttrs = append(respAttrs, slog.String("response_body", respBodyLog))
		}

		slog.InfoContext(r.Context(), "Request completed", respAttrs...)
	})
}
