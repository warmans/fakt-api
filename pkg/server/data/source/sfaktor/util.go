package sfaktor

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goodsign/monday"
)

//ParseTime parses times in the format e.g. "Montag, 21.12.2015", "21:00 Uhr"
func ParseTime(dateString, timeString string, localTime *time.Location) (time.Time, error) {

	if strings.Contains(dateString, ",") {
		cleanDate := fmt.Sprintf("%s %s", timeString, strings.TrimSpace(strings.Split(dateString, ",")[1]))
		return monday.ParseInLocation("15:04 02. January 2006", cleanDate, localTime, monday.LocaleDeDE)
	}

	return time.Time{}, errors.New("header date format looks wrong")
}

//HTML Parsing partially stolen from: https://github.com/kennygrant/sanitize/blob/master/sanitize.go
func StripHTML(s string) string {
	output := ""
	if !strings.ContainsAny(s, "<>") {
		output = s
	} else {

		s = strings.Replace(s, "\n", " ", -1)

		// Walk through the string removing all tags
		b := bytes.NewBufferString("")
		inTag := false
		for _, r := range s {
			switch r {
			case '<':
				inTag = true
			case '>':
				inTag = false
			default:
				if !inTag {
					b.WriteRune(r)
				}
			}
		}
		output = b.String()
	}
	return output
}
