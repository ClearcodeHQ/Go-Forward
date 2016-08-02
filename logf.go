package main

import (
	"time"
	"strings"
	"fmt"
	"errors"
)

type Severity uint8
type Facility uint8

type SyslogPiority struct {
	facility Facility
	severity Severity
}

type SyslogMessage struct {
	facility Facility
	severity Severity
	message string
	timestamp time.Time
	syslogtag string
	hostname string
}

const MAX_MGS_LEN = 2048

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

var EmptyMessage error = errors.New("Message is empty.")
var MessageTooLong error = fmt.Errorf("Message is too big. Max allowed %d", MAX_MGS_LEN)


func (s SyslogMessage) String() string {
	return fmt.Sprintf("FACILITY=%d SEVERITY=%d TIMESTAMP=%q HOSTNAME=%q TAG=%q MESSAGE=%q",
						s.facility, s.severity, s.timestamp, s.hostname, s.syslogtag, s.message)
}


func decodeSyslogPriority(priority uint8) (SyslogPiority) {
	return SyslogPiority{facility: Facility(priority / 8), severity: Severity(priority % 8)}
}


func encodeSyslogPriority(priority SyslogPiority) (uint8) {
	return uint8(priority.facility * 8) + uint8(priority.severity)
}


func decodeMessage(msg string) (SyslogMessage, error) {
	var priority uint8
	var timestamp string
	if len(msg) >= MAX_MGS_LEN {
		return SyslogMessage{}, MessageTooLong
	}
	splited := strings.SplitN(msg, " ", 4)
	header, hname, tag, msg := splited[0], splited[1], splited[2], splited[3]
	msg = strings.Trim(msg, " \n\t")
	if msg == "" {
		return SyslogMessage{}, EmptyMessage
	}

	_, err := fmt.Sscanf(header, "<%d>%s", &priority, &timestamp)
	if err != nil {
		return SyslogMessage{}, err
	}

	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return SyslogMessage{}, err
	}

	syspri := decodeSyslogPriority(priority)

	return SyslogMessage{
		facility: syspri.facility,
		severity: syspri.severity,
		message: msg,
		timestamp: ts,
		syslogtag: tag,
		hostname: hname,
	}, nil
}
