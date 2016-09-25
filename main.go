package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type streamBond struct {
	url    string
	group  string
	stream string
}

type destMsg struct {
	dst   *Destination
	event logEvent
}

type destMap map[receiver]*Destination

func main() {
	bonds := []streamBond{
		{url: "udp://localhost:5514", group: "lkostka", stream: "test"},
	}
	cwlogs := cwlogsSession()
	mapping := createMap(bonds, cwlogs)
	setTokens(mapping)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	received := convertEvents(mapping)
	destQueue := make(map[*Destination]messageBatch)
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
		case <-quit:
			closeAll(mapping)
			return
		}
	}
}

func setTokens(dests destMap) {
	for _, dst := range dests {
		fmt.Println("set tokens err:", dst.setToken(), "token:", dst.token)
	}
}

func closeAll(dests destMap) {
	for recv, _ := range dests {
		recv.Close()
	}
}

func createMap(bonds []streamBond, svc *cloudwatchlogs.CloudWatchLogs) (mapping destMap) {
	mapping = make(destMap)
	for _, bond := range bonds {
		rec, err := newReceiver(bond.url)
		dst := Destination{
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
		go func(m <-chan string, d *Destination) {
			for msg := range m {
				parsed, err := parseRFC3164(msg)
				if err == nil {
					out <- destMsg{
						dst: d,
						event: logEvent{
							msg: fmt.Sprintf("%s %s %s %s %s", parsed.facility, parsed.severity, parsed.hostname, parsed.syslogtag, parsed.message),
							// Timestamp must be in milliseconds
							timestamp: parsed.timestamp.Unix() * 1000,
						},
					}
				}
			}
		}(rout, dst)
	}
	return out
}
