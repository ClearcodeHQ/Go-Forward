package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 970210000, time.UTC),
		Severity:  logInfo,
		Facility:  logAuthpriv,
		Hostname:  "debian",
		Syslogtag: "sudo:",
		Message:   "pam_unix(sudo:session): session closed for user root",
	}
	formated := fmt.Sprintf("FACILITY=%s SEVERITY=%s TIMESTAMP=%s HOSTNAME=%s TAG=%s MESSAGE=%s",
		m.Facility, m.Severity, m.timestamp, m.Hostname, m.Syslogtag, m.Message)

	if result := m.String(); result != formated {
		t.Errorf("Badly formated message: %q Got: %q", formated, result)
	}
}

// Assert that message is rendered with correct fields order
func Test_syslogMessage_render_order(t *testing.T) {
	m := syslogMessage{
		Syslogtag: "tag",
		Message:   "message",
	}
	actual, _ := m.render("{{.Message}} {{.Syslogtag}}")
	assert.Equal(t, "message tag", actual)
}

func Test_syslogMessage_render_Severity(t *testing.T) {
	m := syslogMessage{
		Severity: logInfo,
	}
	expected := fmt.Sprintf("%s", m.Severity)
	actual, _ := m.render("{{.Severity}}")
	assert.Equal(t, expected, actual)
}

func Test_syslogMessage_render_Facility(t *testing.T) {
	m := syslogMessage{
		Facility: logAuthpriv,
	}
	expected := fmt.Sprintf("%s", m.Facility)
	actual, _ := m.render("{{.Facility}}")
	assert.Equal(t, expected, actual)
}

func Test_syslogMessage_render_Hostname(t *testing.T) {
	m := syslogMessage{
		Hostname: "hostname",
	}
	actual, _ := m.render("{{.Hostname}}")
	assert.Equal(t, "hostname", actual)
}

func Test_syslogMessage_render_Syslogtag(t *testing.T) {
	m := syslogMessage{
		Syslogtag: "tag",
	}
	actual, _ := m.render("{{.Syslogtag}}")
	assert.Equal(t, "tag", actual)
}

func Test_syslogMessage_render_Message(t *testing.T) {
	m := syslogMessage{
		Message: "message",
	}
	actual, _ := m.render("{{.Message}}")
	assert.Equal(t, "message", actual)
}

func Test_syslogMessage_render_error(t *testing.T) {
	m := syslogMessage{}
	_, err := m.render("{{.UnexportedField}}")
	assert.NotNil(t, err)
}
