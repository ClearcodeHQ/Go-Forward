package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	maxBatchTimeSpan = 86400000
	// How many bytes to append to each log event.
	eventSizeOverhead = 26
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

func (e *logEvent) size() int {
	return len(e.msg) + eventSizeOverhead
}

func (e *logEvent) validate() error {
	if e.size() > maxEventSize {
		return errMessageTooBig
	}
	return nil
}

type destination struct {
	stream string
	group  string
	token  *string
	svc    *cloudwatchlogs.CloudWatchLogs
}

func newDestination(stream, group string) *destination {
	dst := &destination{
		svc:    cwlogs,
		stream: stream,
		group:  group,
	}
	log.Debugf("%s setting token", dst)
	dst.setToken()
	return dst
}

// Put log events and update sequence token.
// Possible errors http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutLogEvents.html
func (dst *destination) upload(events eventsList) error {
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
		SequenceToken: dst.token,
	}
	// When rejectedLogEventsInfo is not empty, app can not
	// do anything reasonable with rejected logs. Ignore it.
	// Meybe expose some statistics for rejected counters.
	resp, err := dst.svc.PutLogEvents(params)
	if err == nil {
		dst.token = resp.NextSequenceToken
	}
	return err
}

// For newly created log streams, token is an empty string.
func (dst *destination) setToken() error {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(dst.group),
		LogStreamNamePrefix: aws.String(dst.stream),
	}

	return dst.svc.DescribeLogStreamsPages(params,
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			return !findToken(dst, page)
		})
}

// Create log group and stream. If an error is returned, PutLogEvents cannot succeed.
func (dst *destination) create() (err error) {
	err = dst.createGroup()
	if err != nil {
		return
	}
	err = dst.createStream()
	return
}

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
func (dst *destination) createGroup() error {
	params := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(dst.group),
	}
	_, err := dst.svc.CreateLogGroup(params)
	if err, ok := err.(awserr.Error); ok {
		if err.Code() == "ResourceAlreadyExistsException" {
			return nil
		}
	}
	return err
}

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogStream.html
func (dst *destination) createStream() error {
	params := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(dst.group),
		LogStreamName: aws.String(dst.stream),
	}
	_, err := dst.svc.CreateLogStream(params)
	if err, ok := err.(awserr.Error); ok {
		if err.Code() == "ResourceAlreadyExistsException" {
			return nil
		}
	}
	return err
}

func (dst *destination) String() string {
	return fmt.Sprintf("group: %s stream: %s", dst.group, dst.stream)
}

func findToken(dst *destination, page *cloudwatchlogs.DescribeLogStreamsOutput) bool {
	for _, row := range page.LogStreams {
		if dst.stream == *row.LogStreamName {
			dst.token = row.UploadSequenceToken
			return true
		}
	}
	return false
}
