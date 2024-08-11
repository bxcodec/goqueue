package interfaces

import "fmt"

type DelayFn func(retries int64) (delay int64)

var (
	// ExponentialBackoffDelayFn is a delay function that implements exponential backoff.
	// It takes the number of retries as input and returns the delay in seconds.
	ExponentialBackoffDelayFn DelayFn = func(retries int64) (delay int64) {
		fmt.Println(">>>>> ExponentialBackoffDelayFn, retry: ", retries)
		return 2 << (retries - 1)
	}

	// LinearDelayFn is a delay function that implements linear delay.
	// It takes the number of retries as input and returns the delay in seconds.
	LinearDelayFn DelayFn = func(retries int64) (delay int64) {
		fmt.Println(">>>>> LinearDelayFn, retry: ", retries)
		return retries
	}

	// NoDelayFn is a DelayFn implementation that returns 0 delay for retries.
	NoDelayFn DelayFn = func(retries int64) (delay int64) {
		return 0
	}
	DefaultDelayFn DelayFn = LinearDelayFn
)
