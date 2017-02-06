package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var version string

const defaultConfigFile = "/etc/logs_agent.cfg"

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

func init() {
	debug()
}

var wg = &sync.WaitGroup{}

func main() {
	var cfgfile string
	var print_version bool
	flag.StringVar(&cfgfile, "c", defaultConfigFile, "Config file location.")
	flag.BoolVar(&print_version, "v", false, "Print version and exit.")
	flag.Parse()
	if print_version {
		fmt.Println(version)
		os.Exit(0)
	}
	log.SetFormatter(&programFormat{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.ErrorLevel)
	config := NewIniConfig(cfgfile)
	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}
	settings := config.GetMain()
	flows := config.GetFlows()
	cwlogs := cwlogsSession()
	log.SetOutput(ioutil.Discard)
	hook := pickHook(strToOutput[settings.LogOutput])
	log.AddHook(hook)
	log.SetLevel(strToLevel[settings.LogLevel])
	receivers := setupFlows(flows, cwlogs)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signals:
		log.Infof("got SIGINT/SIGTERM")
		break
	}
	closeAll(receivers)
	log.Debugf("waiting for upload to finish")
	wg.Wait()
}

func closeAll(receivers []receiver) {
	log.Info("closing connections")
	for _, receiver := range receivers {
		receiver.Close()
	}
}

func setupFlows(flows []*FlowCfg, service *cloudwatchlogs.CloudWatchLogs) (receivers []receiver) {
	log.Debug("seting flow")
	for _, flow := range flows {
		dst := &destination{
			svc:    service,
			stream: flow.Stream,
			group:  flow.Group,
		}
		receiver := newReceiver(flow.Source)
		receivers = append(receivers, receiver)
		if err := receiver.Listen(); err != nil {
			closeAll(receivers)
			log.Fatal(err)
		}
		in := receiver.Receive()
		out := make(chan logEvent)
		format, _ := template.New("").Parse(flow.CloudwatchFormat)
		go convertEvents(in, out, parserFunctions[flow.SyslogFormat], format)
		go recToDst(out, dst, putLogEventsDelay)
	}
	return
}

// Parse, filter incoming messages and send them to destination.
func convertEvents(in <-chan string, out chan<- logEvent, parsefn syslogParser, tpl *template.Template) {
	defer close(out)
	buf := bytes.NewBuffer([]byte{})
	for msg := range in {
		parsed, err := parsefn(msg)
		if err != nil {
			continue
		}
		err = parsed.render(tpl, buf)
		if err != nil {
			continue
		}
		// Timestamp must be in milliseconds
		event := logEvent{
			msg:       buf.String(),
			timestamp: parsed.timestamp.Unix() * 1000,
		}
		err = event.validate()
		if err != nil {
			continue
		}
		out <- event
	}
}

// Buffer received events and send them to cloudwatch.
func recToDst(in <-chan logEvent, dst *destination, delay time.Duration) {
	wg.Add(1)
	defer wg.Done()
	log.Debugf("%s setting token", dst)
	dst.setToken()
	ticker := time.NewTicker(delay)
	log.Debugf("%s timer set to %s", dst, delay)
	defer ticker.Stop()
	queue := &eventQueue{max_size: 10000}
	var uploadDone chan batchFunc
	var batch eventsList
	for {
		select {
		case event, opened := <-in:
			if !opened {
				in = nil
				break
			}
			queue.add(event)
		case fn := <-uploadDone:
			fn(batch, queue)
			uploadDone = nil
		case <-ticker.C:
			log.Debugf("%s tick", dst)
			if !queue.empty() && uploadDone == nil {
				uploadDone, batch = upload(dst, queue)
			}
		}
		if in == nil && queue.empty() {
			break
		}
	}
}

/*
	Sequence token must change in order to send next messages,
	otherwise DataAlreadyAcceptedException is returned.
	Only one upload can proceed / tick / stream.
*/
func upload(dst *destination, queue *eventQueue) (out chan batchFunc, batch eventsList) {
	batch = queue.getBatch()
	out = make(chan batchFunc)
	log.Debugf("%s sending %d messages", dst, len(batch))
	go func() {
		result := dst.upload(batch)
		out <- handleResult(dst, result)
	}()
	return out, batch
}

func handleResult(dst *destination, result error) batchFunc {
	switch err := result.(type) {
	case awserr.Error:
		switch err.Code() {
		case "InvalidSequenceTokenException":
			log.Debugf("%s invalid sequence token", dst)
			dst.setToken()
			return addBack
		case "ResourceNotFoundException":
			log.Debugf("%s missing group/stream", dst)
			dst.create()
			dst.token = nil
			return addBack
		default:
			log.Errorf("upload to %s failed %s %s", dst, err.Code(), err.Message())
		}
	case nil:
	default:
		log.Errorf("upload to %s failed %s ", dst, result)
	}
	return discard
}

type batchFunc func(batch eventsList, queue *eventQueue)

func addBack(batch eventsList, queue *eventQueue) {
	queue.add(batch...)
}

func discard(batch eventsList, queue *eventQueue) {}
