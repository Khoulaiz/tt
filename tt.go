package main

import (
	"bufio"
	"log"
	"os"
	"time"
	"strings"
	"sort"
	"strconv"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	from         = kingpin.Flag("from", "Time layout to convert from, specified in reference time - see https://golang.org/src/time/format.go#L61").Required().String()
	to           = kingpin.Flag("to", "Time layout to convert to, specified in reference time - see https://golang.org/src/time/format.go#L61").Required().String()
	toTimeZone   = kingpin.Flag("totimezone", "Timezone to convert date to (local zone is default)").Default("Local").String()
	fromTimeZone = kingpin.Flag("fromtimezone", "Timezone to parse date from (UTC is default)").Default("UTC").String()
	toConvert    = kingpin.Arg("timestrings", "Time values to convert. Use `now` for current time.").Required().Strings()
)

// predefined formats

var formats = map[string]string{
		"TIMESTAMP"    : "TIMESTAMP",
		"TIMESTAMPNANO": "TIMESTAMPNANO",
    	"ANSIC"		   : time.ANSIC,
		"UnixDate"     : time.UnixDate,
		"RubyDate"     : time.RubyDate,
		"RFC822"       : time.RFC822,
		"RFC822Z"      : time.RFC822Z,
		"RFC850"       : time.RFC850,
		"RFC1123"      : time.RFC1123,
		"RFC1123Z"     : time.RFC1123Z,
		"RFC3339"      : time.RFC3339,
		"RFC3339Nano"  : time.RFC3339Nano,
		"Kitchen"      : time.Kitchen,
		"Stamp"        : time.Stamp,
		"StampMilli"   : time.StampMilli,
		"StampMicro"   : time.StampMicro,
		"StampNano"    : time.StampNano,
	}

// takes file from stdin and outputs to stdout in the correct time format, one date per line
func main() {

	var formatkeys []string
	for k := range formats {
    	formatkeys = append(formatkeys, k)
	}
	sort.Strings(formatkeys)
	kingpin.CommandLine.Help = "To specify an `from` or `to` pattern, you can use one of the given predefined formats or " +
								"define an example using the reference date: 'Mon Jan 2 15:04:05 MST 2006'. See https://golang.org/src/time/format.go#L61 " +
		                        "for details. The predefined patterns are: " + strings.Join(formatkeys, ", ")
	kingpin.Parse()
	var toTZ *time.Location
	var fromTZ *time.Location
	var err error
	lineWriter := bufio.NewWriter(os.Stdout)
	defer lineWriter.Flush()
	if *toTimeZone != "" {
		toTZ, err = parseTimeZome(*toTimeZone)
	}
	fromTZ, err = parseTimeZome(*fromTimeZone)
	if formats[*from] != "" {
		*from = formats[*from]
	}
	if formats[*to] != "" {
		*to = formats[*to]
	}
	for _, dateString := range *toConvert {
		var date time.Time
		if dateString == "now" {
			date = time.Now()
		} else {
			switch *from {
			case "TIMESTAMP":
				seconds, err := strconv.ParseInt(dateString, 10, 64)
				if err != nil {
					log.Fatalf("Error parsing input timestamp (max 10 digits expected) - %v", err)
				}
				date = time.Unix(seconds, 0)
			case "TIMESTAMPNANO":
				nanos, err := strconv.ParseInt(dateString[len(dateString)-9:], 10, 64)
				if err != nil {
					log.Fatalf("Error parsing input timestamp (max 19 digits expected) - %v", err)
				}
				seconds, err := strconv.ParseInt(dateString[0:len(dateString)-9], 10, 64)
				if err != nil {
					log.Fatalf("Error parsing input timestamp (max 19 digits expected) - %v", err)
				}
				date = time.Unix(seconds, nanos)
			default:
				date, err = time.ParseInLocation(*from, dateString, fromTZ)
				if err != nil {
					log.Fatalf("Error parsing input time - %v", err)
				}
			}
		}
		if *toTimeZone != "" {
			date = date.In(toTZ)
		}
		switch *to {
		case "TIMESTAMP":
			_, err = lineWriter.WriteString(strconv.FormatInt(date.Unix(),10))
		case "TIMESTAMPNANO":
			_, err = lineWriter.WriteString(strconv.FormatInt(date.UnixNano(),10))
		default:
			_, err = lineWriter.WriteString(date.Format(*to))
		}
		if err != nil {
			log.Fatalf("Error writing output time - %v", err)
		}
		lineWriter.WriteString("\n")
	}
}

func parseTimeZome(tzString string) (*time.Location, error) {
	fromTZ, err := time.LoadLocation(tzString)
	if err != nil {
		log.Fatalf("Error loading timezone - %s - %v", *fromTimeZone, err)
	}
	return fromTZ, err
}
