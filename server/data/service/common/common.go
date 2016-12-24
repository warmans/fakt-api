package common

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
)

const DateFormatSQL = "2006-01-02 15:04:05.999999999-07:00"

const DefaultPageSize = 50

func IfOrInt(val bool, trueVal, falseVal int) int {
	if val {
		return trueVal
	}
	return falseVal
}

func SplitConcatIDs(concatIDs string, delimiter string) []int64 {
	performerIDs := []int64{}
	for _, pidStr := range strings.Split(concatIDs, delimiter) {
		if pidInt, _ := strconv.Atoi(pidStr); pidInt != 0 {
			performerIDs = append(performerIDs, int64(pidInt))
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

type CommonFilter struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
	IDs      []int64 `json:"ids"`
}

func (f *CommonFilter) Populate(r *http.Request) {
	if page := r.Form.Get("page"); page != "" {
		if pageInt, err := strconv.Atoi(page); err == nil {
			f.Page = int64(pageInt)
		}
	}
	f.PageSize = DefaultPageSize
	if pageSize := r.Form.Get("page_size"); pageSize != "" {
		if pageSizeInt, err := strconv.Atoi(pageSize); err == nil {
			if pageSizeInt > 0 {
				f.PageSize = int64(pageSizeInt)
			}
		}
	}

	f.IDs = make([]int64, 0)
	if tagIds := r.Form.Get("ids"); tagIds != "" {
		for _, idStr := range strings.Split(tagIds, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				f.IDs = append(f.IDs, int64(idInt))
			}
		}
	}
}
