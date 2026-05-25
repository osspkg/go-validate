/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package validate

import "context"

type Validator struct {
	store     *store
	structVld *structValidator
}

func New() *Validator {
	return &Validator{
		store:     newStore(),
		structVld: newStructValidator(),
	}
}

func (v *Validator) Register(rules ...Rule) error {
	return v.store.Append(rules...)
}

func (v *Validator) ValidateStruct(ctx context.Context, arg any) error {
	return v.structVld.run(ctx, v.store, arg)
}

func (v *Validator) Validate(ctx context.Context, call func(c Callback)) error {
	cb := poolCallbackValidator.Get()
	defer poolCallbackValidator.Put(cb)

	call(cb)

	return cb.run(ctx, v.store)
}
