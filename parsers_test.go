package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse_RFC3164_nil_error(t *testing.T) {
	msg := "<86>Jul 23 14:48:16 debian sudo: pam_unix(sudo:session): session closed for user root"
	_, err := parseRFC3164(msg)
	assert.Nil(t, err)
}

func Test_Parse_RFC3164_Severity(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.severity, parsed.Severity)
}

func Test_Parse_RFC3164_Facility(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.facility, parsed.Facility)
}

func Test_Parse_RFC3164Timestamp(t *testing.T) {
	msg := "Jul 23 14:48:16"
	ts, _ := parseRFC3164Timestamp(msg)
	assert.Equal(t, 7, ts.Month())
}

func Test_ParseMessage_Hostname(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.hostname, parsed.Hostname)
}

func Test_ParseMessage_Syslogtag(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.syslogTag, parsed.Syslogtag)
}

func Test_ParseMessage_Message(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.message, parsed.Message)
}

func Test_parseRFC3164_empty(t *testing.T) {
	emptyMessage := "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]: "
	_, err := parseRFC3164(emptyMessage)
	assert.Equal(t, errEmptyMessage, err)
}

func TestUnknownMessage(t *testing.T) {
	bad_messages := []string{
		"kfjlsdkfdlsjdlfgkdlsfghsdlfgkh",
		"<888>dsfdsfdsgsgd",
		"<aa>bla bla#@$@#4",
		"<84>bla bla#@$@#4",
	}
	for _, msg := range bad_messages {
		_, err := parseRFC3164(msg)
		assert.Equal(t, errUnknownMessageFormat, err)
	}
}
