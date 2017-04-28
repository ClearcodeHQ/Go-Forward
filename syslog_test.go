package main

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

type TestSyslogPriority struct {
	priority SyslogPriority
	severity SyslogSeverity
	facility SyslogFacility
}

func TestSeverityStringUnknown(t *testing.T) {
	sev := SyslogSeverity(254)
	assert.Equal(t, "UNKNOWN", sev.String())
}

func TestFacilityStringUnknown(t *testing.T) {
	fac := SyslogFacility(254)
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
		_, severity := SyslogPriority.decode(elem.priority)
		assert.Equal(t, severity, elem.severity)
	}
}

func TestDecodeSyslogPriority_facility(t *testing.T) {
	for _, elem := range testPriorities {
		facility, _ := SyslogPriority.decode(elem.priority)
		assert.Equal(t, facility, elem.facility)
	}
}

// Assert that message is rendered with correct fields order
func Test_syslogMessage_render_order(t *testing.T) {
	m := syslogMessage{
		Syslogtag: "tag",
		Message:   "message",
	}
	buf := bytes.NewBuffer([]byte{})
	tpl, _ := template.New("").Parse("{{.Message}} {{.Syslogtag}}")
	m.render(tpl, buf)
	assert.Equal(t, "message tag", buf.String())
}

// Assert that every exported field is rendered
func Test_syslogMessage_render_field(t *testing.T) {
	m := syslogMessage{
		Severity:  logInfo,
		Facility:  logAuthpriv,
		Hostname:  "hostname",
		Syslogtag: "tag",
		Message:   "message",
	}
	for field, expected := range map[string]string{
		"Severity":  m.Severity.String(),
		"Facility":  m.Facility.String(),
		"Hostname":  m.Hostname,
		"Syslogtag": m.Syslogtag,
		"Message":   m.Message,
	} {
		buf := bytes.NewBuffer([]byte{})
		tpl, _ := template.New("").Parse(fmt.Sprintf("{{.%s}}", field))
		m.render(tpl, buf)
		assert.Equal(t, expected, buf.String())
	}
}

func Test_syslogMessage_render_error(t *testing.T) {
	m := syslogMessage{}
	buf := bytes.NewBuffer([]byte{})
	tpl, _ := template.New("").Parse("{{.UnexportedField}}")
	assert.NotNil(t, m.render(tpl, buf))
}

// Assert that buffer is reset before rendering template
func Test_syslogMessage_render_reset(t *testing.T) {
	m := syslogMessage{}
	buf := bytes.NewBuffer([]byte("should not be rendered"))
	tpl, _ := template.New("").Parse("")
	m.render(tpl, buf)
	assert.Equal(t, "", buf.String())
}
