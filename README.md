# zapstackdriver

## Benchmarks

Custom [EncodeEntry](https://godoc.org/github.com/piotrkowalczuk/zapstackdriver#Encoder.EncodeEntry) does not slows down zap significantly.

```
goos: darwin
goarch: amd64
pkg: github.com/piotrkowalczuk/zapstackdriver
BenchmarkEncoder/production-2         	 1000000	      1321 ns/op	     321 B/op	       3 allocs/op
BenchmarkEncoder/development-2        	  200000	      7150 ns/op	     536 B/op	      12 allocs/op
BenchmarkEncoder/stackdriver-2        	 1000000	      1332 ns/op	     325 B/op	       3 allocs/op
PASS
ok  	github.com/piotrkowalczuk/zapstackdriver	4.225s
```