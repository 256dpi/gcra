// Package gcra implements the generic cell rate algorithm (GCRA).
package gcra

import (
	"errors"
	"math"
	"time"
)

// ErrInvalidParameter is returned if a parameter is zero.
var ErrInvalidParameter = errors.New("invalid parameter")

// ErrCostHigherThanBurst is returned if the provided cost is higher than the
// specified burst.
var ErrCostHigherThanBurst = errors.New("cost higher than burst")

// Bucket represents a GCRA bucket. The value represents the theoretical arrival
// time (TAT) which encodes the point in time at which the bucket is full again.
type Bucket time.Time

// Options define the GCRA options. Specify burst as the maximum tokens
// available and rate as the regeneration of tokens per period.
type Options struct {
	Burst  int64
	Rate   int64
	Period time.Duration
}

// Result is the result of a GCRA computation.
type Result struct {
	Limited   bool
	Remaining int64
	RetryIn   time.Duration
	ResetIn   time.Duration
}

// Generate will create a bucket that contains the specified amount of tokens
// at this point in time.
func Generate(now time.Time, count int64, opts Options) (Bucket, error) {
	// check arguments
	if count < 0 || opts.Burst <= 0 || opts.Rate <= 0 || opts.Period <= 0 {
		return Bucket{}, ErrInvalidParameter
	} else if count > opts.Burst {
		return Bucket{}, ErrCostHigherThanBurst
	}

	// calculate TAT
	tat := GenerateRaw(now.UnixNano(), count, opts.Burst, opts.Rate, int64(opts.Period))

	// create bucket
	bucket := Bucket(time.Unix(0, tat))

	return bucket, nil
}

// MustGenerate will call Generate and panic on errors.
func MustGenerate(now time.Time, count int64, opts Options) Bucket {
	bucket, err := Generate(now, count, opts)
	if err != nil {
		panic(err)
	}
	return bucket
}

// Compute will perform the GCRA. Cost may be zero to query the bucket.
func Compute(now time.Time, bucket Bucket, cost int64, opts Options) (Bucket, Result, error) {
	// check arguments
	if cost < 0 || opts.Burst <= 0 || opts.Rate <= 0 || opts.Period <= 0 {
		return bucket, Result{}, ErrInvalidParameter
	} else if cost > opts.Burst {
		return bucket, Result{}, ErrCostHigherThanBurst
	}

	// calculate TAT
	tat := time.Time(bucket).UnixNano()

	// compute GCRA
	newTAT, limited, remaining, retryIn, resetIn := ComputeRaw(tat, now.UnixNano(), opts.Burst, opts.Rate, int64(opts.Period), cost)

	// update bucket
	bucket = Bucket(time.Unix(0, newTAT))

	// prepare result
	result := Result{
		Limited:   limited,
		Remaining: remaining,
		RetryIn:   time.Duration(retryIn),
		ResetIn:   time.Duration(resetIn),
	}

	return bucket, result, nil
}

// MustCompute will call Compute and panic on errors.
func MustCompute(now time.Time, bucket Bucket, cost int64, opts Options) (Bucket, Result) {
	bucket, result, err := Compute(now, bucket, cost, opts)
	if err != nil {
		panic(err)
	}
	return bucket, result
}

// GenerateRaw is the underlying raw computation used in Generate.
func GenerateRaw(now, count, burst, rate, period int64) int64 {
	// compute variables
	emissionInterval := roundDiv(period, rate)

	// compute tat
	tat := now + emissionInterval*(burst-count)

	return tat
}

// ComputeRaw us the underlying raw computation used in Compute.
func ComputeRaw(tat, now, burst, rate, period, cost int64) (int64, bool, int64, int64, int64) {
	// compute variables
	emissionInterval := roundDiv(period, rate)
	increment := emissionInterval * cost
	burstOffset := emissionInterval * burst

	// reset TAT if smaller than now
	if now > tat {
		tat = now
	}

	// calculate new TAT
	newTAT := tat + increment

	// compute allow at
	allowAt := newTAT - burstOffset

	// compute difference
	diff := now - allowAt

	// compute remaining
	remaining := roundDiv(diff, emissionInterval)

	// check if not enough
	if remaining < 0 {
		remaining := roundDiv(now-(tat-burstOffset), emissionInterval)
		resetIn := tat - now
		retryIn := diff * -1
		return tat, true, remaining, retryIn, resetIn
	}

	// check if empty
	if remaining == 0 && increment <= 0 {
		resetIn := tat - now
		return tat, true, 0, 0, resetIn
	}

	// calculate reset in
	resetIn := newTAT - now

	return newTAT, false, remaining, 0, resetIn
}

func roundDiv(a, b int64) int64 {
	return int64(math.Round(float64(a) / float64(b)))
}
