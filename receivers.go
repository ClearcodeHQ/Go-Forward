package main

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"sync"
)

type receiver interface {
	// Close connection and channels
	Close()
	// Run a goroutine and pass read messages to channel
	Receive() <-chan string
}

type UDPreceiver struct {
	conn *net.UDPConn
	wg   *sync.WaitGroup
}

func (rec *UDPreceiver) Close() {
	rec.conn.Close()
	rec.wg.Wait()
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
				panic(err)
			}
			out <- string(buf[0:n])
		}
	}()
	return out
}

// Create a new receiver based on passed address. This function can
// return receivers for UDP, TCP, UNIX sockets.
func newReceiver(addr string) (rec receiver, err error) {
	uri, err := url.Parse(addr)
	if err != nil {
		return
	}
	switch uri.Scheme {
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", uri.Host)
		if err != nil {
			return rec, err
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return rec, err
		}
		rec = &UDPreceiver{conn: conn, wg: &sync.WaitGroup{}}
		return rec, nil
	}
	err = errors.New("Unknown url scheme.")
	return
}
