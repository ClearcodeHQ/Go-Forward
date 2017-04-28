package main

import (
	"bytes"
	"text/template"
	"time"
)

type SyslogSeverity uint8
type SyslogFacility uint8
type SyslogPriority uint8

type syslogMessage struct {
	Facility  SyslogFacility
	Severity  SyslogSeverity
	Message   string
	Syslogtag string
	Hostname  string
	timestamp time.Time
}

const maxMsgLen = 2048

// From /usr/include/sys/syslog.h.
const (
	logEmerg SyslogSeverity = iota
	logAlert
	logCrit
	logErr
	logWarning
	logNotice
	logInfo
	logDebug
)

var severityMap = map[SyslogSeverity]string{
	logEmerg:   "EMERG",
	logAlert:   "ALERT",
	logCrit:    "CRIT",
	logErr:     "ERR",
	logWarning: "WARNING",
	logNotice:  "NOTICE",
	logInfo:    "INFO",
	logDebug:   "DEBUG",
}

// From /usr/include/sys/syslog.h.
const (
	logKern SyslogFacility = iota
	logUser
	logMail
	logDaemon
	logAuth
	logSyslog
	logLpr
	logNews
	logUucp
	logClock
	logAuthpriv
	logFtp
	logNtp
	logLogaudit
	logLogalert
	logCron
	logLocal0
	logLocal1
	logLocal2
	logLocal3
	logLocal4
	logLocal5
	logLocal6
	logLocal7
)

var facilityMap = map[SyslogFacility]string{
	logKern:     "KERN",
	logUser:     "USER",
	logMail:     "MAIL",
	logDaemon:   "DAEMON",
	logAuth:     "AUTH",
	logSyslog:   "SYSLOG",
	logLpr:      "LPR",
	logNews:     "NEWS",
	logUucp:     "UUCP",
	logClock:    "CLOCK",
	logAuthpriv: "AUTHPRIV",
	logFtp:      "FTP",
	logNtp:      "NTP",
	logLogaudit: "LOGAUDIT",
	logLogalert: "LOGALERT",
	logCron:     "CRON",
	logLocal0:   "LOCAL0",
	logLocal1:   "LOCAL1",
	logLocal2:   "LOCAL2",
	logLocal3:   "LOCAL3",
	logLocal4:   "LOCAL4",
	logLocal5:   "LOCAL5",
	logLocal6:   "LOCAL6",
	logLocal7:   "LOCAL7",
}

func (s SyslogSeverity) String() string {
	if val, ok := severityMap[s]; ok {
		return val
	}
	return "UNKNOWN"
}

func (f SyslogFacility) String() string {
	if val, ok := facilityMap[f]; ok {
		return val
	}
	return "UNKNOWN"
}

func (p SyslogPriority) decode() (SyslogFacility, SyslogSeverity) {
	return SyslogFacility(p / 8), SyslogSeverity(p % 8)
}

func (s syslogMessage) render(tpl *template.Template, buf *bytes.Buffer) error {
	buf.Reset()
	return tpl.Execute(buf, s)
}
