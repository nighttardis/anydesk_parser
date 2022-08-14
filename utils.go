package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AnydeskSession struct {
	SessionId          int        `json:"sessionid"`
	SessionStart       bool       `json:"sessionstart"`
	SessionStartTime   string     `json:"sessionstarttime"`
	SessionEnd         bool       `json:"sessionend"`
	SessionEndTime     string     `json:"sessionendtime"`
	SessionTime        string     `json:"sessiontime" default:"0"`
	Username           string     `json:"username"`
	Userid             int        `json:"userid"`
	Srcip              string     `json:"scrip"`
	Os                 string     `json:"os"`
	ConnectionFlags    string     `json:"connectionflags"`
	Version            string     `json:"version"`
	Authtype           string     `json:"authtype"`
	Authprofile        string     `json:"authprofile"`
	Authtokenattempted bool       `json:"authtokenattempted" default:"false"`
	Setuptoken         bool       `json:"setuptoken" default:"false"`
	FileTransfer       []LogEntry `json:"filetransfer"`
	TextCopied         []LogEntry `json:"textcopied"`
	LogEntries         []LogEntry `json:"logentries"`
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
	switch le.FunctionName {
	case "anynet.any_socket":
		if strings.HasPrefix(le.Message, "Accept request from") {
			tmp := regexp.MustCompile(`Accept\srequest\sfrom\s(?P<userid>\d+)\s\(via\s(?P<connectionmethod>[^\)]+)\)\.`)
			match := tmp.FindStringSubmatch(le.Message)
			fmt.Println("Found New Connection")
			ads.SessionStart = true
			ads.Userid, _ = strconv.Atoi(match[1]) //userid
			ads.SessionStartTime = le.Datetime
		}
		if strings.HasPrefix(le.Message, "Logged in from") {
			tmp := regexp.MustCompile(`Logged\sin\sfrom\s(?P<srcip>[^\:]+)\:(?P<srcport>\d+)\son\srelay\s(?P<relayname>.+)\.`)
			match := tmp.FindStringSubmatch(le.Message)
			ads.Srcip = match[1] // srcip
		}
	case "app.session":
		if strings.HasPrefix(le.Message, "Session closed by") {
			fmt.Println("Session Closed")
			ads.SessionEnd = true
			ads.SessionEndTime = le.Datetime
			if ads.SessionStart {
				tmpEnd, _ := time.Parse("2006-01-02 15:04:05.000", ads.SessionEndTime)
				tmpStart, _ := time.Parse("2006-01-02 15:04:05.000", ads.SessionStartTime)
				ads.SessionTime = tmpEnd.Sub(tmpStart).String()
			}
		}
		if strings.HasPrefix(le.Message, "Connecting to current session") {
			tmp := regexp.MustCompile(`Connecting\sto\scurrent\ssession\s(?P<sessionid>\d+)\.`)
			match := tmp.FindStringSubmatch(le.Message)
			ads.SessionId, _ = strconv.Atoi(match[1]) // sessionid
		}
		if le.Message == "Authenticated by local user." {
			ads.Authtype = "user-approve"
		}
		if le.Message == "Authenticated with correct passphrase." {
			ads.Authtype = "passphrase"
		}
		if le.Message == "Issuing a permanent token." {
			ads.Setuptoken = true
		}
		if strings.HasPrefix(le.Message, "Profile was used") {
			tmp := regexp.MustCompile(`Profile\swas\sused\:\s(?P<profilename>.+)`)
			match := tmp.FindStringSubmatch(le.Message)
			ads.Authprofile = match[1] // profilename
		}
		if le.Message == "Authenticated with permanent token." {
			ads.Authtype = "token"
		}
		if le.Message == "The remote peer has sent a token." {
			ads.Authtokenattempted = true
		}
	case "app.ctrl_clip_comp":
		if strings.HasPrefix(le.Message, "Got a file offer") {
			ads.FileTransfer = append(ads.FileTransfer, *le)
		}
		if strings.HasPrefix(le.Message, "Got a text offer") {
			ads.TextCopied = append(ads.TextCopied, *le)
		}
	case "clipbrd.capture":
		if le.Message != "Registered for clipboard notifications." {
			ads.FileTransfer = append(ads.FileTransfer, *le)
		}
	case "app.prepare_task":
		if strings.HasPrefix(le.Message, "Preparing files") {
			ads.FileTransfer = append(ads.FileTransfer, *le)
		}
	case "app.backend_session":
		if strings.HasPrefix(le.Message, "Incoming session request") {
			tmp := regexp.MustCompile(`Incoming\ssession\srequest\:\s(?P<username>.+)\s\((?P<userid>\d+)\)`)
			match := tmp.FindStringSubmatch(le.Message)
			ads.Username = match[1] // username
		}
		if strings.HasPrefix(le.Message, "Remote OS") {
			tmp := regexp.MustCompile(`Remote\sOS\:\s(?P<os>[^\,]+)\,\sConnection\sflags\:\s(?P<connectionflags>.+)`)
			match := tmp.FindStringSubmatch(le.Message)
			ads.Os = match[1]              // os
			ads.ConnectionFlags = match[2] //connectionflags
		}
		if strings.HasPrefix(le.Message, "Remote version") {
			tmp := regexp.MustCompile(`Remote\sversion\:\s(?P<version>.+)`)
			match := tmp.FindStringSubmatch(le.Message)
			ads.Version = match[1] // version
		}
	case "winapp.gui.permissions_panel":
		if ads.Authprofile == "" && strings.HasPrefix(le.Message, "Selecting Profile") {
			tmp := regexp.MustCompile(`Selecting\sProfile\:\s(?P<profilename>.+)\,\shasPw\:\s(?P<haspw>.+)`)
			match := tmp.FindStringSubmatch(le.Message)
			if match[1] != "_previous_session" {
				ads.Authprofile = match[1] // profilename
			}
		}
	}
}

func (ads *AnydeskSession) printSession() {
	body, _ := json.Marshal(ads)
	fmt.Println(string(body))
}

func (ads *AnydeskSession) saveSession() {
	f, _ := json.MarshalIndent(ads, "", "    ")
	_ = os.WriteFile("data/test.json", f, 0644)
}
