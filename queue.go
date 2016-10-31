package main

import (
	"sort"
)

type eventQueue struct {
	events []logEvent
}

func (q *eventQueue) put(elements []logEvent) {
	q.events = append(elements, q.events...)
}

func (q *eventQueue) add(event logEvent) {
	q.events = append(q.events, event)
}

func (q *eventQueue) getBatch() (batch messageBatch) {
	sort.Sort(messageBatch(q.events))
	length := len(q.events)
	batch, q.events = q.events[:numEvents(length)], q.events[numEvents(length):]
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
