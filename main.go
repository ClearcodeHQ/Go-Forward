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

	msgChan := convertEvents(receiveEvents(conn))
	var messages, pending messageBatch
	pendingChan := make(chan messageBatch)
	go sendEvents(dst, pendingChan)
	for {
		select {
		case <-time.Tick(time.Second / putLogEventsRPS):
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			i := 100
			pending, messages = messages[:i], messages[i:]
			fmt.Println(len(messages), cap(messages))
		case event := <-msgChan:
			messages = append(messages, event)
		case pendingChan <- pending:
		}
	}
}

func sendEvents(dst Destination, mchan chan messageBatch) {
	for batch := range mchan {
		if len(batch) > 0 {
			err := dst.upload(batch)
			fmt.Println(err)
		}
	}
}

func convertEvents(m <-chan string) <-chan logEvent {
	out := make(chan logEvent)
	go func() {
		for msg := range m {
			parsed, err := decodeMessage(msg)
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
