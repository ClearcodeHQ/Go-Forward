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
	setTokens(mapping)
	received := convertEvents(mapping)
	destQueue := make(map[*destination]messageBatch)
	var pending messageBatch
	var uploadDone chan error
	for {
		select {
		case event := <-received:
			destQueue[event.dst] = append(destQueue[event.dst], event.event)
		case result := <-uploadDone:
			uploadDone = nil
			fmt.Println("Upload result", result)
		case <-time.Tick(putLogEventsDelay):
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			for dst, messages := range destQueue {
				length := len(messages)
				if length > 0 && uploadDone == nil {
					pending, destQueue[dst] = messages[:numEvents(length)], messages[numEvents(length):]
					uploadDone = make(chan error)
					fmt.Printf("Sending %d messages.\n", len(pending))
					go func() {
						uploadDone <- dst.upload(pending)
					}()
					break
				}
			}
		}
	}
}

func setTokens(dests destMap) {
	for _, dst := range dests {
		fmt.Println("set tokens err:", dst.setToken(), "token:", dst.token)
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

func convertEvents(mapping destMap) <-chan destMsg {
	out := make(chan destMsg)

	for rec, dst := range mapping {
		rout := rec.Receive()
		go recToDst(rout, dst, parserFunctions["RFC3164"], formatterFunctions["default"], out)
	}
	return out
}

func recToDst(in <-chan string, dst *destination, parsefn syslogParser, fmtfn syslogFormatter, out chan<- destMsg) {
	for msg := range in {
		if parsed, err := parsefn(msg); err == nil {
			out <- destMsg{
				dst: dst,
				// Timestamp must be in milliseconds
				event: logEvent{msg: fmtfn(parsed), timestamp: parsed.timestamp.Unix() * 1000},
			}
		}
	}
}
