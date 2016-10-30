package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type streamBond struct {
	url    string
	group  string
	stream string
}

type destMap map[receiver]*destination

func do_init() []streamBond {
	var cfgFile string
	flag.StringVar(&cfgFile, "c", "/etc/cwlagent.conf", "Config file location.")
	flag.Parse()
	logger := log.New(os.Stderr, "ERROR: ", 0)
	config, err := getConfig(cfgFile)
	if err != nil {
		logger.Fatal(err)
	}
	for _, section := range config.Sections() {
		if section.Name() != generalSection {
			err := validateSection(section)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}
	return getBonds(config)
}

var logger *log.Logger

func main() {
	bonds := do_init()
	logger = log.New(os.Stdout, "DEBUG: ", 0)
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
		mapping[rec] = &dst
		if err != nil {
			closeAll(mapping)
			panic(err)
		}
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
	var pending messageBatch
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

func handleUploadResult(dst *destination, result error, queue *eventQueue, pending messageBatch) {
	switch err := result.(type) {
	case awserr.Error:
		switch err.Code() {
		case "InvalidSequenceTokenException":
			logger.Printf("%s invalid sequence token", dst)
			dst.setToken()
			queue.put(pending)
		case "ResourceNotFoundException":
			logger.Printf("%s missing group/stream", dst)
			dst.create()
			dst.setToken()
			queue.put(pending)
		default:
			logger.Printf("upload to %s failed %s %s", dst, err.Code(), err.Message())
		}
	case nil:
	default:
		logger.Printf("upload to %s failed %s ", dst, result)
	}
	logger.Printf("%s %d messages in queue", dst, queue.num())
}
