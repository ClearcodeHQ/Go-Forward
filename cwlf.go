package main

import (
	"net"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)


const LISTEN_ADDRESS = "localhost:5514"
// Maximum number of log events in a batch.
const MAX_BATCH_EVENTS = 10000
// Maximum batch size in bytes.
const MAX_BATCH_SIZE = 1048576
// A batch of log events in a single PutLogEvents request cannot span more than 24 hours.
const MAX_BATCH_SPAN_TIME = 24 * time.Hour
// How many bytes to append to each log event.
const EVENT_SIZE_OVERHEAD = 26
// None of the log events in the batch can be more than 2 hours in the future.
const EVENT_FUTURE_TIMEDELTA = 2 * time.Hour
// None of the log events in the batch can be older than 14 days.
const EVENT_PAST_TIMEDELTA = 14 * 24 * time.Hour
// DescribeLogStreams transactions/second.
const DescribeLogStreams_TPS = 5

var params = &cloudwatchlogs.DescribeLogGroupsInput{Limit: aws.Int64(50)}


func main() {
	listen_addr, err := net.ResolveUDPAddr("udp", LISTEN_ADDRESS)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", listen_addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	var buf [2048]byte
	for {
		n, err:= conn.Read(buf[0:])
		if err != nil {
			panic(err)
		}
		parsed, err := decodeMessage(string(buf[0:n]))
		if err == nil {
			fmt.Println(parsed)
		} else {
			fmt.Println(err)
		}
	}
}


func getLogGroups(region string) (groups []string) {
	svc := cloudwatchlogs.New(session.New(), aws.NewConfig().WithRegion(region))

	err := svc.DescribeLogGroupsPages(params,
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			groups = append(groups, getGroupNames(page)...)
			return !lastPage
		})

	if err != nil {
		panic(err)
	}
	return
}


func getGroupNames(page *cloudwatchlogs.DescribeLogGroupsOutput) (names []string) {
	for _, val := range page.LogGroups {
		names = append(names, *val.LogGroupName)
	}
	return
}
