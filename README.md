# gcra

[![Test](https://github.com/256dpi/gcra/actions/workflows/test.yml/badge.svg)](https://github.com/256dpi/gcra/actions/workflows/test.yml)
[![GoDoc](https://godoc.org/github.com/256dpi/gcra?status.svg)](http://godoc.org/github.com/256dpi/gcra)
[![Release](https://img.shields.io/github/release/256dpi/gcra.svg)](https://github.com/256dpi/gcra/releases)

**Package gcra implements the generic cell rate algorithm.**

## Example

```go
opts := Options{
    Burst:  50,
    Rate:   10,
    Period: time.Second,
}

now := time.Now()

bucket := MustGenerate(now, 25, opts)

bucket, result := MustCompute(now, bucket, 10, opts)
fmt.Printf("%+v\n", result)

bucket, result = MustCompute(now, bucket, 30, opts)
fmt.Printf("%+v\n", result)

bucket, result = MustCompute(now, bucket, 15, opts)
fmt.Printf("%+v\n", result)

now = now.Add(2 * time.Second)

bucket, result = MustCompute(now, bucket, 0, opts)
fmt.Printf("%+v\n", result)

// Output:
// {Limited:false Remaining:15 RetryIn:0s ResetIn:3.5s}
// {Limited:true Remaining:15 RetryIn:1.5s ResetIn:3.5s}
// {Limited:false Remaining:0 RetryIn:0s ResetIn:5s}
// {Limited:false Remaining:20 RetryIn:0s ResetIn:3s}
```