package entity

import "time"

type Venue struct {
	ID      string    `json:id`
	Name    string    `json:name`
	Address string    `json:address`
}

type Event struct {
	ID          int       `json:id`
	Date        time.Time `json:date`
	Venue       *Venue    `json:venue`
	Type        string    `json:type`
	Description string    `json:description`
}
