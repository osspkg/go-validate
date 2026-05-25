package validate

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		rules     []Rule
		callback  func(c Callback)
		wantErr   bool
		errString string
	}{
		{
			name: "require success",
			rules: []Rule{
				{Name: "positive", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					if v, ok := value.(int); ok && v > 0 {
						return nil
					}
					return errors.New("must be positive")
				})},
			},
			callback: func(c Callback) {
				c.Require("positive", 42)
			},
			wantErr: false,
		},
		{
			name: "require failure",
			rules: []Rule{
				{Name: "positive", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					return errors.New("value not positive")
				})},
			},
			callback: func(c Callback) {
				c.Require("positive", -1)
			},
			wantErr: true,
		},
		{
			name: "optional zero value skipped",
			rules: []Rule{
				{Name: "positive", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					return errors.New("should not be called")
				})},
			},
			callback: func(c Callback) {
				c.Optional("positive", 0)
			},
			wantErr: false,
		},
		{
			name: "optional non-zero called",
			rules: []Rule{
				{Name: "positive", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					if v, ok := value.(int); ok && v > 0 {
						return nil
					}
					return errors.New("must be positive")
				})},
			},
			callback: func(c Callback) {
				c.Optional("positive", 5)
			},
			wantErr: false,
		},
		{
			name: "rule not found",
			rules: []Rule{
				{Name: "exists", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil })},
			},
			callback: func(c Callback) {
				c.Require("unknown", "val")
			},
			wantErr:   true,
			errString: "validator `unknown` not found",
		},
		{
			name: "multiple calls, first fails",
			rules: []Rule{
				{Name: "alpha", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					return errors.New("alpha error")
				})},
				{Name: "beta", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil })},
			},
			callback: func(c Callback) {
				c.Require("alpha", 1)
				c.Optional("beta", "ignored")
			},
			wantErr: true,
		},
		{
			name: "callback with opts",
			rules: []Rule{
				{Name: "inRange", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					if len(opts) != 2 {
						return fmt.Errorf("expected 2 options, got %d", len(opts))
					}
					intVal, ok := value.(int)
					if !ok {
						return fmt.Errorf("expected int, got %T", value)
					}
					minVal, ok := opts[0].(int)
					if !ok {
						return fmt.Errorf("expected int, got %T", value)
					}
					maxVal, ok := opts[1].(int)
					if !ok {
						return fmt.Errorf("expected int, got %T", value)
					}
					if intVal < minVal || intVal > maxVal {
						return fmt.Errorf("not in range")
					}
					return nil
				})},
			},
			callback: func(c Callback) {
				c.Require("inRange", 18, 1, 19)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			if err := v.Register(tt.rules...); err != nil {
				t.Fatalf("Register() error = %v", err)
			}
			err := v.Validate(ctx, tt.callback)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errString != "" && err.Error() != tt.errString {
				t.Errorf("expected error %q, got %q", tt.errString, err.Error())
			}
		})
	}
}
