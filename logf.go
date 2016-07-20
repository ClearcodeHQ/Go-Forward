package logf

type Severity uint8
type Facility uint8

type SyslogPiority struct {
	facility Facility
	severity Severity
}

const (
	// From /usr/include/sys/syslog.h.
	LOG_EMERG Severity = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

const (
	// From /usr/include/sys/syslog.h.
	LOG_KERN Facility = iota
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CLOCK
	LOG_AUTHPRIV
	LOG_FTP
	LOG_NTP
	LOG_LOGAUDIT
	LOG_LOGALERT
	LOG_CRON
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

func decodeSyslogPriority(priority byte) (SyslogPiority) {
	return SyslogPiority{facility: Facility(priority / 8), severity: Severity(priority % 8)}
}

func encodeSyslogPriority(priority SyslogPiority) (byte) {
	return byte(priority.facility * 8) + byte(priority.severity)
}
