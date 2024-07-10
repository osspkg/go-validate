/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package validate

import (
	"fmt"
	"testing"
)

func TestUnit_GetDomainLevel(t *testing.T) {
	type args struct {
		s     string
		level int
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args: args{
				s:     "www.domain.ltd",
				level: 1,
			},
			want: "ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd",
				level: 2,
			},
			want: "domain.ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd",
				level: 10,
			},
			want: "www.domain.ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd.",
				level: 1,
			},
			want: "ltd.",
		},
		{
			args: args{
				s:     "ltd.",
				level: 3,
			},
			want: "ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd.",
				level: 0,
			},
			want: ".",
		},
		{
			args: args{
				s:     "",
				level: 0,
			},
			want: ".",
		},
		{
			args: args{
				s:     "a a  a",
				level: 1,
			},
			want: ".",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			if got := GetDomainLevel(tt.args.s, tt.args.level); got != tt.want {
				t.Errorf("GetDomainLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_GetDomainLevel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			GetDomainLevel("www.domain.ltd.", 2)
		}
	})
}

func Benchmark_IsValidDomain(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			IsValidDomain("www.domain.ltd.")
		}
	})
}

func Benchmark_NormalizeDomainBytes(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			NormalizeDomainBytes([]byte("  WWW.Domain.ltd    "))
		}
	})
}

func Benchmark_NormalizeDomain(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			NormalizeDomain("  WWW.Domain.ltd    ")
		}
	})
}

func TestUnit_NormalizeDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		want    string
		wantErr bool
	}{
		{
			name:    "Case1",
			domain:  "1www.A-aa.com",
			want:    "1www.a-aa.com.",
			wantErr: false,
		},
		{
			name:    "Case2",
			domain:  "1_www.aaa.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Case3",
			domain:  "com",
			want:    "com.",
			wantErr: false,
		},
		{
			name:    "Case4",
			domain:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Case5",
			domain:  "a",
			want:    "a.",
			wantErr: false,
		},
		{
			name:    "Case6",
			domain:  "Ф",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Case6.1",
			domain:  " a aaaaaa",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Case6.2",
			domain:  "aaaaaa a ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Case7",
			domain:  " a ",
			want:    "a.",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeDomain() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnit_CountDomainLevels(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "Case1",
			arg:  "",
			want: 0,
		},
		{
			name: "Case2",
			arg:  "aaa.",
			want: 1,
		},
		{
			name: "Case3",
			arg:  "aaa.bbb.",
			want: 2,
		},
		{
			name: "Case4",
			arg:  "aaa.bbb",
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountDomainLevels(tt.arg); got != tt.want {
				t.Errorf("CountDomainLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}
