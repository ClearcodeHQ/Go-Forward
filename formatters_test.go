package main

import (
	"fmt"
	"testing"
)

func TestDefaultFormatter(t *testing.T) {
	m := syslogMessage{
		severity:  logInfo,
		facility:  logAuthpriv,
		hostname:  "debian",
		syslogtag: "sudo:",
		message:   "pam_unix(sudo:session): session closed for user root",
	}
	formated := fmt.Sprintf("%s %s %s %s %s",
		m.facility, m.severity, m.hostname, m.syslogtag, m.message)

	if result := defaultFormatter(m); result != formated {
		t.Errorf("Badly formated message: %q Got: %q", formated, result)
	}
}
