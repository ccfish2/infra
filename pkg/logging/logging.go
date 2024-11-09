package logging

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"

	"github.com/ccfish2/infra/pkg/logging/logfields"
)

type LogFormat string

const (
	Syslog    = "syslog"
	LevelOpt  = "level"
	FormatOpt = "format"

	LogFormatText          LogFormat = "text"
	LogFormatTextTimestamp LogFormat = "text-ts"
	LogFormatJSON          LogFormat = "json"
	LogFormatJSONTimestamp LogFormat = "json-ts"

	DefaultLogFormat LogFormat = LogFormatText

	DefaultLogFormatTimestamp LogFormat = LogFormatTextTimestamp

	DefaultLogLevel logrus.Level = logrus.InfoLevel
)

var DefaultLogger = initializeDefaultLogger()

func initializeKLog() {
	log := DefaultLogger.WithField(logfields.LogSubsys, "klog")

	klogFlags := flag.NewFlagSet("dolphin", flag.ExitOnError)

	klog.InitFlags(klogFlags)

	klogFlags.Set("logtostderr", "false")

	klogFlags.Set("skip_headers", "true")

	klog.SetOutputBySeverity("INFO", log.WriterLevel(logrus.InfoLevel))
	klog.SetOutputBySeverity("WARNING", log.WriterLevel(logrus.WarnLevel))
	klog.SetOutputBySeverity("ERROR", log.WriterLevel(logrus.ErrorLevel))
	klog.SetOutputBySeverity("FATAL", log.WriterLevel(logrus.FatalLevel))

	klogFlags.Set("one_output", "true")
}

type LogOptions map[string]string

func initializeDefaultLogger() (logger *logrus.Logger) {
	logger = logrus.New()
	logger.SetFormatter(GetFormatter(DefaultLogFormatTimestamp))
	logger.SetLevel(DefaultLogLevel)
	return
}

func (o LogOptions) GetLogLevel() (level logrus.Level) {
	levelOpt, ok := o[LevelOpt]
	if !ok {
		return DefaultLogLevel
	}

	var err error
	if level, err = logrus.ParseLevel(levelOpt); err != nil {
		logrus.WithError(err).Warning("Ignoring user-configured log level")
		return DefaultLogLevel
	}

	return
}

func (o LogOptions) GetLogFormat() LogFormat {
	formatOpt, ok := o[FormatOpt]
	if !ok {
		return DefaultLogFormatTimestamp
	}

	formatOpt = strings.ToLower(formatOpt)
	re := regexp.MustCompile(`^(text|text-ts|json|json-ts)$`)
	if !re.MatchString(formatOpt) {
		logrus.WithError(
			fmt.Errorf("incorrect log format configured '%s', expected ", formatOpt),
		).Warning("Ignoring user-configured log format")
		return DefaultLogFormatTimestamp
	}

	return LogFormat(formatOpt)
}

func SetLogLevel(logLevel logrus.Level) {
	DefaultLogger.SetLevel(logLevel)
}

func SetDefaultLogLevel() {
	DefaultLogger.SetLevel(DefaultLogLevel)
}

func SetLogLevelToDebug() {
	DefaultLogger.SetLevel(logrus.DebugLevel)
}

func SetLogFormat(logFormat LogFormat) {
	DefaultLogger.SetFormatter(GetFormatter(logFormat))
}

func SetDefaultLogFormat() {
	DefaultLogger.SetFormatter(GetFormatter(DefaultLogFormatTimestamp))
}

func AddHooks(hooks ...logrus.Hook) {
	for _, hook := range hooks {
		DefaultLogger.AddHook(hook)
	}
}

func SetupLogging(loggers []string, logOpts LogOptions, tag string, debug bool) error {
	initializeKLog()

	SetLogFormat(logOpts.GetLogFormat())

	if len(loggers) == 0 {
		logrus.SetOutput(os.Stdout)
	}

	if debug {
		SetLogLevelToDebug()
	} else {
		SetLogLevel(logOpts.GetLogLevel())
	}

	logrus.SetLevel(logrus.PanicLevel)

	for _, logger := range loggers {
		switch logger {
		case Syslog:
			if err := setupSyslog(logOpts, tag, debug); err != nil {
				return fmt.Errorf("failed to set up syslog: %w", err)
			}
		default:
			return fmt.Errorf("provided log driver %q is not a supported log driver", logger)
		}
	}

	return nil
}

func GetFormatter(format LogFormat) logrus.Formatter {
	switch format {
	case LogFormatText:
		return &logrus.TextFormatter{
			DisableTimestamp: true,
			DisableColors:    true,
		}
	case LogFormatTextTimestamp:
		return &logrus.TextFormatter{
			DisableTimestamp: false,
			DisableColors:    true,
		}
	case LogFormatJSON:
		return &logrus.JSONFormatter{
			DisableTimestamp: true,
		}
	case LogFormatJSONTimestamp:
		return &logrus.JSONFormatter{
			DisableTimestamp: false,
			TimestampFormat:  time.RFC3339Nano,
		}
	}

	return nil
}

func (o LogOptions) validateOpts(logDriver string, supportedOpts map[string]bool, validKVs map[string][]string) error {
	for k, v := range o {
		if !supportedOpts[k] {
			return fmt.Errorf("provided configuration key %q is not supported as a logging option for log driver %s", k, logDriver)
		}
		if validValues, ok := validKVs[k]; ok {
			valid := false
			for _, vv := range validValues {
				if v == vv {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("provided configuration value %q is not a valid value for %q in log driver %s, valid values: %v", v, k, logDriver, validValues)
			}

		}
	}
	return nil
}

func getLogDriverConfig(logDriver string, logOpts LogOptions) LogOptions {
	keysToValidate := make(LogOptions)
	for k, v := range logOpts {
		ok, err := regexp.MatchString(logDriver+".*", k)
		if err != nil {
			DefaultLogger.Fatal(err)
		}
		if ok {
			keysToValidate[k] = v
		}
	}
	return keysToValidate
}

func MultiLine(logFn func(args ...interface{}), output string) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(output)))
	for scanner.Scan() {
		logFn(scanner.Text())
	}
}

func CanLogAt(logger *logrus.Logger, level logrus.Level) bool {
	return GetLevel(logger) >= level
}

func GetLevel(logger *logrus.Logger) logrus.Level {
	return logrus.Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}
