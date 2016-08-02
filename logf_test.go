package main

import (
	"testing"
	"time"
	"math/rand"
)

type TestSyslogPriority struct {
	priority uint8
	severity Severity
	facility Facility
}

type TestMessage struct {
	raw string
	// Parsed expected fields
	severity Severity
	facility Facility
	timestamp time.Time
	hostname string
	tag string
	message string
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

var test_priorities = []TestSyslogPriority {
	TestSyslogPriority{severity: LOG_ERR, facility: LOG_MAIL, priority: 19},
	TestSyslogPriority{severity: LOG_EMERG, facility: LOG_KERN, priority: 0},
	TestSyslogPriority{severity: LOG_ALERT, facility: LOG_USER, priority: 9},
}

var test_messages = []TestMessage {
	TestMessage{
		raw:  "<86>2016-07-23T14:48:16.970210+02:00 debian sudo: pam_unix(sudo:session): session closed for user root",
		severity: LOG_INFO,
		facility: LOG_AUTHPRIV,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 970210000, time.UTC),
		hostname: "debian",
		tag: "sudo:",
		message: "pam_unix(sudo:session): session closed for user root",
	},
	TestMessage{
		raw:  "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]: \tpam_unix(su:session): session closed for user root\n",
		severity: LOG_INFO,
		facility: LOG_AUTHPRIV,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC),
		hostname: "debian",
		tag: "su[2106]:",
		message: "pam_unix(su:session): session closed for user root",
	},
	TestMessage{
		raw:  "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]:  pam_unix(su:session): session closed for user root \n",
		severity: LOG_INFO,
		facility: LOG_AUTHPRIV,
		timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC),
		hostname: "debian",
		tag: "su[2106]:",
		message: "pam_unix(su:session): session closed for user root",
	},
}

var empty_message = "<86>2016-07-23T14:48:16.969683+02:00 debian su[2106]: "


func TestDecodeSyslogPriority(t *testing.T) {

	for _, elem := range test_priorities {
		decoded := decodeSyslogPriority(elem.priority)
		if decoded.severity != elem.severity {
			t.Errorf("Wrong decoded severity: %v. Should be: %v", decoded.severity, elem.severity)
		}
		if decoded.facility != elem.facility {
			t.Errorf("Wrong decoded facility: %v. Should be: %v", decoded.facility, elem.facility)
		}
	}
}


func TestEncodeSyslogPriority(t *testing.T) {

	for _, elem := range test_priorities {
		pri := SyslogPiority{severity: elem.severity, facility: elem.facility}
		encoded := encodeSyslogPriority(pri)
		if encoded != elem.priority {
			t.Errorf("Wrong encoded priority %v. Should be: %v", encoded, elem.priority)
		}
	}
}


func TestParseMessage(t *testing.T) {
	for _, elem := range test_messages {
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
	_, err := decodeMessage(empty_message)
	if err != EmptyMessage {
		t.Errorf("Should return: %q. Got: %q", err)
	}
}


func TestMessageTooLong(t *testing.T) {
	msg := RandomString(MAX_MGS_LEN)
	_, err := decodeMessage(msg)
	if err != MessageTooLong {
		t.Errorf("Should return: %q. Got: %q", err)
	}
}
