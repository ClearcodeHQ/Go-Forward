package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type severity uint8
type facility uint8

type syslogPiority struct {
	facility facility
	severity severity
}

type syslogMessage struct {
	facility  facility
	severity  severity
	message   string
	timestamp time.Time
	syslogtag string
	hostname  string
}

// Used to sort by unix timestamp in ascending order.
type byUnixTimeStamp []syslogMessage

const maxMsgLen = 2048

// From /usr/include/sys/syslog.h.
const (
	logEmerg severity = iota
	logAlert
	logCrit
	logErr
	logWarning
	logNotice
	logInfo
	logDebug
)

var severityMap = map[severity]string{
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
	logKern facility = iota
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

var facilityMap = map[facility]string{
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

var errEmptyMessage = errors.New("Message is empty.")
var errMsgTooLong = fmt.Errorf("Message is too big. Max allowed %d", maxMsgLen)

func (s syslogMessage) String() string {
	return fmt.Sprintf("FACILITY=%d SEVERITY=%d TIMESTAMP=%q HOSTNAME=%q TAG=%q MESSAGE=%q",
		s.facility, s.severity, s.timestamp, s.hostname, s.syslogtag, s.message)
}

func (m byUnixTimeStamp) Len() int           { return len(m) }
func (m byUnixTimeStamp) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byUnixTimeStamp) Less(i, j int) bool { return m[i].timestamp.Unix() < m[j].timestamp.Unix() }

func decodeSyslogPriority(priority uint8) syslogPiority {
	return syslogPiority{facility: facility(priority / 8), severity: severity(priority % 8)}
}

func decodeMessage(msg string) (syslogMessage, error) {
	var priority uint8
	var timestamp string
	if len(msg) >= maxMsgLen {
		return syslogMessage{}, errMsgTooLong
	}
	splited := strings.SplitN(msg, " ", 4)
	header, hname, tag, msg := splited[0], splited[1], splited[2], splited[3]
	msg = strings.Trim(msg, " \n\t")
	if msg == "" {
		return syslogMessage{}, errEmptyMessage
	}

	_, err := fmt.Sscanf(header, "<%d>%s", &priority, &timestamp)
	if err != nil {
		return syslogMessage{}, err
	}

	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return syslogMessage{}, err
	}

	syspri := decodeSyslogPriority(priority)

	return syslogMessage{
		facility:  syspri.facility,
		severity:  syspri.severity,
		message:   msg,
		timestamp: ts,
		syslogtag: tag,
		hostname:  hname,
	}, nil
}
