package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	expected := eventsList{
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

// Assert that batch size does not exceed its allowed maximum
func Test_sizeIndex_multi(t *testing.T) {
	events := eventsList{
		logEvent{msg: RandomString(maxEventSize)},
		logEvent{msg: RandomString(maxEventSize)},
		logEvent{msg: RandomString(maxEventSize)},
		logEvent{msg: RandomString(maxEventSize)},
	}
	assert.Equal(t, 3, sizeIndex(events))
}

// Assert that batch time span does not exceed its allowed maximum
func Test_timeIndex_multi(t *testing.T) {
	events := eventsList{
		logEvent{timestamp: maxBatchTimeSpan},
		logEvent{timestamp: maxBatchTimeSpan},
		logEvent{timestamp: maxBatchTimeSpan * 3},
	}
	assert.Equal(t, 2, timeIndex(events))
}

// Assert that batch time span does not exceed it allowed maximum
func Test_timeIndex_single(t *testing.T) {
	events := eventsList{
		logEvent{timestamp: maxBatchTimeSpan},
	}
	assert.Equal(t, 1, timeIndex(events))
}

// Assert that lowest index is returned
func Test_numEvents_min(t *testing.T) {
	events := make(eventsList, 0)
	funcA := func(e eventsList) int { return maxBatchEvents - 1 }
	funcB := func(e eventsList) int { return maxBatchEvents - 2 }
	assert.Equal(t, maxBatchEvents-2, numEvents(events, funcA, funcB))
}

// Assert that maximum index is returned
func Test_numEvents_max(t *testing.T) {
	events := make(eventsList, 0)
	funcA := func(e eventsList) int { return maxBatchEvents }
	funcB := func(e eventsList) int { return maxBatchEvents }
	assert.Equal(t, maxBatchEvents, numEvents(events, funcA, funcB))
}

func Test_eventList_size(t *testing.T) {
	events := eventsList{
		logEvent{msg: "123456"},
		logEvent{msg: "12345"},
		logEvent{msg: "123"},
	}
	assert.Equal(t, 92, events.size())
}

// Assert that events are sorted by timestamp in ascending order
func Test_eventList_Sort(t *testing.T) {
	sorted := eventsList{
		logEvent{timestamp: 1},
		logEvent{timestamp: 2},
	}
	to_sort := eventsList{
		logEvent{timestamp: 2},
		logEvent{timestamp: 1},
	}
	sort.Sort(to_sort)
	assert.Equal(t, sorted, to_sort)
}
