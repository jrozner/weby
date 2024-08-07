package rlog

import (
	"context"
	"log/slog"

	"github.com/jrozner/weby/middleware"
)

type RequestIDHandler struct {
	slog.Handler
}

func (h RequestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		r.Add("id", id)
	}

	return h.Handler.Handle(ctx, r)
}
