package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	logreg := `^\s+(?P<loglevel>[^\s]+)\s(?P<datetime>[\d-]+\s[\d:\.]+)\s+(?P<appname>\w+)\s+(?P<threadname>[a-z][^\s]+)?\s+(?P<pid>\d+)\s+(?P<threadid>\d+)\s+(?P<fiberid>\d+)?\s+(?P<functionname>[^\s]+)\s+-(?P<message>.+)`
	logre := regexp.MustCompile(logreg)

	file, err := os.Open("data/test_data")

	if err != nil {
		fmt.Println(err)
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
			body, _ := json.Marshal(ads)
			fmt.Println(string(body))
			if !ads.SessionStart {
				fmt.Println("[+] WARNING Found session need before finding a session begin.")
				fmt.Println("[+] WARNING This could be because of rolled logs or parsing errors.")
			}
			ads = AnydeskSession{}
		}

	}

}
