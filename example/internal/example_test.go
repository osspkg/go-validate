package internal

import (
	"context"
	"testing"

	"go.osspkg.com/validate"
)

func TestExample1(t *testing.T) {
	vld := validate.New()

	failIfError(t,
		vld.Register(
			validate.Rule{
				Name:   RuleNameUID,
				Handle: validate.HandlerFunc(ValidateUIDAdaptHandler),
			},
			validate.Rule{
				Name:   RuleNameGID,
				Handle: validate.HandlerFunc(ValidateGIDAdaptHandler),
			},
		),
		"Register()")

	failIfError(t,
		vld.Validate(context.TODO(), func(c validate.Callback) {
			c.Require(RuleNameUID, "123", int64(10))
		}),
		"Validate()",
	)

	type Demo struct {
		UserID int64 `json:"user_id" validate:"required;uid=10;gid=100"`
	}

	failIfError(t,
		vld.ValidateStruct(context.TODO(), &Demo{
			UserID: 123,
		}),
		"ValidateStruct()",
	)
}

func failIfError(t testing.TB, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

func BenchmarkExample1(b *testing.B) {
	vld := validate.New()

	type Demo struct {
		UserID int64 `json:"user_id" validate:"required;uid=10;gid=100"`
	}

	model := &Demo{UserID: 123}
	ctx := context.Background()

	failIfError(b,
		vld.Register(
			validate.Rule{
				Name:   RuleNameUID,
				Handle: validate.HandlerFunc(ValidateUIDAdaptHandler),
			},
			validate.Rule{
				Name:   RuleNameGID,
				Handle: validate.HandlerFunc(ValidateGIDAdaptHandler),
			},
		),
		"Register()")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			failIfError(b,
				vld.Validate(ctx, func(c validate.Callback) {
					c.Optional(RuleNameUID, "", int64(1000))
				}),
				"Validate()",
			)

			failIfError(b, vld.ValidateStruct(ctx, model), "ValidateStruct()")
		}
	})

}
