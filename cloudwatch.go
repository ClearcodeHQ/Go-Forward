package main

import (
	"time"
)

/*
AWS CloudWatch specific constants.
Also see http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/cloudwatch_limits_cwl.html
*/
const (
	// Maximum number of log events in a batch.
	maxBatchEvents = 10000
	// Maximum batch size in bytes.
	maxBatchSize = 1048576
	// Maximum event size in bytes.
	maxEventSize = 262144
	// A batch of log events in a single PutLogEvents request cannot span more than 24 hours.
	maxBatchTimeSpan = 24 * time.Hour
	// How many bytes to append to each log event.
	eventSizeOverhead = 26
	// None of the log events in the batch can be more than 2 hours in the future.
	eventFutureTimeDelta = 2 * time.Hour
	// None of the log events in the batch can be older than 14 days.
	eventPastTimeDelta = 14 * 24 * time.Hour
	// DescribeLogStreams transactions/second.
	describeLogstreamsTPS = 5
	// PutLogEvents 5 requests/second/log stream.
	putLogEventsRPS = 5
)

type logEvent struct {
	msg string
	// Timestamp in milliseconds
	timestamp int64
}

type messageBatch []logEvent

// Calculate batch size including each event overhead.
func (m messageBatch) size() (size int) {
	for _, elem := range m {
		size += len(elem.msg) + eventSizeOverhead
	}
	return
}

// Calculate timespan for events.
// !!!! This functions assumes that batch is already sorted by unix timestamp in ascending order.
func (m messageBatch) timeSpan() time.Duration {
	newest := m[len(m)-1].timestamp
	oldest := m[0].timestamp
	return time.Duration(newest-oldest) * time.Millisecond
}

func (m messageBatch) Len() int {
	return len(m)
}

func (m messageBatch) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m messageBatch) Less(i, j int) bool {
	return m[i].timestamp < m[j].timestamp
}
