package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseMessage_nil_error(t *testing.T) {
	_, err := parseRFC3164(testMessage.raw)
	assert.Nil(t, err)
}

func Test_ParseMessage_Severity(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.severity, parsed.Severity)
}

func Test_ParseMessage_Facility(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	assert.Equal(t, testMessage.facility, parsed.Facility)
}

func Test_ParseMessage_timestamp(t *testing.T) {
	parsed, _ := parseRFC3164(testMessage.raw)
	if !parsed.timestamp.Equal(testMessage.timestamp) {
		t.Errorf("Wrong timestamp %v. Should be: %v", parsed.timestamp, testMessage.timestamp)
	}
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
	msg := RandomString(maxMsgLen)
	_, err := parseRFC3164(msg)
	assert.Equal(t, errUnknownMessageFormat, err)
}

func Test_valid_parseRFC3339(t *testing.T) {
	str := "2016-07-23T14:48:16.969683+02:00"
	_, err := parseRFC3339(str)
	assert.Nil(t, err)
}

func Test_invalid_parseRFC3339(t *testing.T) {
	str := "201-07-23T14:48:16.969683+02:00"
	_, err := parseRFC3339(str)
	assert.NotNil(t, err)
}
