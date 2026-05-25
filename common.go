package validate

import (
	"context"
	"errors"
	"strings"
)

type (
	Name string

	Handle interface {
		ValidateHandle(ctx context.Context, value any, opts ...any) error
	}

	resolver interface {
		Resolve(name Name) (Rule, bool)
	}

	tagLookup interface {
		Lookup(key string) (value string, ok bool)
	}
)

type HandlerFunc func(ctx context.Context, value any, opts ...any) error

func (f HandlerFunc) ValidateHandle(ctx context.Context, value any, opts ...any) error {
	return f(ctx, value, opts...)
}

type Rule struct {
	Name   Name
	Handle Handle
}

func (r Rule) Validate() error {
	if r.Handle == nil {
		return errors.New("no handler for rule")
	}
	if len(strings.TrimSpace(string(r.Name))) == 0 {
		return errors.New("no name for rule")
	}
	return nil
}

type tagInfo struct {
	Name Name
	Opts []any
}
