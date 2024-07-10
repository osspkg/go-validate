/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package validate

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

var dot = byte('.')

func GetDomainLevel(s string, level int) string {
	if level == 0 {
		return "."
	}
	var err error
	s, err = NormalizeDomain(s)
	if err != nil {
		return "."
	}
	maxPos := len(s) - 1
	count, pos := 0, 0
	if s[maxPos] == dot {
		maxPos--
	}

	for i := maxPos; i >= 0; i-- {
		if s[i] == dot {
			count++
			if count == level {
				pos = i + 1
				break
			}
		}
	}
	return s[pos:]
}

func CountDomainLevels(s string) int {
	ss, err := NormalizeDomain(s)
	if err != nil {
		return 0
	}
	return strings.Count(ss, ".")
}

func IsValidDomain(s string) bool {
	d, b := 0, []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] == '.' {
			d++
		} else {
			d = 0
		}
		switch true {
		case b[i] >= utf8.RuneSelf || d > 1:
			return false
		case b[i] == '.', b[i] == '-',
			'a' <= b[i] && b[i] <= 'z',
			'0' <= b[i] && b[i] <= '9',
			'A' <= b[i] && b[i] <= 'Z':
			continue
		default:
			return false
		}
	}
	return true
}

func NormalizeDomain(s string) (string, error) {
	b, err := NormalizeDomainBytes([]byte(s))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func NormalizeDomainBytes(b []byte) ([]byte, error) {
	if len(b) < 1 {
		return nil, fmt.Errorf("invalid domain")
	}
	d := 0
	for i := 0; i < len(b); i++ {
		if b[i] == '.' {
			d++
		} else {
			d = 0
		}
		switch true {
		case b[i] >= utf8.RuneSelf || d > 1:
			return nil, fmt.Errorf("invalid domain")
		case b[i] == '.', b[i] == '-',
			'a' <= b[i] && b[i] <= 'z',
			'0' <= b[i] && b[i] <= '9':
			continue
		case 'A' <= b[i] && b[i] <= 'Z':
			b[i] += 'a' - 'A'
		default:
			b[i] = ' '
		}
	}
	f, t, m := 0, len(b), len(b)-1
	for i := 0; i < (len(b)+1)/2; i++ {
		if b[i] == ' ' {
			if i-f > 0 {
				return nil, fmt.Errorf("invalid domain")
			}
			f++
		}
		if b[m-i] == ' ' {
			if m-i-t < -1 {
				return nil, fmt.Errorf("invalid domain")
			}
			t--
		}
	}
	if b[t-1] != '.' {
		if t <= m {
			b[t] = '.'
		} else {
			b = append(b, '.')
		}
		t++
	}
	return b[f:t], nil
}
