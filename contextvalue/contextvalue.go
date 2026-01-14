package contextvalue

import (
	"context"
	"errors"
)

type Key string

var (
	// ErrNoValue is returned when the value does not exist within the context
	ErrNoValue = errors.New("no value in context")
	// ErrAssertFailed is returned when the value cannot be asserted to the specified type
	ErrAssertFailed = errors.New("assert failed")
)

// Value returns the asserted value from a context or an error if the key exists, and it can assert to the specified
// type. The failure conditions unfortunately force an allocation of T because it's not possible to create a literal
// of T generically. Instead, we need to allocate it and then de-reference.
func Value[T any](ctx context.Context, key Key) (T, error) {
	var (
		v         interface{}
		vAsserted T
		ok        bool
	)

	if v = ctx.Value(key); v == nil {
		return vAsserted, ErrNoValue
	}

	if vAsserted, ok = v.(T); ok {
		return vAsserted, nil
	}

	return vAsserted, ErrAssertFailed
}
