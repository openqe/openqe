package utils

import (
	"errors"
	"fmt"
	"time"
)

const (
	DoubleLongTimeout = 1800 * time.Second // 30 minutes
	LongTimeout       = 900 * time.Second  // 15 minutes
	DefaultTimeout    = 300 * time.Second  // 5 minutes
	ShortTimeout      = 60 * time.Second   // 1 minute

	ShortInterval      = 5 * time.Second  // 5 seconds
	DefaultInterval    = 10 * time.Second // 10 seconds
	LongInterval       = 30 * time.Second // 30 seconds
	DoubleLongInterval = 60 * time.Second // 1 minute
)

// Supplier is a function that produces a value of type T or an error.
type Supplier[T any] func() (T, error)

// Predicate is a function that checks if a value of type T meets some condition.
type Predicate[T any] func(T) bool

func EventuallyDefault[T any](
	supplier Supplier[T],
	predicate Predicate[T],
) (T, error) {
	return Eventually(supplier, predicate, DefaultInterval, DefaultTimeout)
}

func EventuallyShort[T any](
	supplier Supplier[T],
	predicate Predicate[T],
) (T, error) {
	return Eventually(supplier, predicate, ShortInterval, ShortTimeout)
}

// Eventually repeatedly calls `supplier` until `predicate` returns true
// or the timeout is reached.
//
// - supplier: the function to call each interval (returns T, error)
// - predicate: checks if the result is acceptable
// - interval: how often to retry
// - timeout: maximum wait time
//
// It returns the last value from supplier if the predicate succeeds,
// or error if it fails after timeout.
func Eventually[T any](
	supplier Supplier[T],
	predicate Predicate[T],
	interval time.Duration,
	timeout time.Duration,
) (T, error) {
	var last T
	deadline := time.Now().Add(timeout)
	var next T
	var err error
	for time.Now().Before(deadline) {
		next, err = supplier()
		if err != nil {
			// shall we continue even on error !?
			return last, fmt.Errorf("eventually: supplier failed: %w", err)
		}
		if predicate(next) {
			return next, nil
		}
		last = next
		time.Sleep(interval)
	}
	return last, errors.New("eventually: condition not met within timeout")
}
