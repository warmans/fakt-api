package common

type Tag struct {
	ID             int64  `json:"id"`
	Tag            string `json:"tag"`
	StatPerformers int64  `json:"stat_performers"`
	StatEvents     int64  `json:"stat_events"`
}
