package main

import (
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	log "github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
)

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
