package validate

import (
	"context"
	"fmt"

	"go.osspkg.com/validate/internal/pool"
	"go.osspkg.com/validate/internal/util"
)

type Callback interface {
	Optional(name Name, value any, opts ...any)
	Require(name Name, value any, opts ...any)
}

var poolCallbackValidator = pool.New[*callbackValidator](func() *callbackValidator {
	return &callbackValidator{params: make([]cvParam, 0, 32)}
})

type cvParam struct {
	require bool
	name    Name
	value   any
	opts    []any
}

type callbackValidator struct {
	params []cvParam
}

func (v *callbackValidator) Reset() {
	v.params = v.params[:0]
}

func (v *callbackValidator) Optional(name Name, value any, opts ...any) {
	v.params = append(v.params, cvParam{require: false, name: name, value: value, opts: opts})
}

func (v *callbackValidator) Require(name Name, value any, opts ...any) {
	v.params = append(v.params, cvParam{require: true, name: name, value: value, opts: opts})
}

func (v *callbackValidator) handler(ctx context.Context, r resolver, p cvParam) error {
	rule, ok := r.Resolve(p.name)
	if !ok {
		return fmt.Errorf("validator `%s` not found", p.name)
	}

	if !p.require && util.IsDefaultValue(p.value) {
		return nil
	}

	return rule.Handle.ValidateHandle(ctx, p.value, p.opts...)
}

func (v *callbackValidator) run(ctx context.Context, r resolver) error {
	for i := range v.params {
		if err := v.handler(ctx, r, v.params[i]); err != nil {
			return err
		}
	}
	return nil
}
