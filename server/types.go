package main

import "time"

type Reading struct {
	Id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    int64     `json:"uptime"`
	MeatTemp  float64   `json:"meatTemp"`
	SmokeTemp float64   `json:"smokeTemp"`
}

type Cook struct {
	Id          int       `json:"id"`
	Created     time.Time `json:"created"`
	Description string    `json:"description"`
	Readings    []Reading `json:"readings"`
}

type CookReadingUpdate struct {
	Add    []int `json:"add"`
	Remove []int `json:"remove"`
}

type Message struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type HTTPReqInfo struct {
	method    string
	uri       string
	referer   string
	ipaddr    string
	code      int
	size      int64
	duration  time.Duration
	userAgent string
}
