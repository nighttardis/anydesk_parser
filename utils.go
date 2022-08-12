package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AnydeskSession struct {
	SessionStart bool
	SessionEnd   bool
	Username     string
	Userid       int
	LogEntries   []LogEntry
}

type LogEntry struct {
	LogLevel     string `json:"loglevel"`
	Datetime     string `json:"datetime"`
	AppName      string `json:"appname"`
	ThreadName   string `json:"threadname"`
	Pid          int    `json:"pid"`
	Threadid     int    `json:"threadid"`
	Fiberid      int    `json:"fiberid"`
	FunctionName string `json:"functionname"`
	Message      string `json:"message"`
}

func (le *LogEntry) parseFunction(ads *AnydeskSession) {
	if le.FunctionName == "anynet.any_socket" && strings.HasPrefix(strings.TrimSpace(le.Message), "Accept request from") {
		tmp := regexp.MustCompile(`Accept\srequest\sfrom\s(?P<userid>\d+)\s\(via\s(?P<connectionmethod>[^\)]+)\)\.`)
		match := tmp.FindStringSubmatch(le.Message)
		fmt.Println("Found New Connection")
		ads.SessionStart = true
		ads.Userid, _ = strconv.Atoi(match[1])
	}
	if le.FunctionName == "app.session" && strings.HasPrefix(strings.TrimSpace(le.Message), "Session closed by") {
		fmt.Println("Session Closed")
		ads.SessionEnd = true
	}

}
