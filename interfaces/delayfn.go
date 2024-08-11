package interfaces

type DelayFn func(currenRetries int64) (delay int64)

var (
	// ExponentialBackoffDelayFn is a delay function that implements exponential backoff.
	// It takes the number of retries as input and returns the delay in seconds.
	ExponentialBackoffDelayFn DelayFn = func(currenRetries int64) (delay int64) {
		return 2 << (currenRetries - 1)
	}

	// LinearDelayFn is a delay function that implements linear delay.
	// It takes the number of retries as input and returns the delay in seconds.
	LinearDelayFn DelayFn = func(currenRetries int64) (delay int64) {
		return currenRetries
	}

	// NoDelayFn is a DelayFn implementation that returns 0 delay for retries.
	NoDelayFn DelayFn = func(currenRetries int64) (delay int64) {
		return 0
	}
	DefaultDelayFn DelayFn = LinearDelayFn
)
