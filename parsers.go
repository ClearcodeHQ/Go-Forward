package main

import (
	"fmt"
	"strings"
	"time"
)

type syslogParser func(msg string) (syslogMessage, error)

var parserFunctions = map[string]syslogParser{
	"RFC3164": parseRFC3164,
}

// https://tools.ietf.org/html/rfc3164
func parseRFC3164(msg string) (parsed syslogMessage, err error) {
	pri_end_index := strings.Index(msg, ">")
	if pri_end_index == -1 {
		err = errUnknownMessageFormat
		return
	}
	// Priority string length can be at most 4 chars long.
	if pri_end_index > 3 {
		err = errUnknownMessageFormat
		return
	}
	pri_end_index += 1
	priority := msg[0:pri_end_index]
	var pri SyslogPriority
	_, err = fmt.Sscanf(priority, "<%d>", &pri)
	if err != nil {
		err = errUnknownMessageFormat
		return
	}

	if len(msg[pri_end_index:]) < pri_end_index+len(time.Stamp) {
		err = errUnknownMessageFormat
		return
	}
	ts := msg[pri_end_index:(pri_end_index + len(time.Stamp))]
	timestamp, err := parseRFC3164Timestamp(ts)
	if err != nil {
		return
	}
	parsed.timestamp = timestamp

	splitted := strings.SplitN(msg[(pri_end_index+len(time.Stamp)+1):], " ", 3)
	hostname, syslog_tag, message := splitted[0], splitted[1], splitted[2]

	message = strings.Trim(message, " \n\t")
	if message == "" {
		err = errEmptyMessage
		return
	}
	parsed.Message = msg
	parsed.Syslogtag = syslog_tag
	parsed.Hostname = hostname

	fac, sev := pri.decode()
	parsed.Facility = fac
	parsed.Severity = sev

	return
}

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func parseRFC3164Timestamp(timestamp string) (ts time.Time, err error) {
	ts, err = time.Parse(time.Stamp, timestamp)
	if err != nil {
		return
	}
	now := time.Now()
	ts = time.Date(now.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(),
		ts.Second(), ts.Nanosecond(), ts.Location())
	return
}
