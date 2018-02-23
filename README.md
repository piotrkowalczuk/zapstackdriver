# zapstackdriver [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/zapstackdriver?status.svg)](https://godoc.org/github.com/piotrkowalczuk/zapstackdriver)

## Compatibility

* [LogSeverity](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity) - is mapped from zap log level.
* [HttpRequest](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest) - using [HTTPRequest](https://godoc.org/github.com/piotrkowalczuk/zapstackdriver#HTTPRequest)
* [LogEntryOperation](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogEntryOperation) - through [operation](https://godoc.org/github.com/piotrkowalczuk/zapstackdriver/operation) package interface
* [LogEntrySourceLocation](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogEntrySourceLocation) - it's filled automatically, due to Go limitations function name is missing
* [serviceContext](https://cloud.google.com/error-reporting/docs/formatting-error-messages) - using [ServiceContext](https://godoc.org/github.com/piotrkowalczuk/zapstackdriver#ServiceContext) 

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