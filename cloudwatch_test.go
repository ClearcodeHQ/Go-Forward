package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_logEvent_size(t *testing.T) {
	event := logEvent{msg: "123", timestamp: 123}
	expected := 3 + eventSizeOverhead
	assert.Equal(t, expected, event.size())
}

func Test_logEvent_tooBig(t *testing.T) {
	event := logEvent{msg: RandomString(maxEventSize + 1)}
	assert.Equal(t, errMessageTooBig, event.validate())
}

func Test_destination_string(t *testing.T) {
	dst := destination{group: "group", stream: "stream"}
	assert.Equal(t, "group: group stream: stream", dst.String())
}
