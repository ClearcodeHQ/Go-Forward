package logf

import (
	"testing"
)

type TestSyslogPriority struct {
	priority byte
	severity Severity
	facility Facility
}

var tests = []TestSyslogPriority {
	TestSyslogPriority{severity: 3, facility: 2, priority: 19},
	TestSyslogPriority{severity: 0, facility: 0, priority: 0},
	TestSyslogPriority{severity: 1, facility: 1, priority: 9},
}

func TestDecodeSyslogPriority(t *testing.T) {

	for i := range tests {
		decoded := decodeSyslogPriority(tests[i].priority)
		if decoded.severity != tests[i].severity {
			t.Errorf("Wrong decoded severity: %d. Should be: %d.", decoded.severity, tests[i].severity)
		}
		if decoded.facility != tests[i].facility {
			t.Errorf("Wrong decoded facility: %d. Should be: %d.", decoded.facility, tests[i].facility)
		}
	}
}

func TestEncodeSyslogPriority(t *testing.T) {

	for i := range tests {
		pri := SyslogPiority{severity: tests[i].severity, facility: tests[i].facility}
		encoded := encodeSyslogPriority(pri)
		if encoded != tests[i].priority {
			t.Errorf("Wrong encoded priority %d. Should be: %d", encoded, tests[i].priority)
		}
	}
}
