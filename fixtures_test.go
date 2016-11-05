package main

import (
	"math/rand"
	"time"

	"github.com/go-ini/ini"
)

type numPair struct {
	expected int
	passed   int
}

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

func empty_ini_section() *ini.Section {
	i, _ := ini.Load([]byte(""))
	i.DeleteSection(ini.DEFAULT_SECTION)
	sec, _ := i.NewSection("fixture")
	return sec
}
