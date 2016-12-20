package common

import (
	"strconv"
	"strings"
	"time"
	"github.com/jinzhu/now"
)

const DATE_FORMAT_SQL = "2006-01-02 15:04:05.999999999-07:00"

func IfOrInt(val bool, trueVal, falseVal int) int {
	if val {
		return trueVal
	}
	return falseVal
}

func SplitConcatIDs(concatIDs string, delimiter string) []int {
	performerIDs := []int{}
	for _, pidStr := range strings.Split(concatIDs, ",") {
		if pidInt, _ := strconv.Atoi(pidStr); pidInt != 0 {
			performerIDs = append(performerIDs, pidInt)
		}
	}
	return performerIDs
}


//GetRelativeDateRange takes e.g. this weekend and returns the start and end date in SQL format
func GetRelativeDateRange(name string) (time.Time, time.Time) {

	//end days a few hours past midnight since e.g. 1am Saturday should still count as Friday night
	switch strings.ToLower(name) {
	case "this week":
		return time.Now(), now.EndOfSunday().Add(time.Hour * 4)
	case "this weekend":
		return now.BeginningOfWeek().Add(time.Hour * 24 * 5), now.EndOfSunday().Add(time.Hour * 4)
	case "tomorrow":
		return now.BeginningOfDay().Add(time.Hour * 24), now.EndOfDay().Add(time.Hour * 28)
	default:
		//unknown values including "today" get today-ish
		return now.BeginningOfDay(), now.EndOfDay().Add(time.Hour * 4)
	}
}