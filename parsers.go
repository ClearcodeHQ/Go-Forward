package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var errEmptyMessage = errors.New("Message is empty.")

type syslogParser func(msg string) (syslogMessage, error)

func parseRFC3164(msg string) (parsed syslogMessage, err error) {
	var pri priority
	var timestamp string
	var ts time.Time
	splited := strings.SplitN(msg, " ", 4)
	if len(splited) != 4 {
		err = errUnknownMessageFormat
		return
	}
	header, hname, tag, msg := splited[0], splited[1], splited[2], splited[3]
	msg = strings.Trim(msg, " \n\t")
	if msg == "" {
		err = errEmptyMessage
		return
	}

	_, err = fmt.Sscanf(header, "<%d>%s", &pri, &timestamp)
	if err != nil {
		return
	}

	ts, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return
	}

	fac, sev := pri.decode()

	parsed = syslogMessage{
		facility:  fac,
		severity:  sev,
		message:   msg,
		timestamp: ts,
		syslogtag: tag,
		hostname:  hname,
	}
	return
}

var parserFunctions = map[string]syslogParser{
	"RFC3339": parseRFC3164,
}
