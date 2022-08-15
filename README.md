# anydesk_parser

Attempts to parse Anydesk ad.trace files into something more readable.
It currently doesn't parse out tunneled connections, but does handle files transfered both with the download/upload method as well as using the copy and paste method. Will also call out if data is copied and pasted.

Accepts the input file to process using the --inputFile flags.

Will write output to a file "ad_session_#.json" in the directory defined by --outputDir (default is current directory). The # is the session is from the ad.trace file.

---
## Execution
Select the [releases](https://github.com/nighttardis/anydesk_parser/releases) on the right and download the version that works for your platform.

OR

Clone the repo and run

`go run main.go utils.go --inputFile <path to input file>`