package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/ccfish2/infra/pkg/logging/logfields"
)

const logrErrorKey = "err"

var slogHandlerOpts = &slog.HandlerOptions{
	AddSource:   false,
	Level:       slogLeveler,
	ReplaceAttr: replaceAttrFn,
}

var slogLeveler = func() *slog.LevelVar {
	var levelVar slog.LevelVar
	levelVar.Set(slog.LevelInfo)
	return &levelVar
}()

var defaultMultiSlogHandler = NewMultiSlogHandler(slog.NewTextHandler(
	os.Stderr,
	slogHandlerOpts,
))

var DefaultSlogLogger = slog.New(defaultMultiSlogHandler)

func replaceAttrFn(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		// Adjust to timestamp format that logrus uses; except that we can't
		// force slog to quote the value like logrus does...
		return slog.String(slog.TimeKey, a.Value.Time().Format(time.RFC3339Nano))
	case slog.LevelKey:
		switch level := a.Value; {
		case level.Equal(levelFatalValue):
			return slog.Attr{
				Key:   slog.LevelKey,
				Value: slog.StringValue("fatal"),
			}
		case level.Equal(levelPanicValue):
			return slog.Attr{
				Key:   slog.LevelKey,
				Value: slog.StringValue("panic"),
			}
		}
		// Lower-case the log level
		return slog.Attr{
			Key:   slog.LevelKey,
			Value: slog.StringValue(strings.ToLower(a.Value.String())),
		}
	case logrErrorKey:
		// Uniform the attribute identifying the error
		return slog.Attr{
			Key:   logfields.Error,
			Value: a.Value,
		}
	}
	return a
}
