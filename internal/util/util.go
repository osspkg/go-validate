/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package util

import "reflect"

func IsDefaultValue(arg any) bool {
	switch v := arg.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case bool:
		return !v
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	default:
		rv := reflect.ValueOf(arg)
		if !rv.IsValid() {
			return true
		}
		return rv.IsZero()
	}
}
