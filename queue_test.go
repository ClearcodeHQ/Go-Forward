package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_numEvents(t *testing.T) {
	cases := []numPair{
		numPair{expected: 100, passed: 100},
		numPair{expected: maxBatchEvents, passed: maxBatchEvents * 2},
	}
	for _, pair := range cases {
		if result := numEvents(pair.passed); result != pair.expected {
			t.Errorf("Should return %d. Got: %d", pair.expected, result)
		}
	}
}

func Test_queue_empty(t *testing.T) {
	queue := new(eventQueue)
	assert.True(t, queue.empty())
}

func Test_queue_not_empty(t *testing.T) {
	queue := new(eventQueue)
	queue.add(logEvent{})
	assert.False(t, queue.empty())
}

func Test_queue_length(t *testing.T) {
	queue := new(eventQueue)
	queue.add(logEvent{})
	assert.Equal(t, 1, queue.num())
}

// Assert that event is added at the end of slice.
func Test_queue_add(t *testing.T) {
	queue := new(eventQueue)
	queue.add(logEvent{msg: "first"})
	queue.add(logEvent{msg: "second"})
	queue.add(logEvent{msg: "third"})
	expected := []logEvent{
		logEvent{msg: "first"},
		logEvent{msg: "second"},
		logEvent{msg: "third"},
	}
	assert.Equal(t, expected, queue.events)
}

// Assert that events are put before head of the slice.
func Test_queue_put(t *testing.T) {
	queue := new(eventQueue)
	queue.add(logEvent{msg: "second"})
	queue.add(logEvent{msg: "third"})
	queue.put([]logEvent{logEvent{msg: "first"}})
	expected := []logEvent{
		logEvent{msg: "first"},
		logEvent{msg: "second"},
		logEvent{msg: "third"},
	}
	assert.Equal(t, expected, queue.events)
}

// Assert that batch is sorted.
func Test_queue_sorted_batch(t *testing.T) {
	queue := new(eventQueue)
	queue.add(logEvent{timestamp: 2})
	queue.add(logEvent{timestamp: 1})
	assert.Equal(t, logEvent{timestamp: 1}, queue.getBatch()[0])
}
