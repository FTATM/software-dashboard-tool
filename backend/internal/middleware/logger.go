package middleware

import (
	"context"
	"fmt"
	"log/slog"
)

// Define the key ONCE here
type ctxKey string

const RequestIDKey ctxKey = "request_id"

// ContextHandler wraps the default handler to extract the Request ID
type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. We must type assert the UUID to a string!
	// (Your middleware uses uuid.New(), which returns a uuid.UUID, not a string)
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		r.AddAttrs(slog.String("request_id", reqID))
	} else if reqID, ok := ctx.Value(RequestIDKey).(fmt.Stringer); ok {
		// Fallback: If it's a uuid.UUID, we can call .String() on it
		r.AddAttrs(slog.String("request_id", reqID.String()))
	}
	return h.Handler.Handle(ctx, r)
}
