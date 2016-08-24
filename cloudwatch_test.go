package main

import (
	"sort"
	"testing"
	"time"
)

func TestMessageSorting(t *testing.T) {
	unsorted := messageBatch{
		logEvent{msg: &syslogMessage{timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC)}},
		logEvent{msg: &syslogMessage{timestamp: time.Date(2016, 7, 23, 12, 48, 11, 969683000, time.UTC)}},
	}
	sorted := messageBatch{
		logEvent{msg: &syslogMessage{timestamp: time.Date(2016, 7, 23, 12, 48, 11, 969683000, time.UTC)}},
		logEvent{msg: &syslogMessage{timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC)}},
	}

	sort.Sort(unsorted)

	for i, elem := range sorted {
		unsortedUnix := unsorted[i].msg.timestamp.Unix()
		sortedUnix := elem.msg.timestamp.Unix()
		if unsortedUnix != sortedUnix {
			t.Errorf("Timestamps should be equal. Unsorted: %v Sorted: %v", unsortedUnix, sortedUnix)
		}
	}
}

func TestBatchTimeSpan(t *testing.T) {
	oldest := time.Date(2016, 7, 23, 12, 48, 11, 969683000, time.UTC)
	newest := time.Date(2016, 7, 23, 12, 48, 18, 969683000, time.UTC)
	events := messageBatch{
		logEvent{msg: &syslogMessage{timestamp: oldest}},
		logEvent{msg: &syslogMessage{timestamp: time.Date(2016, 7, 23, 12, 48, 16, 969683000, time.UTC)}},
		logEvent{msg: &syslogMessage{timestamp: newest}},
	}

	result := events.timeSpan()
	expected := newest.Sub(oldest)

	if result != expected {
		t.Errorf("Time duration shoud be equal. Result: %v Expected: %v", result, expected)
	}
}

func TestBatchSize(t *testing.T) {
	events := messageBatch{
		logEvent{msg: &syslogMessage{}, formatted: "123456"},
		logEvent{msg: &syslogMessage{}, formatted: "12345"},
		logEvent{msg: &syslogMessage{}, formatted: "123"},
	}

	result := events.size()
	expected := (6+5+3) + (eventSizeOverhead * 3)

	if result != expected {
		t.Errorf("Batch size shoud be %v. Got: %v", expected, result)
	}
}
