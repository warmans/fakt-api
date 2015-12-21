package crawler

import (
	"time"
	"strings"
	"github.com/vektra/errors"
	"fmt"
)

//ParseTime parses times in the format e.g. "Montag, 21.12.2015", "21:00 Uhr"
func ParseTime(dateString, timeString string) (time.Time, error) {

	if strings.Contains(dateString, ",") {
		cleanDate := fmt.Sprintf(
			"%s %s",
			strings.Trim(timeString, " Uhr"),
			strings.TrimSpace(strings.Split(dateString, ",")[1]),
		)
		return time.Parse("15.04 02.01.2006", cleanDate)
	}

	return time.Time{}, errors.New("header date format looks wrong")
}
