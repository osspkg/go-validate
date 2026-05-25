package validate

import (
	"context"
	"fmt"
	"testing"
)

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/validate
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_ValidateStruct_WithAdapt
Benchmark_ValidateStruct_WithAdapt-24    	 6333072	       188.0 ns/op	     376 B/op	      17 allocs/op
*/
func Benchmark_ValidateStruct_WithAdapt(b *testing.B) {
	v := New()
	v.Register(
		Rule{
			Name: "eq",
			Handle: AdaptHandlerFunc(func(ctx context.Context, value int64, reference int64) error {
				if value != reference {
					return fmt.Errorf("expected %d, got %d", reference, value)
				}
				return nil
			}),
		},
		Rule{
			Name: "eq2",
			Handle: AdaptHandlerFunc(func(ctx context.Context, value, reference float64) error {
				if value != reference {
					return fmt.Errorf("expected %v, got %v", reference, value)
				}
				return nil
			}),
		},
	)

	type mock struct {
		Id1 int `validate:"required;eq=1"`
		Id2 int `validate:"required;eq2=2"`
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
