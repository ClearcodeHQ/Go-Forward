package main

import (
	"fmt"
	"testing"
	"time"
)

type TestSyslogPriority struct {
	priority priority
	severity severity
	facility facility
}

var testPriorities = []TestSyslogPriority{
	{severity: logErr, facility: logMail, priority: 19},
	{severity: logEmerg, facility: logKern, priority: 0},
	{severity: logAlert, facility: logUser, priority: 9},
}

func TestSeverityStringUnknown(t *testing.T) {
	sev := severity(254)
	if str := sev.String(); str != "UNKNOWN" {
		t.Errorf("Should return UNKNOWN. Got: %q", str)
	}
}

func TestFacilityStringUnknown(t *testing.T) {
	fac := facility(254)
	if str := fac.String(); str != "UNKNOWN" {
		t.Errorf("Should return UNKNOWN. Got: %q", str)
	}
}

func TestSeverityString(t *testing.T) {
	for sev, val := range severityMap {
		if str := sev.String(); str != val {
			t.Errorf("Should return %q. Got: %q", val, str)
		}
	}
}

func TestFacilityString(t *testing.T) {
	for fac, val := range facilityMap {
		if str := fac.String(); str != val {
			t.Errorf("Should return %q. Got: %q", val, str)
		}
	}
}

func TestDecodeSyslogPriority(t *testing.T) {
	for _, elem := range testPriorities {
		facility, severity := priority.decode(elem.priority)
		if severity != elem.severity {
			t.Errorf("Wrong decoded severity: %v. Should be: %v", severity, elem.severity)
		}
		if facility != elem.facility {
			t.Errorf("Wrong decoded facility: %v. Should be: %v", facility, elem.facility)
		}
	}
}

func TestSyslogMessageString(t *testing.T) {
	m := syslogMessage{
		severity:  logInfo,
		facility:  logAuthpriv,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 970210000, time.UTC),
		hostname:  "debian",
		syslogtag: "sudo:",
		message:   "pam_unix(sudo:session): session closed for user root",
	}
	formated := fmt.Sprintf("FACILITY=%s SEVERITY=%s TIMESTAMP=%s HOSTNAME=%s TAG=%s MESSAGE=%s",
		m.facility, m.severity, m.timestamp, m.hostname, m.syslogtag, m.message)

	if result := m.String(); result != formated {
		t.Errorf("Badly formated message: %q Got: %q", formated, result)
	}
}
