package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const listenAddress = "udp://localhost:5514"

func main() {
	rec, err := newReceiver(listenAddress)
	if err != nil {
		panic(err)
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		panic(err)
	}

	dst := Destination{
		group:  "lkostka",
		stream: "test",
		svc:    cloudwatchlogs.New(sess),
	}
	dst.setToken()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	received := convertEvents(rec.Receive())
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
		case <-quit:
			rec.Close()
			return
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
