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
	events eventsList
}

func (q *eventQueue) add(event ...logEvent) {
	q.events = append(q.events, event...)
}

func (q *eventQueue) getBatch() (batch eventsList) {
	sort.Sort(q.events)
	batchSize, maxIndex := 0, 0
	for i, event := range q.events {
		batchSize += event.size()
		if batchSize > maxBatchSize {
			break
		}
		maxIndex = i + 1
	}
	batch, q.events = q.events[:numEvents(maxIndex)], q.events[numEvents(maxIndex):]
	return
}

func (q *eventQueue) empty() bool {
	return len(q.events) == 0
}

func (q *eventQueue) num() int {
	return len(q.events)
}

// Calculate batch size based on ammount of received events.
func numEvents(length int) int {
	if length <= maxBatchEvents {
		return length
	}
	return maxBatchEvents
}
