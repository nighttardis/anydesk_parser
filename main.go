package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func main() {
	logreg := `^\s+(?P<loglevel>[^\s]+)\s(?P<datetime>[\d-]+\s[\d:\.]+)\s+(?P<appname>\w+)\s+(?P<threadname>[a-z][^\s]+)?\s+(?P<pid>\d+)\s+(?P<threadid>\d+)\s+(?P<fiberid>\d+)?\s+(?P<functionname>[^\s]+)\s+-(?P<message>.+)`
	logre := regexp.MustCompile(logreg)

	file, err := os.Open("data/ad.trace_windows_user_approval")

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
				le[name] = match[i]
			}
		}
		jsonbody, _ := json.Marshal(le)
		leS := LogEntry{}
		json.Unmarshal(jsonbody, &leS)

		leS.parseFunction(&ads)

		if ads.SessionStart {
			ads.LogEntries = append(ads.LogEntries, leS)
		}

		fmt.Println(ads)

		if ads.SessionEnd {
			ads = AnydeskSession{}
		}

	}

}
