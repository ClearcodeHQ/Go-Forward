package main

import (
	"sort"
	"testing"
	"time"
)

func TestMessageSorting(t *testing.T) {
	unsorted := messageBatch{
		logEvent{timestamp: 2},
		logEvent{timestamp: 1},
	}

	sort.Sort(unsorted)

	for i, elem := range []int64{1,2} {
		if unsorted[i].timestamp != elem {
			t.Errorf("Timestamps should be equal. Got: %v Expected: %v", unsorted[i].timestamp, elem)
		}
	}
}

func TestBatchTimeSpan(t *testing.T) {
	events := messageBatch{
		logEvent{timestamp: 1},
		logEvent{timestamp: 2},
		logEvent{timestamp: 3},
	}

	result := events.timeSpan()
	expected := time.Duration(time.Millisecond * 2)

	if result != expected {
		t.Errorf("Time duration shoud be equal. Result: %v Expected: %v", result, expected)
	}
}

func TestBatchSize(t *testing.T) {
	events := messageBatch{
		logEvent{msg: "123456"},
		logEvent{msg: "12345"},
		logEvent{msg: "123"},
	}

	result := events.size()
	expected := (6 + 5 + 3) + (eventSizeOverhead * 3)

	if result != expected {
		t.Errorf("Batch size shoud be %v. Got: %v", expected, result)
	}
}
