package resource

type ErrorAction string

var (
	ErrorActionRetry  ErrorAction = "retry"
	ErrorActionIgnore ErrorAction = "ignore"
	ErrorActionStop   ErrorAction = "stop"
)

type ErrorHandler func(key Key, numRetries int, err error) ErrorAction

func AlwaysRetry(Key, int, error) ErrorAction {
	return ErrorActionRetry
}

func RetryUpTo(n int) ErrorHandler {
	return func(key Key, numRetries int, err error) ErrorAction {
		if numRetries >= n {
			return ErrorActionStop
		}
		return ErrorActionRetry
	}
}
