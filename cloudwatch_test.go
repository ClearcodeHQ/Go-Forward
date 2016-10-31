package main

import (
	"testing"
)

func TestEventSize(t *testing.T) {
	event := logEvent{msg: "123", timestamp: 123}
	expected := 3 + eventSizeOverhead
	if event.size() != expected {
		t.Errorf("Event size shoud be %v. Got: %v", expected, event.size())
	}
}

func TestEventValidateTooBig(t *testing.T) {
	event := logEvent{msg: RandomString(maxEventSize + 1)}
	if err := event.validate(); err != errMessageTooBig {
		t.Errorf("Should return %q. Got: %q", errMessageTooBig, err)
	}
}

func TestDestinationString(t *testing.T) {
	dst := destination{group: "group", stream: "stream"}
	expected := "group: group stream: stream"
	if str := dst.String(); str != expected {
		t.Errorf("Should return '%s'. Got: '%s'", expected, str)
	}
}
