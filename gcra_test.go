package gcra

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Example() {
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
}

var now = time.Date(2022, 1, 23, 10, 52, 0, 0, time.UTC)

func TestGenerate(t *testing.T) {
	opts := Options{
		Burst:  100,
		Rate:   10,
		Period: time.Second,
	}

	bucket := MustGenerate(now, 75, opts)
	bucket, result := MustCompute(now, bucket, 0, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 75,
		RetryIn:   0,
		ResetIn:   2500 * time.Millisecond,
	}, result)

	bucket = MustGenerate(now, 20, opts)
	bucket, result = MustCompute(now, bucket, 0, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 20,
		RetryIn:   0,
		ResetIn:   8 * time.Second,
	}, result)

	bucket = MustGenerate(now, 50, opts)
	bucket, result = MustCompute(now, bucket, 0, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 50,
		RetryIn:   0,
		ResetIn:   5 * time.Second,
	}, result)
}

func TestCompute(t *testing.T) {
	opts := Options{
		Burst:  4,
		Rate:   10,
		Period: 10 * time.Second,
	}

	bucket, result := MustCompute(now, Bucket{}, 1, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 3,
		RetryIn:   0,
		ResetIn:   1 * time.Second,
	}, result)

	bucket, result = MustCompute(now, bucket, 1, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 2,
		RetryIn:   0,
		ResetIn:   2 * time.Second,
	}, result)

	bucket, result = MustCompute(now, bucket, 1, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 1,
		RetryIn:   0,
		ResetIn:   3 * time.Second,
	}, result)

	bucket, result = MustCompute(now, bucket, 1, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 0,
		RetryIn:   0,
		ResetIn:   4 * time.Second,
	}, result)

	bucket, result = MustCompute(now, bucket, 1, opts)
	assert.Equal(t, Result{
		Limited:   true,
		Remaining: 0,
		RetryIn:   1 * time.Second,
		ResetIn:   4 * time.Second,
	}, result)

	bucket, result = MustCompute(now, bucket, 0, opts)
	assert.Equal(t, Result{
		Limited:   true,
		Remaining: 0,
		RetryIn:   0,
		ResetIn:   4 * time.Second,
	}, result)

	now = now.Add(2 * time.Second)

	bucket, result = MustCompute(now, bucket, 1, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 1,
		RetryIn:   0,
		ResetIn:   3 * time.Second,
	}, result)

	now = now.Add(time.Second)

	bucket, result = MustCompute(now, bucket, 1, opts)
	assert.Equal(t, Result{
		Limited:   false,
		Remaining: 1,
		RetryIn:   0,
		ResetIn:   3 * time.Second,
	}, result)

	bucket, result = MustCompute(now, bucket, 2, opts)
	assert.Equal(t, Result{
		Limited:   true,
		Remaining: 1,
		RetryIn:   1 * time.Second,
		ResetIn:   3 * time.Second,
	}, result)
}

func TestGenerateErrors(t *testing.T) {
	_, err := Generate(now, 0, Options{Burst: 0, Rate: 1, Period: 1})
	assert.Equal(t, ErrZeroParameter, err)

	_, err = Generate(now, 0, Options{Burst: 1, Rate: 0, Period: 1})
	assert.Equal(t, ErrZeroParameter, err)

	_, err = Generate(now, 0, Options{Burst: 1, Rate: 1, Period: 0})
	assert.Equal(t, ErrZeroParameter, err)

	_, err = Generate(now, 2, Options{Burst: 1, Rate: 1, Period: 1})
	assert.Equal(t, ErrCostHigherThanBurst, err)

	assert.Panics(t, func() {
		MustGenerate(now, 0, Options{Burst: 0, Rate: 1, Period: 1})
	})
}

func TestComputeErrors(t *testing.T) {
	_, _, err := Compute(now, Bucket{}, 1, Options{0, 1, 1})
	assert.Equal(t, ErrZeroParameter, err)

	_, _, err = Compute(now, Bucket{}, 1, Options{1, 0, 1})
	assert.Equal(t, ErrZeroParameter, err)

	_, _, err = Compute(now, Bucket{}, 1, Options{1, 1, 0})
	assert.Equal(t, ErrZeroParameter, err)

	_, _, err = Compute(now, Bucket{}, 2, Options{1, 1, 1})
	assert.Equal(t, ErrCostHigherThanBurst, err)

	assert.Panics(t, func() {
		MustCompute(now, Bucket{}, 1, Options{0, 1, 1})
	})
}

func BenchmarkCompute(b *testing.B) {
	opts := Options{
		Burst:  int64(b.N),
		Rate:   10,
		Period: 10 * time.Second,
	}

	b.ReportAllocs()
	b.ResetTimer()

	var bucket Bucket
	for i := 0; i < b.N; i++ {
		bucket, _ = MustCompute(now, bucket, 1, opts)
	}
}
