[![Go Report Card](https://goreportcard.com/badge/github.com/liangyaopei/bloom)](https://goreportcard.com/report/github.com/liangyaopei/bloom)
[![GoDoc](https://godoc.org/github.com/liangyaopei/bloom?status.svg)](http://godoc.org/github.com/liangyaopei/bloom)
# High-performance and user-friendly bloom filter in Golang
This library provides bloom filter which has high performance and is concurrent safe. 

## Install
```
go get -u github.com/liangyaopei/bloom
```

## Usage
Users only need to specify the parameter of `m`,`k`,`race` and then can use it.
Example.
```go
    filter := bloom.New(1024, 3, false)
	// chaining operation
	filter.Add([]byte("Hello")).
		AddString("World").
		AddUInt16(uint16(16)).
		AddUInt32(uint32(32)).
		AddUInt64(uint64(64)).
		AddUint16Batch([]uint16{1, 2, 3})

	t.Logf("Hello exist:%t", filter.Test([]byte("Hello")))
	t.Logf("World exist:%t", filter.TestString("World"))
	t.Logf("uint 16 exist:%t", filter.TestUInt16(uint16(16)))
	t.Logf("uint 32 exist:%t", filter.TestUInt32(uint32(32)))
	t.Logf("uint 64 exist:%t", filter.TestUInt64(uint64(64)))
```

## Performance: high performance with low false positive rate
In `_bloom_test`, the `TestFalsePositive` function demonstrates that after inserting 1 million random string key, whose string length is 30 each, this library has 
a false positive ratio of 2.82%(0.028200).

