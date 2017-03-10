package main

import (
	"sort"
)

type eventsList []logEvent

// Calculate size including each event overhead.
func (m eventsList) size() (size int) {
	for _, elem := range m {
		size += elem.size()
	}
	return
}

func (m eventsList) Len() int {
	return len(m)
}

func (m eventsList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m eventsList) Less(i, j int) bool {
	return m[i].timestamp < m[j].timestamp
}

type eventQueue struct {
	events   eventsList
	max_size queue_size
}

func (q *eventQueue) add(event ...logEvent) {
	left := int(q.max_size) - len(q.events)
	many := event[:min(left, len(event))]
	q.events = append(q.events, many...)
}

func (q *eventQueue) getBatch() (batch eventsList) {
	sort.Sort(q.events)
	index := numEvents(q.events, sizeIndex, timeIndex)
	batch, q.events = q.events[:index], q.events[index:]
	return
}

func (q *eventQueue) empty() bool {
	return len(q.events) == 0
}

func (q *eventQueue) num() int {
	return len(q.events)
}

// Return lowest index based on all check functions
// This function assumes that events are sorted by timestamp in ascending order
func numEvents(events eventsList, checkFn ...indexNumFn) int {
	index := maxBatchEvents
	for _, fn := range checkFn {
		result := fn(events)
		if result < index {
			index = result
		}
	}
	return index
}

type indexNumFn func(events eventsList) int

// This function assumes that events are sorted by timestamp in ascending order
func sizeIndex(events eventsList) int {
	size, index := 0, 0
	for i, event := range events {
		size += event.size()
		if size > maxBatchSize {
			break
		}
		index = i + 1
	}
	return index
}

// This function assumes that events are sorted by timestamp in ascending order
func timeIndex(events eventsList) (index int) {
	first := events[0]
	for i, event := range events {
		if (event.timestamp - first.timestamp) > maxBatchTimeSpan {
			break
		}
		index = i + 1
	}
	return index
}

func min(ints ...int) int {
	m := ints[0]
	for _, i := range ints {
		if i < m {
			m = i
		}
	}
	return m
}
