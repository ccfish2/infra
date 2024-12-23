package metrics

import "sync"

var (
	metricInitialized   chan struct{} = make(chan struct{})
	flashLoggingMetrics               = sync.Once{}
)

func FlushLoggingMetrics() {
	flashLoggingMetrics.Do(func() {
		if metricInitialized != nil {
			close(metricInitialized)
			metricInitialized = nil
		}
	})
}
