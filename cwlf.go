package main

import (
	"net"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)


const LISTEN_ADDRESS = "localhost:5514"
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
