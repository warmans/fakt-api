package entity

import "time"

type Venue struct {
	ID      int64     `json:id`
	Name    string    `json:name`
	Address string    `json:address`
}

type Event struct {
	ID          int64     `json:id`
	Date        time.Time `json:date`
	Venue       *Venue    `json:venue`
	Type        string    `json:type`
	Description string    `json:description`
}
