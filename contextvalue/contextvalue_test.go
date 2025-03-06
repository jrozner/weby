package contextvalue

import (
	"context"
	"testing"
)

var key Key = "hi"

func TestNoValue(t *testing.T) {
	ctx := context.Background()
	_, err := Value[string](ctx, key)
	if err != ErrNoValue {
		t.Errorf("want ErrNoValue, got %v", err)
	}
}

func TestAssertFailed(t *testing.T) {
	value := "hello"

	ctx := context.WithValue(context.Background(), key, value)
	_, err := Value[int](ctx, "hi")
	if err != ErrAssertFailed {
		t.Errorf("want ErrAssertFailed, got %v", err)
	}
}

func TestString(t *testing.T) {
	value := "hello"

	ctx := context.WithValue(context.Background(), key, value)
	v, err := Value[string](ctx, "hi")
	if err != nil {
		t.Errorf("got error when expected success: %s", err)
	}

	if v != value {
		t.Errorf("want %v, got %v", value, v)
	}
}
