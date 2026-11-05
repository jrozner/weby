package rlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/jrozner/weby/middleware"
)

const testID = "0123-request-id"

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && (a.Key == slog.TimeKey || a.Key == slog.LevelKey) {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(RequestIDHandler{h})
}

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode log output %q: %v", buf.String(), err)
	}
	return got
}

func TestRequestIDHandler(t *testing.T) {
	withID := context.WithValue(context.Background(), middleware.RequestIDKey, testID)

	tests := []struct {
		name   string
		ctx    context.Context
		derive func(*slog.Logger) *slog.Logger
		want   map[string]any
	}{
		{
			name:   "plain",
			ctx:    withID,
			derive: func(l *slog.Logger) *slog.Logger { return l },
			want:   map[string]any{"msg": "hello", "k": "v", "id": testID},
		},
		{
			name:   "no request id",
			ctx:    context.Background(),
			derive: func(l *slog.Logger) *slog.Logger { return l },
			want:   map[string]any{"msg": "hello", "k": "v"},
		},
		{
			name:   "with",
			ctx:    withID,
			derive: func(l *slog.Logger) *slog.Logger { return l.With("a", "b") },
			want:   map[string]any{"msg": "hello", "a": "b", "k": "v", "id": testID},
		},
		{
			name:   "with chained",
			ctx:    withID,
			derive: func(l *slog.Logger) *slog.Logger { return l.With("a", "b").With("c", "d") },
			want:   map[string]any{"msg": "hello", "a": "b", "c": "d", "k": "v", "id": testID},
		},
		{
			name:   "with no request id",
			ctx:    context.Background(),
			derive: func(l *slog.Logger) *slog.Logger { return l.With("a", "b") },
			want:   map[string]any{"msg": "hello", "a": "b", "k": "v"},
		},
		{
			name:   "group",
			ctx:    withID,
			derive: func(l *slog.Logger) *slog.Logger { return l.WithGroup("g") },
			want: map[string]any{
				"msg": "hello",
				"id":  testID,
				"g":   map[string]any{"k": "v"},
			},
		},
		{
			name:   "group no request id",
			ctx:    context.Background(),
			derive: func(l *slog.Logger) *slog.Logger { return l.WithGroup("g") },
			want: map[string]any{
				"msg": "hello",
				"g":   map[string]any{"k": "v"},
			},
		},
		{
			name: "with then group",
			ctx:  withID,
			derive: func(l *slog.Logger) *slog.Logger {
				return l.With("a", "b").WithGroup("g").With("c", "d")
			},
			want: map[string]any{
				"msg": "hello",
				"id":  testID,
				"a":   "b",
				"g":   map[string]any{"c": "d", "k": "v"},
			},
		},
		{
			name: "nested groups",
			ctx:  withID,
			derive: func(l *slog.Logger) *slog.Logger {
				return l.WithGroup("g").With("a", "b").WithGroup("h")
			},
			want: map[string]any{
				"msg": "hello",
				"id":  testID,
				"g": map[string]any{
					"a": "b",
					"h": map[string]any{"k": "v"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := tt.derive(newTestLogger(&buf))

			if _, ok := logger.Handler().(RequestIDHandler); !ok {
				t.Fatalf("handler = %T, want RequestIDHandler", logger.Handler())
			}

			logger.InfoContext(tt.ctx, "hello", "k", "v")

			if got := decode(t, &buf); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("log = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestIDHandlerDerivedDoesNotModifyParent(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, testID)

	var buf bytes.Buffer
	parent := newTestLogger(&buf).WithGroup("g")
	_ = parent.With("a", "b")
	_ = parent.WithGroup("h")

	parent.InfoContext(ctx, "hello")

	want := map[string]any{"msg": "hello", "id": testID}
	if got := decode(t, &buf); !reflect.DeepEqual(got, want) {
		t.Fatalf("log = %v, want %v", got, want)
	}
}

type consumingHandler struct {
	slog.Handler
}

func (h consumingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	owned := append([]slog.Attr(nil), attrs...)
	next := h.Handler.WithAttrs(owned)
	clear(attrs)
	return consumingHandler{next}
}

func (h consumingHandler) WithGroup(name string) slog.Handler {
	return consumingHandler{h.Handler.WithGroup(name)}
}

func TestRequestIDHandlerRetainsOwnedAttrs(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, testID)

	var buf bytes.Buffer
	h := RequestIDHandler{consumingHandler{slog.NewJSONHandler(&buf, nil)}}
	logger := slog.New(h).With("a", "b").WithGroup("g")
	logger.InfoContext(ctx, "hello", "k", "v")

	got := decode(t, &buf)
	if got["a"] != "b" {
		t.Fatalf("a = %v, want b; log = %v", got["a"], got)
	}
}

func TestRequestIDHandlerIDPrecedence(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, testID)

	var buf bytes.Buffer
	logger := newTestLogger(&buf).With("id", "static").WithGroup("g")
	logger.InfoContext(ctx, "hello", "k", "v")

	if got := decode(t, &buf)["id"]; got != testID {
		t.Fatalf("id = %v, want %q", got, testID)
	}
}

func TestRequestIDHandlerDoesNotModifySharedRecord(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, testID)

	var buf bytes.Buffer
	h := RequestIDHandler{slog.NewJSONHandler(&buf, nil)}
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "hello", 0)
	for i := range 8 {
		r.AddAttrs(slog.Int("attr", i))
	}

	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	r.AddAttrs(slog.String("after", "value"))

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "!BUG" {
			t.Fatalf("handler modified shared record state: %v", a)
		}
		return true
	})
}

func TestRequestIDHandlerEnabled(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	logger := slog.New(RequestIDHandler{h}).With("a", "b").WithGroup("g")

	if logger.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info enabled, want disabled")
	}
	if !logger.Enabled(context.Background(), slog.LevelWarn) {
		t.Fatal("warn disabled, want enabled")
	}
}

func TestRequestIDHandlerConformance(t *testing.T) {
	var buf bytes.Buffer
	h := RequestIDHandler{slog.NewJSONHandler(&buf, nil)}

	results := func() []map[string]any {
		var ms []map[string]any
		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	}

	if err := slogtest.TestHandler(h, results); err != nil {
		t.Fatal(err)
	}
}
