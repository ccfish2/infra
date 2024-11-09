package logging

import (
	"log/syslog"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus"
)

const (
	SLevel    = "syslog.level"
	SNetwork  = "syslog.network"
	SAddress  = "syslog.address"
	SSeverity = "syslog.severity"
	SFacility = "syslog.facility"
	STag      = "syslog.tag"
)

var (
	syslogOpts = map[string]bool{
		SLevel:    true,
		SNetwork:  true,
		SAddress:  true,
		SSeverity: true,
		SFacility: true,
		STag:      true,
	}

	syslogSeverityMap = map[string]syslog.Priority{
		"emerg":   syslog.LOG_EMERG,
		"panic":   syslog.LOG_EMERG,
		"alert":   syslog.LOG_ALERT,
		"crit":    syslog.LOG_CRIT,
		"err":     syslog.LOG_ERR,
		"error":   syslog.LOG_ERR,
		"warn":    syslog.LOG_WARNING,
		"warning": syslog.LOG_WARNING,
		"notice":  syslog.LOG_NOTICE,
		"info":    syslog.LOG_INFO,
		"debug":   syslog.LOG_DEBUG,
	}

	syslogFacilityMap = map[string]syslog.Priority{
		"kern":     syslog.LOG_KERN,
		"user":     syslog.LOG_USER,
		"mail":     syslog.LOG_MAIL,
		"daemon":   syslog.LOG_DAEMON,
		"auth":     syslog.LOG_AUTH,
		"syslog":   syslog.LOG_SYSLOG,
		"lpr":      syslog.LOG_LPR,
		"news":     syslog.LOG_NEWS,
		"uucp":     syslog.LOG_UUCP,
		"cron":     syslog.LOG_CRON,
		"authpriv": syslog.LOG_AUTHPRIV,
		"ftp":      syslog.LOG_FTP,
		"local0":   syslog.LOG_LOCAL0,
		"local1":   syslog.LOG_LOCAL1,
		"local2":   syslog.LOG_LOCAL2,
		"local3":   syslog.LOG_LOCAL3,
		"local4":   syslog.LOG_LOCAL4,
		"local5":   syslog.LOG_LOCAL5,
		"local6":   syslog.LOG_LOCAL6,
		"local7":   syslog.LOG_LOCAL7,
	}
	syslogLevelMap = map[logrus.Level]syslog.Priority{
		logrus.PanicLevel: syslog.LOG_ALERT,
		logrus.FatalLevel: syslog.LOG_CRIT,
		logrus.ErrorLevel: syslog.LOG_ERR,
		logrus.WarnLevel:  syslog.LOG_WARNING,
		logrus.InfoLevel:  syslog.LOG_INFO,
		logrus.DebugLevel: syslog.LOG_DEBUG,
		logrus.TraceLevel: syslog.LOG_DEBUG,
	}
)

func mapStringPriorityToSlice(m map[string]syslog.Priority) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

func setupSyslog(logOpts LogOptions, tag string, debug bool) error {
	opts := getLogDriverConfig(Syslog, logOpts)
	syslogOptValues := make(map[string][]string)
	syslogOptValues[SSeverity] = mapStringPriorityToSlice(syslogSeverityMap)
	syslogOptValues[SFacility] = mapStringPriorityToSlice(syslogFacilityMap)
	if err := opts.validateOpts(Syslog, syslogOpts, syslogOptValues); err != nil {
		return err
	}
	if stag, ok := opts[STag]; ok {
		tag = stag
	}

	logLevel, ok := opts[SLevel]
	if !ok {
		if debug {
			logLevel = "debug"
		} else {
			logLevel = "info"
		}
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		DefaultLogger.Fatal(err)
	}

	SetLogLevel(level)

	network := ""
	address := ""
	severity := syslogLevelMap[level]
	facility := syslog.LOG_KERN
	if networkStr, ok := opts[SNetwork]; ok {
		network = networkStr
	}
	if addressStr, ok := opts[SAddress]; ok {
		address = addressStr
	}
	if severityStr, ok := opts[SSeverity]; ok {
		severity = syslogSeverityMap[severityStr]
	}
	if facilityStr, ok := opts[SFacility]; ok {
		facility = syslogFacilityMap[facilityStr]
	}

	h, err := logrus_syslog.NewSyslogHook(network, address, severity|facility, tag)
	if err != nil {
		DefaultLogger.Fatal(err)
	}
	logrus.AddHook(h)
	DefaultLogger.AddHook(h)

	return nil
}
