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
