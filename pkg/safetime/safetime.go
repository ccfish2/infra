package safetime

import (
	"runtime"
	"time"

	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/sirupsen/logrus"
)

func TimeSinceSafe(t time.Time, logger *logrus.Entry) (time.Duration, bool) {
	n := time.Now()
	d := n.Sub(t)

	if d < 0 {
		logger.Warn("Invalid time value")

		_, file, line, ok := runtime.Caller(1)
		if ok {
			logger = logger.WithFields(logrus.Fields{
				logfields.Path: file,
				logfields.Line: line,
			})
		}
		logger.Warn("BUG: negative duration")
		return time.Duration(0), false
	}

	return d, true
}
