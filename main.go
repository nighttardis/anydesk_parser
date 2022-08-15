package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {

	inputFile := flag.String("inputFile", "default", "Required - Path to ad.trace file")
	outputDir := flag.String("outputDir", ".", "Directory to write output to.")

	flag.Parse()

	if *inputFile == "default" {
		fmt.Println("[+] ERROR inputFile was not provided")
		os.Exit(1)
	}

	logreg := `^\s+(?P<loglevel>[^\s]+)\s(?P<datetime>[\d-]+\s[\d:\.]+)\s+(?P<appname>\w+)\s+(?P<threadname>[a-z][^\s]+)?\s+(?P<pid>\d+)\s+(?P<threadid>\d+)\s+(?P<fiberid>\d+)?\s+(?P<functionname>[^\s]+)\s+-(?P<message>.+)`
	logre := regexp.MustCompile(logreg)

	file, err := os.Open(*inputFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	var ads AnydeskSession

	for fileScanner.Scan() {
		match := logre.FindStringSubmatch(fileScanner.Text())
		if len(match) == 0 {
			continue
		}
		le := make(map[string]interface{})
		for i, name := range logre.SubexpNames() {
			if i != 0 {
				switch name {
				case "pid", "fiberid", "threadid":
					le[name], _ = strconv.Atoi(match[i])
				default:
					le[name] = strings.TrimSpace(match[i])
				}

			}
		}
		jsonbody, _ := json.Marshal(le)
		leS := LogEntry{}
		json.Unmarshal(jsonbody, &leS)

		leS.parseFunction(&ads)

		if ads.SessionStart && leS.FunctionName != "main" {
			ads.LogEntries = append(ads.LogEntries, leS)
		}

		if ads.SessionEnd {
			if len(ads.FileTransfer) > 0 {
				fmt.Printf("[+] INFO File transfer occured in Session: %d\n", ads.SessionId)
			}
			if len(ads.TextCopied) > 0 {
				fmt.Printf("[+] INFO Text coping occured in Session: %d\n", ads.SessionId)
			}
			// ads.printSession()
			ads.saveSession(*outputDir)
			if !ads.SessionStart {
				fmt.Println("[+] WARNING Found session need before finding a session begin.")
				fmt.Println("[+] WARNING This could be because of rolled logs or parsing errors.")
			}
			ads = AnydeskSession{}
		}

	}
	if ads.SessionStart && !ads.SessionEnd {
		fmt.Println("[+] WARNING Found Session start but didn't find the end of the session before finish reading hte file.")
		fmt.Println("[+] WARNING This could be because of on going connection when pulling the logs or parsing errors.")
		if len(ads.FileTransfer) > 0 {
			fmt.Printf("[+] INFO File transfer occured in Session: %d\n", ads.SessionId)
		}
		if len(ads.TextCopied) > 0 {
			fmt.Printf("[+] INFO Text coping occured in Session: %d\n", ads.SessionId)
		}
		ads.printSession()
	}

}
