/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package internal

import (
	"context"
	"fmt"

	"go.osspkg.com/validate"
)

const (
	RuleNameUID validate.Name = "uid"
	RuleNameGID validate.Name = "gid"
)

//govld:gen
func ValidateUID(_ context.Context, value int64, ref int64) error {
	if value <= ref {
		return fmt.Errorf("invalid UID: %d", value)
	}
	return nil
}

type GUID string

func (g *GUID) UnmarshalText(b []byte) error {
	*g = GUID(b)
	return nil
}

//govld:gen
func ValidateGID(_ context.Context, value int64, ref GUID) error {
	if value <= 0 {
		return fmt.Errorf("invalid GID: %d", value)
	}
	return nil
}
