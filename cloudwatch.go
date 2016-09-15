package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
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
	describeLogstreamsDelay = 200 * time.Millisecond
	// PutLogEvents 5 requests/second/log stream.
	putLogEventsDelay = 200 * time.Millisecond
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

// Calculate batch size based on ammount of received events.
func numEvents(length int) int {
	if length <= maxBatchEvents {
		return length
	}
	return maxBatchEvents
}

type Destination struct {
	stream string
	group  string
	token  string
	svc    *cloudwatchlogs.CloudWatchLogs
}

// Put log events and update sequence token.
// Possible errors http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutLogEvents.html
func (dst *Destination) upload(events messageBatch) error {
	logevents := make([]*cloudwatchlogs.InputLogEvent, 0, len(events))
	for _, elem := range events {
		logevents = append(logevents, &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(elem.msg),
			Timestamp: aws.Int64(elem.timestamp),
		})
	}
	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     logevents,
		LogGroupName:  aws.String(dst.group),
		LogStreamName: aws.String(dst.stream),
		SequenceToken: aws.String(dst.token),
	}
	resp, err := dst.svc.PutLogEvents(params)
	if err == nil {
		// Assign value (not pointer) so that response may be garbage collected.
		dst.token = *resp.NextSequenceToken
	}
	return err
}

func (dst *Destination) setToken() error {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(dst.group),
		LogStreamNamePrefix: aws.String(dst.stream),
	}

	return dst.svc.DescribeLogStreamsPages(params,
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			return !findToken(dst, page)
		})
}

// Create log group and stream. If an error is returned, PutLogEvents can not succeed.
func (dst *Destination) create() (err error) {
	// LimitExceededException
	err = dst.createGroup()
	err = dst.createStream()
	return
}

func (dst *Destination) createGroup() error {
	params := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(dst.group),
	}
	_, err := dst.svc.CreateLogGroup(params)
	// ResourceAlreadyExistsException
	// InvalidParameterException when name is invalid
	return err
}

func (dst *Destination) createStream() error {
	params := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(dst.group),
		LogStreamName: aws.String(dst.stream),
	}
	_, err := dst.svc.CreateLogStream(params)
	// ResourceAlreadyExistsException
	// ResourceNotFoundException when there is no group
	// InvalidParameterException when name is invalid
	return err
}

func findToken(dst *Destination, page *cloudwatchlogs.DescribeLogStreamsOutput) bool {
	for _, row := range page.LogStreams {
		if dst.stream == *row.LogStreamName {
			// Assign value (not pointer) so that page may be garbage collected.
			dst.token = *row.UploadSequenceToken
			return true
		}
	}
	return false
}
