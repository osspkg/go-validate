/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package validate

import (
	"context"
	"fmt"
	"reflect"
)

var (
	ctxType = reflect.TypeFor[context.Context]()
	errType = reflect.TypeFor[error]()
)

const requireArgs = 2

func adaptError(err error) HandlerFunc {
	return func(ctx context.Context, value any, opts ...any) error {
		return err
	}
}

func AdaptHandlerFunc(fn any) HandlerFunc {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return adaptError(fmt.Errorf("AdaptHandlerFunc: expects a function"))
	}

	numIn := fnType.NumIn()
	if numIn < requireArgs {
		return adaptError(fmt.Errorf("AdaptHandlerFunc: function must have at least 2 arguments (ctx, value)"))
	}

	if !fnType.In(0).Implements(ctxType) {
		return adaptError(fmt.Errorf("AdaptHandlerFunc: first argument must implement context.Context"))
	}

	if fnType.NumOut() != 1 || !fnType.Out(0).Implements(errType) {
		return adaptError(fmt.Errorf("AdaptHandlerFunc: function must return exactly one value of type error"))
	}

	targetValType := fnType.In(1)

	targetOptTypes := make([]reflect.Type, 0, numIn-2)
	for i := 0; i < numIn-2; i++ {
		targetOptTypes = append(targetOptTypes, fnType.In(i+2))
	}

	return func(ctx context.Context, value any, opts ...any) error {
		if len(opts) < numIn-2 {
			return fmt.Errorf("AdaptHandlerFunc: missing arguments, expected %d options, got %d", numIn-2, len(opts))
		}

		args := make([]reflect.Value, numIn)
		args[0] = reflect.ValueOf(ctx)

		if value == nil {
			args[1] = reflect.Zero(targetValType)
		} else {
			rVal, err := castReflect(value, targetValType)
			if err != nil {
				return fmt.Errorf("AdaptHandlerFunc: value type mismatch error: %w", err)
			}
			if rValRef, ok := rVal.(reflect.Value); ok {
				args[1] = rValRef
			} else {
				args[1] = reflect.ValueOf(rVal)
			}
		}

		for i, targetOptType := range targetOptTypes {
			rVal, err := castReflect(opts[i], targetOptType)
			if err != nil {
				return fmt.Errorf("AdaptHandlerFunc: option %d error: %w", i, err)
			}
			if rValRef, ok := rVal.(reflect.Value); ok {
				args[i+2] = rValRef
			} else {
				args[i+2] = reflect.ValueOf(rVal)
			}
		}

		res := fnVal.Call(args)
		if !res[0].IsNil() {
			return res[0].Interface().(error)
		}

		return nil
	}
}

func castReflect(val any, target reflect.Type) (any, error) {
	var err error
	if s, ok := val.(string); ok {
		switch target.Kind() {
		case reflect.String:
			return s, nil

		case reflect.Int:
			var v int
			err = StringDecode(&v, s)
			return v, err
		case reflect.Int8:
			var v int8
			err = StringDecode(&v, s)
			return v, err
		case reflect.Int16:
			var v int16
			err = StringDecode(&v, s)
			return v, err
		case reflect.Int32:
			var v int32
			err = StringDecode(&v, s)
			return v, err
		case reflect.Int64:
			var v int64
			err = StringDecode(&v, s)
			return v, err

		case reflect.Uint:
			var v uint
			err = StringDecode(&v, s)
			return v, err
		case reflect.Uint8:
			var v uint8
			err = StringDecode(&v, s)
			return v, err
		case reflect.Uint16:
			var v uint16
			err = StringDecode(&v, s)
			return v, err
		case reflect.Uint32:
			var v uint32
			err = StringDecode(&v, s)
			return v, err
		case reflect.Uint64:
			var v uint64
			err = StringDecode(&v, s)
			return v, err

		case reflect.Float32:
			var v float32
			err = StringDecode(&v, s)
			return v, err
		case reflect.Float64:
			var v float64
			err = StringDecode(&v, s)
			return v, err

		case reflect.Bool:
			var v bool
			err = StringDecode(&v, s)
			return v, err

		default:
		}
	}

	v := reflect.ValueOf(val)
	if v.Type() == target {
		return v, nil
	}

	if v.Type().ConvertibleTo(target) {
		return v.Convert(target), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %v", val, target)
}
