package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AnydeskSession struct {
	SessionId        int        `json:"sessionid"`
	SessionStart     bool       `json:"sessionstart"`
	SessionStartTime string     `json:"sessionstarttime"`
	SessionEnd       bool       `json:"sessionend"`
	SessionEndTime   string     `json:"sessionendtime"`
	Username         string     `json:"username"`
	Userid           int        `json:"userid"`
	Srcip            string     `json:"scrip"`
	Os               string     `json:"os"`
	ConnectionFlags  string     `json:"connectionflags"`
	Version          string     `json:"version"`
	Authtype         string     `json:"authtype"`
	Setuptoken       bool       `json:"setuptoken"`
	FileTransfer     []LogEntry `json:"filetransfer"`
	TextCopied       []LogEntry `json:"textcopied"`
	LogEntries       []LogEntry `json:"logentries"`
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
		ads.Userid, _ = strconv.Atoi(match[1]) //userid
		ads.SessionStartTime = le.Datetime
	}
	if le.FunctionName == "app.session" && strings.HasPrefix(le.Message, "Session closed by") {
		fmt.Println("Session Closed")
		ads.SessionEnd = true
		ads.SessionEndTime = le.Datetime
	}
	if le.FunctionName == "anynet.any_socket" && strings.HasPrefix(le.Message, "Logged in from") {
		tmp := regexp.MustCompile(`Logged\sin\sfrom\s(?P<srcip>[^\:]+)\:(?P<srcport>\d+)\son\srelay\s(?P<relayname>.+)\.`)
		match := tmp.FindStringSubmatch(le.Message)
		ads.Srcip = match[1] // srcip
	}
	if le.FunctionName == "app.session" && strings.HasPrefix(le.Message, "Connecting to current session") {
		tmp := regexp.MustCompile(`Connecting\sto\scurrent\ssession\s(?P<sessionid>\d+)\.`)
		match := tmp.FindStringSubmatch(le.Message)
		ads.SessionId, _ = strconv.Atoi(match[1]) // sessionid
	}
	if le.FunctionName == "app.ctrl_clip_comp" && strings.HasPrefix(le.Message, "Got a file offer") {
		ads.FileTransfer = append(ads.FileTransfer, *le)
	}
	if le.FunctionName == "app.ctl_clip_comp" && strings.HasPrefix(le.Message, "Got a text offer") {
		ads.TextCopied = append(ads.TextCopied, *le)
	}
	if le.FunctionName == "app.backend_session" && strings.HasPrefix(le.Message, "Incoming session request") {
		tmp := regexp.MustCompile(`Incoming\ssession\srequest\:\s(?P<username>.+)\s\((?P<userid>\d+)\)`)
		match := tmp.FindStringSubmatch(le.Message)
		ads.Username = match[1] // username
	}
	if le.FunctionName == "app.backend_session" && strings.HasPrefix(le.Message, "Remote OS") {
		tmp := regexp.MustCompile(`Remote\sOS\:\s(?P<os>[^\,]+)\,\sConnection\sflags\:\s(?P<connectionflags>.+)`)
		match := tmp.FindStringSubmatch(le.Message)
		ads.Os = match[1]              // os
		ads.ConnectionFlags = match[2] //connectionflags
	}
	if le.FunctionName == "app.backend_session" && strings.HasPrefix(le.Message, "Remote version") {
		tmp := regexp.MustCompile(`Remote\sversion\:\s(?P<version>.+)`)
		match := tmp.FindStringSubmatch(le.Message)
		ads.Version = match[1] // version
	}
	if le.FunctionName == "clipbrd.capture" && le.Message != "Registered for clipboard notifications." {
		ads.FileTransfer = append(ads.FileTransfer, *le)
	}
	if le.FunctionName == "app.prepare_task" && strings.HasPrefix(le.Message, "Preparing files") {
		ads.FileTransfer = append(ads.FileTransfer, *le)
	}
}
