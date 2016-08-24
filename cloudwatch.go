package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
	describeLogstreamsTPS = 5
	// PutLogEvents 5 requests/second/log stream.
	putLogEventsRPS = 5
)

type logEvent struct {
	msg       *syslogMessage
	formatted string
}

type messageBatch []logEvent

// Calculate batch size including each event overhead.
func (m messageBatch) size() (size int) {
	for _, elem := range m {
		size += len(elem.formatted) + eventSizeOverhead
	}
	return
}

// Calculate timespan for events.
// !!!! This functions assumes that batch is already sorted by unix timestamp in ascending order.
func (m messageBatch) timeSpan() time.Duration {
	newest := m[len(m)-1].msg.timestamp
	oldest := m[0].msg.timestamp
	return newest.Sub(oldest)
}

func (m messageBatch) Len() int           { return len(m) }
func (m messageBatch) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m messageBatch) Less(i, j int) bool { return m[i].msg.timestamp.Unix() < m[j].msg.timestamp.Unix() }

var params = &cloudwatchlogs.DescribeLogGroupsInput{Limit: aws.Int64(50)}

// Return all log groups in a given region.
func getLogGroups(region string) (groups []string) {
	svc := cloudwatchlogs.New(session.New(), aws.NewConfig().WithRegion(region))

	err := svc.DescribeLogGroupsPages(params,
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			groups = append(groups, getGroupNames(page)...)
			return !lastPage
		})

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		// awsErr  := err.(awserr.Error)
		// Generic AWS Error with Code, Message, and original error (if any)
		// fmt.Println(awsErr.Code(), awsErr.Message())
		panic(err)
	}
	return
}

// Return all log group names from DescribeLogGroupsOutput.
func getGroupNames(page *cloudwatchlogs.DescribeLogGroupsOutput) (names []string) {
	for _, val := range page.LogGroups {
		names = append(names, *val.LogGroupName)
	}
	return
}
