package main

import (
	"github.com/go-ini/ini"
	"math/rand"
	"time"
)

// Create a N long random string
func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func fixture_valid_config() *ini.File {
	i, _ := ini.Load([]byte(""))
	i.DeleteSection(ini.DEFAULT_SECTION)
	sec, _ := i.NewSection("valid")
	sec.NewKey("group", "group")
	sec.NewKey("stream", "stream")
	sec.NewKey("source", "udp://localhost:5514")
	return i
}
