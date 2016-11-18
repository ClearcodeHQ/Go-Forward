package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestSyslogPriority struct {
	priority priority
	severity severity
	facility facility
}

func TestSeverityStringUnknown(t *testing.T) {
	sev := severity(254)
	assert.Equal(t, "UNKNOWN", sev.String())
}

func TestFacilityStringUnknown(t *testing.T) {
	fac := facility(254)
	assert.Equal(t, "UNKNOWN", fac.String())
}

func TestSeverityString(t *testing.T) {
	for sev, val := range severityMap {
		assert.Equal(t, val, sev.String())
	}
}

func TestFacilityString(t *testing.T) {
	for fac, val := range facilityMap {
		assert.Equal(t, val, fac.String())
	}
}

func TestDecodeSyslogPriority_severity(t *testing.T) {
	for _, elem := range testPriorities {
		_, severity := priority.decode(elem.priority)
		assert.Equal(t, severity, elem.severity)
	}
}

func TestDecodeSyslogPriority_facility(t *testing.T) {
	for _, elem := range testPriorities {
		facility, _ := priority.decode(elem.priority)
		assert.Equal(t, facility, elem.facility)
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
