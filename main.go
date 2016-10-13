package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type streamBond struct {
	url    string
	group  string
	stream string
}

type destMap map[receiver]*destination

func main() {
	var logger = log.New(os.Stdout, "DEBUG: ", 0)
	bonds := []streamBond{
		{url: "udp://localhost:5514", group: "lkostka", stream: "test"},
	}
	cwlogs := cwlogsSession()
	mapping := createMap(bonds, cwlogs)
	logger.Print("Seting tokens.")
	setTokens(mapping)
	logs := make(chan string)
	logger.Print("Seting flow.")
	setupFlow(mapping, logs)
	for {
		select {
		case log := <-logs:
			logger.Print(log)
		}
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

func setupFlow(mapping destMap, logs chan<- string) {
	for recv, dst := range mapping {
		in := recv.Receive()
		out := make(chan logEvent)
		go convertEvents(in, out, parserFunctions["RFC3339"], formatterFunctions["default"])
		go recToDst(out, dst, logs)
	}
}

// Parse,filter incimming messages and send them to destination.
func convertEvents(in <-chan string, out chan<- logEvent, parsefn syslogParser, fmtfn syslogFormatter) {
	defer close(out)
	for msg := range in {
		if parsed, err := parsefn(msg); err == nil {
			// Timestamp must be in milliseconds
			event := logEvent{msg: fmtfn(parsed), timestamp: parsed.timestamp.Unix() * 1000}
			if err := event.validate(); err == nil {
				out <- event
			}
		}
	}
}

// Buffer received events and send them to cloudwatch.
func recToDst(in <-chan logEvent, dst *destination, logs chan<- string) {
	var pending, received messageBatch
	var uploadDone chan error
	for {
		select {
		case event := <-in:
			received = append(received, event)
		case result := <-uploadDone:
			logs <- fmt.Sprint(dst, result)
			uploadDone = nil
		case <-time.Tick(putLogEventsDelay):
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			length := len(received)
			logs <- fmt.Sprintf("%d messages in buffer for %s", length, dst)
			if length > 0 && uploadDone == nil {
				pending, received = received[:numEvents(length)], received[numEvents(length):]
				logs <- fmt.Sprintf("Sending %d messages to %s", len(pending), dst)
				uploadDone = make(chan error)
				go func() {
					uploadDone <- dst.upload(pending)
				}()
			}
		}
	}
}
