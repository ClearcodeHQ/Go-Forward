package main

import (
	"fmt"
	"net"
)

const listenAddress = "localhost:5514"

func main() {
	listenAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
}

func receiveEvents(conn *net.UDPConn, c chan syslogMessage) {
	var buf [2048]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			panic(err)
		}
		parsed, err := decodeMessage(string(buf[0:n]))
		if err == nil {
			c <-parsed
		} else {
			fmt.Println(err)
		}
	}
}
