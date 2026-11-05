package rlog

import (
	"context"
	"log/slog"
	"slices"

	"github.com/jrozner/weby/middleware"
)

// RequestIDHandler is a slog.Handler that adds the request ID stored in the
// context under middleware.RequestIDKey to every record as the attribute "id".
// Records logged with a context that carries no request ID are passed through
// unchanged.
//
// Loggers derived with Logger.With or Logger.WithGroup (or handlers derived
// with WithAttrs or WithGroup) keep adding the request ID.
//
// The "id" attribute is always written at the top level of the record, even
// for loggers derived with WithGroup: a logger created with
// logger.WithGroup("g") logs "id=..." rather than "g.id=...". This keeps the
// request ID under a single key regardless of how a logger was derived, so log
// lines can always be correlated by it. To achieve this, a handler derived
// with WithGroup reapplies its groups and attributes to the underlying handler
// for each record that carries a request ID, which costs some extra
// allocations compared to an ungrouped handler.
type RequestIDHandler struct {
	slog.Handler
}

// Handle adds the request ID from ctx, if any, to r and passes it to the
// wrapped handler.
func (h RequestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	id, ok := ctx.Value(middleware.RequestIDKey).(string)
	if !ok {
		return h.Handler.Handle(ctx, r)
	}

	if d, ok := h.Handler.(*derivedHandler); ok && d.grouped {
		// Attributes added to r would land inside the open groups, so
		// replay the top-level attributes, apply the ID, and then replay
		// the groups.
		return d.replayWithID(id).Handle(ctx, r)
	}

	r = r.Clone()
	r.Add("id", id)
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a RequestIDHandler wrapping the underlying handler with
// attrs applied.
func (h RequestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return RequestIDHandler{Handler: h.derived().with(op{attrs: attrs})}
}

// WithGroup returns a RequestIDHandler wrapping the underlying handler with
// the group name applied. The request ID remains at the top level.
func (h RequestIDHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return RequestIDHandler{Handler: h.derived().with(op{group: name})}
}

func (h RequestIDHandler) derived() *derivedHandler {
	if d, ok := h.Handler.(*derivedHandler); ok {
		return d
	}
	return &derivedHandler{base: h.Handler, current: h.Handler}
}

// op is a single WithAttrs or WithGroup call.
type op struct {
	group string
	attrs []slog.Attr
}

func (o op) apply(h slog.Handler) slog.Handler {
	if o.group != "" {
		return h.WithGroup(o.group)
	}
	// WithAttrs takes ownership of its argument, while the derived handler
	// needs to retain these attributes for future replays.
	return h.WithAttrs(slices.Clone(o.attrs))
}

// derivedHandler records the WithAttrs and WithGroup calls made on top of a
// base handler so they can be replayed after the request ID is added.
type derivedHandler struct {
	base    slog.Handler
	ops     []op
	current slog.Handler // base with ops applied
	grouped bool         // whether ops contains a group
}

func (d *derivedHandler) with(o op) *derivedHandler {
	ops := make([]op, len(d.ops), len(d.ops)+1)
	copy(ops, d.ops)
	return &derivedHandler{
		base:    d.base,
		ops:     append(ops, o),
		current: o.apply(d.current),
		grouped: d.grouped || o.group != "",
	}
}

func (d *derivedHandler) replayWithID(id string) slog.Handler {
	h := d.base
	idApplied := false
	for _, o := range d.ops {
		if !idApplied && o.group != "" {
			h = h.WithAttrs([]slog.Attr{slog.String("id", id)})
			idApplied = true
		}
		h = o.apply(h)
	}
	if !idApplied {
		h = h.WithAttrs([]slog.Attr{slog.String("id", id)})
	}
	return h
}

func (d *derivedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.current.Enabled(ctx, level)
}

func (d *derivedHandler) Handle(ctx context.Context, r slog.Record) error {
	return d.current.Handle(ctx, r)
}

func (d *derivedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return d
	}
	return d.with(op{attrs: attrs})
}

func (d *derivedHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return d
	}
	return d.with(op{group: name})
}
