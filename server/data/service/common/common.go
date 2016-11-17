package common

import (
	"strconv"
	"strings"
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
