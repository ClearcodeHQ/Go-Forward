package main

import (
	"net"
	"net/url"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type receiver interface {
	// Close connection and channels
	Close()
	// Run a goroutine and pass read messages to channel
	Receive() <-chan string
	// Listen for incoming packets
	Listen() error
}

type UDPreceiver struct {
	conn *net.UDPConn
	url  *url.URL
	wg   *sync.WaitGroup
}

func (rec *UDPreceiver) Close() {
	if rec.conn != nil {
		rec.conn.Close()
		rec.wg.Wait()
	}
}

func (rec *UDPreceiver) Listen() error {
	addr, err := net.ResolveUDPAddr("udp", rec.url.Host)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	rec.conn = conn
	return err
}

func (rec *UDPreceiver) Receive() <-chan string {
	out := make(chan string, maxBatchEvents)
	rec.wg.Add(1)
	go func() {
		var buf [maxEventSize]byte
		defer rec.wg.Done()
		defer close(out)
		for {
			n, err := rec.conn.Read(buf[0:])
			// For more info why string comparison see https://github.com/golang/go/issues/4373
			if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
				return
			} else if err != nil {
				log.Fatal(err)
			}
			out <- string(buf[0:n])
		}
	}()
	return out
}

// Create a new receiver based on passed address.
func newReceiver(address string) receiver {
	url, _ := url.Parse(address)
	switch url.Scheme {
	case "udp":
		return &UDPreceiver{url: url, wg: &sync.WaitGroup{}}
	}
	return nil
}
