package main

import (
	"strconv"
	"strings"
	"time"
)

type syslogParser func(msg string) (syslogMessage, error)

var parserFunctions = map[string]syslogParser{
	"RFC3164": parseRFC3164,
}

// https://tools.ietf.org/html/rfc3164
func parseRFC3164(str string) (parsed syslogMessage, err error) {
	str = strings.Replace(str, "<", "", 1)
	str = strings.Replace(str, ">", " ", 1)
	str = strings.Replace(str, "  ", " ", 1)
	strs := strings.SplitN(str, " ", 7)
	if len(strs) != 7 {
		err = errUnknownMessageFormat
		return
	}

	priority, err := strconv.Atoi(strs[0])
	if err != nil {
		return
	}
	fac, sev := SyslogPriority(priority).decode()
	parsed.Facility = fac
	parsed.Severity = sev

	date := strings.Join(strs[1:4], " ")
	timestamp, err := parseRFC3164Timestamp(date)
	if err != nil {
		return
	}
	parsed.timestamp = timestamp

	parsed.Message = strings.TrimSpace(strs[6])
	if parsed.Message == "" {
		err = errEmptyMessage
		return
	}

	parsed.Syslogtag = strs[5]
	parsed.Hostname = strs[4]

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
