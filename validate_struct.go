/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package validate

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"go.osspkg.com/validate/internal/cache"
)

const tagName = "validate"

type structFieldInfo struct {
	Index       int
	Name        string
	ParentIndex int
	HasParent   bool
	Required    bool
	Tags        []tagInfo
}

type structValidator struct {
	cache *cache.Cache[reflect.Type, []structFieldInfo]
	mux   sync.Mutex
}

func newStructValidator() *structValidator {
	return &structValidator{
		cache: cache.New[reflect.Type, []structFieldInfo](),
	}
}

func (v *structValidator) run(ctx context.Context, r resolver, arg any) error {
	ref, err := v.refStruct(arg)
	if err != nil {
		return err
	}

	fields, err := v.getStructInfo(ref.Type())
	if err != nil {
		return err
	}

	for i := 0; i < len(fields); i++ {
		field := &fields[i]

		var refField reflect.Value

		if field.HasParent {
			refField = ref.Field(field.ParentIndex)
			if refField.Kind() == reflect.Ptr {
				if refField.IsNil() {
					if field.Required {
						return fmt.Errorf("got nullible reference to required field `%s`", field.Name)
					}
					continue
				}
				refField = refField.Elem()
			}
			refField = refField.Field(field.Index)
		} else {
			refField = ref.Field(field.Index)
		}

		if !field.Required && (!refField.IsValid() || refField.IsZero()) {
			continue
		}

		var value any
		if refField.Kind() == reflect.Ptr {
			if refField.IsNil() {
				value = nil
			} else {
				value = refField.Elem().Interface()
			}
		} else {
			value = refField.Interface()
		}

		for j := 0; j < len(field.Tags); j++ {
			tag := &field.Tags[j]

			rule, ok := r.Resolve(tag.Name)
			if !ok {
				return fmt.Errorf("validator `%s` not found for field `%s`", tag.Name, field.Name)
			}

			if err = rule.Handle.ValidateHandle(ctx, value, tag.Opts...); err != nil {
				return fmt.Errorf("validate field `%s`: %w", field.Name, err)
			}
		}
	}

	return nil
}

func (v *structValidator) getStructInfo(ref reflect.Type) ([]structFieldInfo, error) {
	sfi, ok := v.cache.Get(ref)
	if ok {
		return sfi, nil
	}

	v.mux.Lock()
	defer v.mux.Unlock()

	sfi, ok = v.cache.Get(ref)
	if ok {
		return sfi, nil
	}

	return v.resolveStructInfo(ref)
}

func (v *structValidator) resolveStructInfo(ref reflect.Type) ([]structFieldInfo, error) {
	sfi, ok := v.cache.Get(ref)
	if ok {
		return sfi, nil
	}

	sfi = make([]structFieldInfo, 0, ref.NumField())

	for i := 0; i < ref.NumField(); i++ {

		field := ref.Field(i)

		list, req, ok := v.resolveTag(field.Tag)
		if !ok {
			ft := field.Type

			if ft.Kind() == reflect.Ptr {
				ft = field.Type.Elem()
			}

			if ft.Kind() == reflect.Struct {
				sub, err := v.resolveStructInfo(ft)
				if err != nil {
					return nil, err
				}

				for _, info := range sub {
					item := structFieldInfo{
						Name:        info.Name,
						Index:       info.Index,
						ParentIndex: i,
						HasParent:   true,
						Required:    info.Required,
						Tags:        info.Tags,
					}
					sfi = append(sfi, item)
				}
			}

			continue
		}

		if !field.IsExported() {
			return nil, fmt.Errorf("`%s` is not exported", field.Name)
		}

		sfi = append(sfi, structFieldInfo{
			Name:     ref.Name() + "." + field.Name,
			Index:    i,
			Required: req,
			Tags:     list,
		})
	}

	v.cache.Set(ref, sfi)

	return sfi, nil
}

func (v *structValidator) resolveTag(l tagLookup) ([]tagInfo, bool, bool) {
	valueStr, ok := l.Lookup(tagName)
	if !ok {
		return nil, false, false
	}

	var required bool
	tags := make([]tagInfo, 0, 2)

	for item := range strings.SplitSeq(valueStr, ";") {
		switch item {
		case "required":
			required = true
			continue

		case "":
			continue

		default:
			if len(item) == 0 {
				continue
			}

			ti := tagInfo{}

			inx := strings.IndexByte(item, '=')
			if inx < 0 {
				ti.Name = Name(item)
			} else {
				ti.Name = Name(item[:inx])

				for s := range strings.SplitSeq(item[inx+1:], ",") {
					ti.Opts = append(ti.Opts, s)
				}
			}

			tags = append(tags, ti)
		}
	}

	return tags, required, true
}

func (v *structValidator) refStruct(arg any) (reflect.Value, error) {
	ref := reflect.ValueOf(arg)

	for ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			return reflect.Value{}, fmt.Errorf("got nil-pointer object")
		}
		ref = ref.Elem()
	}

	if ref.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("got non-struct object")
	}

	return ref, nil
}
