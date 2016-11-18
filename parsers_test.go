package main

import (
	"testing"
	"time"
)

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

func TestParseMessage(t *testing.T) {
	for _, elem := range testMessages {
		parsed, err := parseRFC3164(elem.raw)
		if err != nil {
			t.Errorf("Error while parsing: %q", err)
		}

		if parsed.Severity != elem.severity {
			t.Errorf("Wrong severity %q. Should be: %q", parsed.Severity, elem.severity)
		}
		if parsed.Facility != elem.facility {
			t.Errorf("Wrong facility %q. Should be: %q", parsed.Facility, elem.facility)
		}
		// Two times can be equal even if they are in different locations.
		// For example, 6:00 +0200 CEST and 4:00 UTC are Equal
		if !parsed.timestamp.Equal(elem.timestamp) {
			t.Errorf("Wrong timestamp %v. Should be: %v", parsed.timestamp, elem.timestamp)
		}
		if parsed.Hostname != elem.hostname {
			t.Errorf("Wrong hostname %q. Should be: %q", parsed.Hostname, elem.hostname)
		}
		if parsed.Syslogtag != elem.tag {
			t.Errorf("Wrong tag %q. Should be: %q", parsed.Syslogtag, elem.tag)
		}
		if parsed.Message != elem.message {
			t.Errorf("Wrong message %q. Should be: %q", parsed.Message, elem.message)
		}
	}
}

func TestEmptyMessage(t *testing.T) {
	emptyMessage := "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]: "
	_, err := parseRFC3164(emptyMessage)
	if err != errEmptyMessage {
		t.Errorf("Should return: %q. Got: %q", errEmptyMessage, err)
	}
}

func TestUnknownMessage(t *testing.T) {
	msg := RandomString(maxMsgLen)
	_, err := parseRFC3164(msg)
	if err != errUnknownMessageFormat {
		t.Errorf("Should return: %q. Got: %q", errUnknownMessageFormat, err)
	}
}
