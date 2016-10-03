package main

import "fmt"

type syslogFormatter func(msg syslogMessage) string

func defaultFormatter(msg syslogMessage) string {
	return fmt.Sprintf("%s %s %s %s %s", msg.facility, msg.severity, msg.hostname, msg.syslogtag, msg.message)
}

var formatterFunctions = map[string]syslogFormatter{
	"default": defaultFormatter,
}
