package goqu

type DelayFn func(retries int64) (delay int64)

var (
	// ExponentialBackoffDelayFn is a delay function that implements exponential backoff.
	// It takes the number of retries as input and returns the delay in milliseconds.
	ExponentialBackoffDelayFn DelayFn = func(retries int64) (delay int64) {
		return 2 << retries
	}
	// NoDelayFn is a DelayFn implementation that returns 0 delay for retries.
	NoDelayFn DelayFn = func(retries int64) (delay int64) {
		return 0
	}
	DefaultDelayFn DelayFn = NoDelayFn
)
