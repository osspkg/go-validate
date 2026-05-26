/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package validate

import (
	"context"
	"fmt"
	"testing"
)

func TestValidator_Register(t *testing.T) {
	v := New()
	if err := v.Register(Rule{}); err == nil {
		t.Fatalf("Register(): empty rule")
	}

	if err := v.Register(
		Rule{
			Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil }),
		}); err == nil {
		t.Fatalf("Register(): rule without name")
	}

	if err := v.Register(
		Rule{
			Name:   "1",
			Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil }),
		},
	); err != nil {
		t.Fatalf("Register(): error = %v", err)
	}

	if err := v.Register(
		Rule{
			Name:   "1",
			Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error { return nil }),
		},
	); err == nil {
		t.Fatalf("Register(): duplicate rule name")
	}
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/validate
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_Validate
Benchmark_Validate-24    	 7162855	       165.8 ns/op	      48 B/op	       3 allocs/op
*/
func Benchmark_Validate(b *testing.B) {
	v := New()
	v.Register(
		Rule{
			Name: "eq",
			Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
				intVal, ok := value.(int)
				if !ok {
					return fmt.Errorf("expected int, got %T", value)
				}
				if len(opts) != 1 {
					return fmt.Errorf("expected 1 options, got %d", len(opts))
				}
				argVal, ok := opts[0].(int)
				if !ok {
					return fmt.Errorf("expected int, got %T", opts[0])
				}
				if intVal != argVal {
					return fmt.Errorf("expected %d, got %d", intVal, argVal)
				}
				return nil
			}),
		},
	)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := v.Validate(ctx, func(cb Callback) {
				cb.Require("eq", 1, 1)
				cb.Require("eq", 2, 2)
				cb.Optional("eq", 0, 2)
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/validate
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_ValidateStruct
Benchmark_ValidateStruct-24    	10004199	       118.1 ns/op	      16 B/op	       2 allocs/op
*/
func Benchmark_ValidateStruct(b *testing.B) {
	v := New()
	v.Register(
		Rule{
			Name: "eq",
			Handle: HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
				intVal, ok := value.(int)
				if !ok {
					return fmt.Errorf("expected int, got %T", value)
				}
				if len(opts) != 1 {
					return fmt.Errorf("expected 1 options, got %d", len(opts))
				}
				var argVal int
				if err := StringDecode(&argVal, opts[0].(string)); err != nil {
					return err
				}
				if intVal != argVal {
					return fmt.Errorf("expected %d, got %d", argVal, intVal)
				}
				return nil
			}),
		},
	)

	type mock struct {
		Id1 int `validate:"required;eq=1"`
		Id2 int `validate:"required;eq=2"`
		Id3 int `validate:"eq=2"`
	}

	mod := &mock{
		Id1: 1,
		Id2: 2,
		Id3: 0,
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := v.ValidateStruct(ctx, mod); err != nil {
				b.Fatal(err)
			}
		}
	})
}
