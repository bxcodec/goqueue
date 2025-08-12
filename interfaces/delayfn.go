package interfaces

// DelayFn is a function type that represents a delay function.
// It takes the current number of retries as input and returns the delay in seconds.
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
	NoDelayFn DelayFn = func(_ int64) (delay int64) {
		return 0
	}
	// DefaultDelayFn is the default delay function that will be used if no delay function is provided.
	// It is set to LinearDelayFn by default.
	DefaultDelayFn DelayFn = LinearDelayFn
)
