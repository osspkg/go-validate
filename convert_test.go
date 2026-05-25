package validate

import (
	"testing"
)

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/validate
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_ConvertFloat
Benchmark_ConvertFloat-24    	260549199	         4.410 ns/op	       8 B/op	       1 allocs/op
*/
func Benchmark_ConvertFloat(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var f float64
			if err := StringDecode(&f, "3.141592653589793"); err != nil {
				b.Fatal(err)
			}
		}
	})
}
