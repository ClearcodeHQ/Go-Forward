package main

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type streamBond struct {
	url    *url.URL
	group  string
	stream string
}

type destMap map[receiver]*destination

type options struct {
	cfgfile string
	debug   bool
	logOut  io.Writer
}

func getOptions(logger *log.Logger) options {
	opts := options{}
	flag.StringVar(&opts.cfgfile, "c", "/etc/logs_agent.cfg", "Config file location.")
	flag.BoolVar(&opts.debug, "d", false, "Turn on debug mode.")
	flag.Parse()
	if opts.debug {
		opts.logOut = os.Stdout
	} else {
		opts.logOut, _ = os.Open(os.DevNull)
	}
	return opts
}

func do_init() (bonds []streamBond, opts options) {
	logger := log.New(os.Stderr, "ERROR: ", 0)
	opts = getOptions(logger)
	config := getConfig(logger, opts.cfgfile)
	bonds = getBonds(config)
	return
}

var logger *log.Logger

func main() {
	bonds, opts := do_init()
	logger = log.New(opts.logOut, "DEBUG: ", 0)
	cwlogs := cwlogsSession()
	mapping := createMap(bonds, cwlogs)
	createAll(mapping)
	setTokens(mapping)
	setupFlow(mapping)
	select {}
}

func setTokens(dests destMap) {
	logger.Print("Seting tokens.")
	for _, dst := range dests {
		logger.Print(dst.setToken())
	}
}

func closeAll(dests destMap) {
	for recv := range dests {
		recv.Close()
	}
}

func createAll(dests destMap) {
	logger.Print("Creating destinations")
	for _, dst := range dests {
		logger.Print(dst.create())
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
		if err != nil {
			closeAll(mapping)
			panic(err)
		}
		mapping[rec] = &dst
	}
	return
}

func setupFlow(mapping destMap) {
	logger.Print("Seting flow.")
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
	queue := new(eventQueue)
	var pending eventsList
	var uploadDone chan error
	for {
		select {
		case event := <-in:
			queue.add(event)
		case result := <-uploadDone:
			handleUploadResult(dst, result, queue, pending)
			uploadDone = nil
		case <-time.Tick(putLogEventsDelay):
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			if !queue.empty() && uploadDone == nil {
				pending = queue.getBatch()
				logger.Printf("%s sending %d messages", dst, len(pending))
				uploadDone = make(chan error)
				go func() {
					uploadDone <- dst.upload(pending)
				}()
			}
		}
	}
}

func handleUploadResult(dst *destination, result error, queue *eventQueue, pending eventsList) {
	switch err := result.(type) {
	case awserr.Error:
		switch err.Code() {
		case "InvalidSequenceTokenException":
			logger.Printf("%s invalid sequence token", dst)
			dst.setToken()
			queue.add(pending...)
		case "ResourceNotFoundException":
			logger.Printf("%s missing group/stream", dst)
			dst.create()
			dst.token = nil
			queue.add(pending...)
		default:
			logger.Printf("upload to %s failed %s %s", dst, err.Code(), err.Message())
		}
	case nil:
	default:
		logger.Printf("upload to %s failed %s ", dst, result)
	}
	logger.Printf("%s %d messages in queue", dst, queue.num())
}
