package main

import (
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const listenAddress = "localhost:5514"

func main() {
	listenAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sess, _ := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	dst := Destination{
		group:  "lkostka",
		stream: "test",
		svc:    cloudwatchlogs.New(sess),
	}
	dst.setToken()

	received := convertEvents(receiveEvents(conn))
	var messages, pending messageBatch
	var uploadDone chan error
	for {
		select {
		case event := <-received:
			messages = append(messages, event)
		case result := <-uploadDone:
			uploadDone = nil
			fmt.Println("Upload result", result)
		case <-time.Tick(putLogEventsDelay):
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			length := len(messages)
			if length > 0 && uploadDone == nil {
				pending, messages = messages[:numEvents(length)], messages[numEvents(length):]
				uploadDone = make(chan error)
				fmt.Printf("Sending %d messages.\n", length)
				go func() {
					uploadDone <- dst.upload(pending)
				}()
			}
		}
	}
}

func convertEvents(m <-chan string) <-chan logEvent {
	out := make(chan logEvent)
	go func() {
		for msg := range m {
			parsed, err := parseRFC3164(msg)
			if err == nil {
				out <- logEvent{
					msg: fmt.Sprintf("%s %s %s %s %s", parsed.facility, parsed.severity, parsed.hostname, parsed.syslogtag, parsed.message),
					// Timestamp must be in milliseconds
					timestamp: parsed.timestamp.Unix() * 1000,
				}
			}
		}
	}()
	return out
}

func receiveEvents(conn *net.UDPConn) <-chan string {
	out := make(chan string, maxBatchEvents)
	go func() {
		var buf [maxEventSize]byte
		for {
			n, err := conn.Read(buf[0:])
			if err != nil {
				panic(err)
			}
			out <- string(buf[0:n])
		}
	}()
	return out
}
