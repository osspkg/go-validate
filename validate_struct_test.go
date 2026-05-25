package validate

import (
	"context"
	"errors"
	"testing"
)

func TestValidator_ValidateStruct(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		rules     []Rule
		input     any
		wantErr   bool
		errString string
	}{
		{
			name: "success simple",
			rules: []Rule{
				{Name: "nonempty", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					s, ok := value.(string)
					if !ok || s == "" {
						return errors.New("must be non-empty")
					}
					return nil
				})},
			},
			input: struct {
				Name string `validate:"nonempty"`
			}{Name: "test"},
			wantErr: false,
		},
		{
			name: "required field missing",
			rules: []Rule{
				{Name: "nonempty", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					s, ok := value.(string)
					if !ok || s == "" {
						return errors.New("must be non-empty")
					}
					return nil
				})},
			},
			input: struct {
				Name string `validate:"required;nonempty"`
			}{Name: ""},
			wantErr: true,
		},
		{
			name: "required field zero but optional ok",
			rules: []Rule{
				{Name: "nonempty", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					return nil
				})},
			},
			input: struct {
				Name string `validate:"required;nonempty"`
			}{Name: ""},
			wantErr: false,
		},
		{
			name: "rule not found",
			rules: []Rule{
				{Name: "exists", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil })},
			},
			input: struct {
				Field string `validate:"unknown"`
			}{Field: "test"},
			wantErr:   true,
			errString: "validator `unknown` not found for field `.Field`",
		},
		{
			name:  "non-exported field",
			rules: []Rule{},
			input: struct {
				hidden string `validate:"something"`
			}{},
			wantErr:   true,
			errString: "`hidden` is not exported",
		},
		{
			name: "multiple rules on one field",
			rules: []Rule{
				{Name: "minLen", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					s, _ := value.(string)
					if len(s) < 2 {
						return errors.New("too short")
					}
					return nil
				})},
				{Name: "maxLen", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					s, _ := value.(string)
					if len(s) > 10 {
						return errors.New("too long")
					}
					return nil
				})},
			},
			input: struct {
				Field string `validate:"minLen;maxLen"`
			}{Field: "foo"},
			wantErr: false,
		},
		{
			name: "first rule fails",
			rules: []Rule{
				{Name: "minLen", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					return errors.New("minLen failed")
				})},
				{Name: "maxLen", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil })},
			},
			input: struct {
				Field string `validate:"minLen;maxLen"`
			}{Field: "foo"},
			wantErr: true,
		},
		{
			name:    "nil pointer to struct",
			rules:   []Rule{},
			input:   (*struct{ Name string })(nil),
			wantErr: true,
		},
		{
			name: "pointer to struct",
			rules: []Rule{
				{Name: "nonempty", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
					if value.(string) == "" {
						return errors.New("empty")
					}
					return nil
				})},
			},
			input: &struct {
				Tag string `validate:"nonempty"`
			}{Tag: "ok"},
			wantErr: false,
		},
		{
			name:    "non-struct input",
			rules:   []Rule{},
			input:   42,
			wantErr: true,
		},
		{
			name: "required with zero value but handler returns nil",
			rules: []Rule{
				{Name: "check", Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil })},
			},
			input: struct {
				ID int `validate:"required;check"`
			}{ID: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			if err := v.Register(tt.rules...); err != nil {
				t.Fatalf("Register() error = %v", err)
			}
			err := v.ValidateStruct(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errString != "" && err.Error() != tt.errString {
				t.Errorf("expected error %q, got %q", tt.errString, err.Error())
			}
		})
	}
}
