package main

import (
	"fmt"
	"testing"
)

func TestDefaultFormatter(t *testing.T) {
	m := syslogMessage{
		Severity:  logInfo,
		Facility:  logAuthpriv,
		Hostname:  "debian",
		Syslogtag: "sudo:",
		Message:   "pam_unix(sudo:session): session closed for user root",
	}
	formated := fmt.Sprintf("%s %s %s %s %s",
		m.Facility, m.Severity, m.Hostname, m.Syslogtag, m.Message)

	if result := defaultFormatter(m); result != formated {
		t.Errorf("Badly formated message: %q Got: %q", formated, result)
	}
}
