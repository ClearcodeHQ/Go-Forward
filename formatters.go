package main

import "fmt"

type syslogFormatter func(msg syslogMessage) string

func defaultFormatter(msg syslogMessage) string {
	return fmt.Sprintf("%s %s %s %s %s", msg.Facility, msg.Severity, msg.Hostname, msg.Syslogtag, msg.Message)
}

var formatterFunctions = map[string]syslogFormatter{
	"default": defaultFormatter,
}
