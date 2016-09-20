package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type TestSyslogPriority struct {
	priority priority
	severity severity
	facility facility
}

type TestMessage struct {
	raw string
	// Parsed expected fields
	severity  severity
	facility  facility
	timestamp time.Time
	hostname  string
	tag       string
	message   string
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

var testPriorities = []TestSyslogPriority{
	{severity: logErr, facility: logMail, priority: 19},
	{severity: logEmerg, facility: logKern, priority: 0},
	{severity: logAlert, facility: logUser, priority: 9},
}

var testMessages = []TestMessage{
	{
		raw:       "<86>2016-07-23T14:48:16.970210+02:00 debian sudo: pam_unix(sudo:session): session closed for user root",
		severity:  logInfo,
		facility:  logAuthpriv,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 970210000, time.UTC),
		hostname:  "debian",
		tag:       "sudo:",
		message:   "pam_unix(sudo:session): session closed for user root",
	},
	{
		raw:       "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]: \tpam_unix(su:session): session closed for user root\n",
		severity:  logInfo,
		facility:  logAuthpriv,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC),
		hostname:  "debian",
		tag:       "su[2106]:",
		message:   "pam_unix(su:session): session closed for user root",
	},
	{
		raw:       "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]:  pam_unix(su:session): session closed for user root \n",
		severity:  logInfo,
		facility:  logAuthpriv,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC),
		hostname:  "debian",
		tag:       "su[2106]:",
		message:   "pam_unix(su:session): session closed for user root",
	},
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

func TestParseMessage(t *testing.T) {
	for _, elem := range testMessages {
		parsed, err := decodeMessage(elem.raw)
		if err != nil {
			t.Errorf("Error while parsing: %q", err)
		}

		if parsed.severity != elem.severity {
			t.Errorf("Wrong severity %q. Should be: %q", parsed.severity, elem.severity)
		}
		if parsed.facility != elem.facility {
			t.Errorf("Wrong facility %q. Should be: %q", parsed.facility, elem.facility)
		}
		// Two times can be equal even if they are in different locations.
		// For example, 6:00 +0200 CEST and 4:00 UTC are Equal
		if !parsed.timestamp.Equal(elem.timestamp) {
			t.Errorf("Wrong timestamp %v. Should be: %v", parsed.timestamp, elem.timestamp)
		}
		if parsed.hostname != elem.hostname {
			t.Errorf("Wrong hostname %q. Should be: %q", parsed.hostname, elem.hostname)
		}
		if parsed.syslogtag != elem.tag {
			t.Errorf("Wrong tag %q. Should be: %q", parsed.syslogtag, elem.tag)
		}
		if parsed.message != elem.message {
			t.Errorf("Wrong message %q. Should be: %q", parsed.message, elem.message)
		}
	}
}

func TestEmptyMessage(t *testing.T) {
	emptyMessage := "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]: "
	_, err := decodeMessage(emptyMessage)
	if err != errEmptyMessage {
		t.Errorf("Should return: %q. Got: %q", errEmptyMessage, err)
	}
}

func TestUnknownMessage(t *testing.T) {
	msg := RandomString(maxMsgLen)
	_, err := decodeMessage(msg)
	if err != errUnknownMessageFormat {
		t.Errorf("Should return: %q. Got: %q", errUnknownMessageFormat, err)
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
