package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type streamBond struct {
	url    string
	group  string
	stream string
}

type destMsg struct {
	dst   *destination
	event logEvent
}

type destMap map[receiver]*destination

func main() {
	bonds := []streamBond{
		{url: "udp://localhost:5514", group: "lkostka", stream: "test"},
	}
	cwlogs := cwlogsSession()
	mapping := createMap(bonds, cwlogs)
	fmt.Println("Seting tokens.")
	setTokens(mapping)
	fmt.Println("Seting flow.")
	setupFlow(mapping)
	for {
	}
}

func setTokens(dests destMap) {
	for _, dst := range dests {
		dst.setToken()
	}
}

func closeAll(dests destMap) {
	for recv := range dests {
		recv.Close()
	}
}

func createMap(bonds []streamBond, svc *cloudwatchlogs.CloudWatchLogs) (mapping destMap) {
	mapping = make(destMap)
	for _, bond := range bonds {
		rec, err := newReceiver(bond.url)
		dst := destination{
			group:  bond.group,
			stream: bond.stream,
			svc:    svc,
		}
		mapping[rec] = &dst
		if err != nil {
			closeAll(mapping)
			panic(err)
		}
	}
	return
}

func setupFlow(mapping destMap) {
	for recv, dst := range mapping {
		in := recv.Receive()
		out := make(chan logEvent)
		go convertEvents(in, out, parserFunctions["RFC3339"], formatterFunctions["default"])
		go recToDst(out, dst)
	}
}

// Parse,filter incimming messages and send them to destination.
func convertEvents(in <-chan string, out chan<- logEvent, parsefn syslogParser, fmtfn syslogFormatter) {
	defer close(out)
	for msg := range in {
		if parsed, err := parsefn(msg); err == nil {
			// No point in sending empty messages.
			if parsed.message == "" {
				continue
			}
			// Timestamp must be in milliseconds
			event := logEvent{msg: fmtfn(parsed), timestamp: parsed.timestamp.Unix() * 1000}
			if err := event.validate(); err == nil {
				out <- event
			}
		}
	}
}

// Buffer received events and send them to cloudwatch.
func recToDst(in <-chan logEvent, dst *destination) {
	var pending, received messageBatch
	var uploadDone chan error
	for {
		select {
		case event := <-in:
			received = append(received, event)
		case result := <-uploadDone:
			fmt.Println(result)
			uploadDone = nil
		case <-time.Tick(putLogEventsDelay):
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			length := len(received)
			if length > 0 && uploadDone == nil {
				pending, received = received[:numEvents(length)], received[numEvents(length):]
				uploadDone = make(chan error)
				go func() {
					uploadDone <- dst.upload(pending)
				}()
				fmt.Println(length, "events sent")
			}
		}
	}
}
