package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"
	"os/signal"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const defaultConfigFile = "/etc/awslogs.conf"

type writerHook struct {
	out io.Writer
}

func (hook *writerHook) Fire(entry *log.Entry) error {
	line, _ := entry.String()
	io.WriteString(hook.out, line)
	return nil
}

func (hook *writerHook) Levels() []log.Level {
	return log.AllLevels
}

type programFormat struct{}

func (f *programFormat) Format(e *log.Entry) ([]byte, error) {
	buf := []byte(e.Message)
	buf = append(buf, byte('\n'))
	return buf, nil
}

func pickHook(out logoutput) log.Hook {
	switch out {
	case sysLog:
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_DAEMON, "awslogs")
		if err != nil {
			log.Fatal("Unable to connect to local syslog daemon")
		}
		return hook
	case stdOut:
		return &writerHook{out: os.Stdout}
	case stdErr:
		return &writerHook{out: os.Stderr}
	case null:
		return &writerHook{out: ioutil.Discard}
	default:
		return &writerHook{out: os.Stderr}
	}
}

func main() {
	debug()
	log.SetFormatter(&programFormat{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.ErrorLevel)
	var cfgfile string
	flag.StringVar(&cfgfile, "c", defaultConfigFile, "Config file location.")
	flag.Parse()
	config := getConfig(cfgfile)
	settings := getMainConfig(config)
	flows := getFlows(config)
	cwlogs := cwlogsSession()
	listenAll(flows)
	log.SetOutput(ioutil.Discard)
	hook := pickHook(settings.logOutput)
	log.AddHook(hook)
	log.SetLevel(settings.logLevel)
	setupFlows(flows, cwlogs)
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		log.Infof("got SIGINT")
		break
	}
	log.Info("closing connections")
	closeAll(flows)
}

func closeAll(flows []flowCfg) {
	for _, flow := range flows {
		flow.recv.Close()
	}
}

func listenAll(flows []flowCfg) {
	for _, flow := range flows {
		if err := flow.recv.Listen(); err != nil {
			log.Fatal(err)
			closeAll(flows)
		}
	}
}

func setupFlows(flows []flowCfg, service *cloudwatchlogs.CloudWatchLogs) {
	log.Debug("seting flow")
	for _, flow := range flows {
		flow.dst.svc = service
		in := flow.recv.Receive()
		out := make(chan logEvent)
		go convertEvents(in, out, flow.syslogFn, flow.format)
		go recToDst(out, flow.dst)
	}
}

// Parse,filter incimming messages and send them to destination.
func convertEvents(in <-chan string, out chan<- logEvent, parsefn syslogParser, tpl *template.Template) {
	buf := bytes.NewBuffer([]byte{})
	for msg := range in {
		if parsed, err := parsefn(msg); err == nil {
			if err := parsed.render(tpl, buf); err == nil {
				// Timestamp must be in milliseconds
				event := logEvent{msg: buf.String(), timestamp: parsed.timestamp.Unix() * 1000}
				if err := event.validate(); err == nil {
					out <- event
				}
			}
		}
	}
	out = nil
}

// Buffer received events and send them to cloudwatch.
func recToDst(in <-chan logEvent, dst *destination) {
	log.Debugf("seting token for %s", dst)
	dst.setToken()
	ticker := time.NewTicker(putLogEventsDelay)
	defer ticker.Stop()
	queue := new(eventQueue)
	var pending eventsList
	var uploadDone chan error
	for {
		select {
		case event := <-in:
			if queue.num() < maxBatchEvents {
				queue.add(event)
			}
		case result := <-uploadDone:
			handleUploadResult(dst, result, queue, pending)
			uploadDone = nil
		case <-ticker.C:
			/*
				Sequence token must change in order to send next messages,
				otherwise DataAlreadyAcceptedException is returned.
				Only one upload can proceed / tick / stream.
			*/
			if !queue.empty() && uploadDone == nil {
				pending = queue.getBatch()
				log.Debugf("%s sending %d messages", dst, len(pending))
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
			log.Debugf("%s invalid sequence token", dst)
			dst.setToken()
			queue.add(pending...)
		case "ResourceNotFoundException":
			log.Debugf("%s missing group/stream", dst)
			dst.create()
			dst.token = nil
			queue.add(pending...)
		default:
			log.Errorf("upload to %s failed %s %s", dst, err.Code(), err.Message())
		}
	case nil:
	default:
		log.Errorf("upload to %s failed %s ", dst, result)
	}
	log.Debugf("%s %d messages in queue", dst, queue.num())
}
