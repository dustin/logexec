package main

import (
	"errors"
	"log/syslog"
)

var errInvalidLevel = errors.New("invalid log level")
var errInvalidFacility = errors.New("invalid log facility")

var facilityStrings = map[syslog.Priority]string{
	syslog.LOG_KERN:     "kern",
	syslog.LOG_USER:     "user",
	syslog.LOG_MAIL:     "mail",
	syslog.LOG_DAEMON:   "daemon",
	syslog.LOG_AUTH:     "auth",
	syslog.LOG_SYSLOG:   "syslog",
	syslog.LOG_LPR:      "lpr",
	syslog.LOG_NEWS:     "news",
	syslog.LOG_UUCP:     "uucp",
	syslog.LOG_CRON:     "cron",
	syslog.LOG_AUTHPRIV: "authpriv",
	syslog.LOG_FTP:      "ftp",
	syslog.LOG_LOCAL0:   "local0",
	syslog.LOG_LOCAL1:   "local1",
	syslog.LOG_LOCAL2:   "local2",
	syslog.LOG_LOCAL3:   "local3",
	syslog.LOG_LOCAL4:   "local4",
	syslog.LOG_LOCAL5:   "local5",
	syslog.LOG_LOCAL6:   "local6",
	syslog.LOG_LOCAL7:   "local7",
}

var facilityByName = map[string]syslog.Priority{}

func init() {
	for k, v := range facilityStrings {
		facilityByName[v] = k
	}
}

var levelStrings = map[syslog.Priority]string{
	syslog.LOG_EMERG:   "emerg",
	syslog.LOG_ALERT:   "alert",
	syslog.LOG_CRIT:    "crit",
	syslog.LOG_ERR:     "err",
	syslog.LOG_WARNING: "warning",
	syslog.LOG_NOTICE:  "notice",
	syslog.LOG_INFO:    "info",
	syslog.LOG_DEBUG:   "debug",
}

var levelByName = map[string]syslog.Priority{}

func init() {
	for k, v := range levelStrings {
		levelByName[v] = k
	}
}

type logLevel syslog.Priority
type logFacility syslog.Priority

func (l logFacility) String() string {
	return facilityStrings[syslog.Priority(l)&^7]
}

func (l *logFacility) Set(to string) error {
	v, ok := facilityByName[to]
	if !ok {
		return errInvalidFacility
	}
	*l = logFacility(v)
	return nil
}

func (l logLevel) String() string {
	return levelStrings[syslog.Priority(l)&7]
}

func (l *logLevel) Set(to string) error {
	v, ok := levelByName[to]
	if !ok {
		return errInvalidLevel
	}
	*l = logLevel(v)
	return nil
}
