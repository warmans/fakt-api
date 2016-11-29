package common

import (
	"crypto/md5"
	"encoding/hex"
)

type Performer struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Info       string            `json:"info"`
	Genre      string            `json:"genre"`
	Home       string            `json:"home"`
	ListenURL  string            `json:"listen_url"`
	Activity   string            `json:"activity"`   //todo: e.g. high/medium/low based on number of gigs within last X days
	Popularity float64           `json:"popularity"` //todo:figure out based on banccamp downloads etc?
	Events     []*Event          `json:"event,omitempty"`
	Links      []*Link           `json:"link,omitempty"`
	Tags       []string          `json:"tag"`
	UTags      []UTags           `json:"utag,omitempty"`
	Images     map[string]string `json:"images"`
	EmbedURL   string            `json:"embed_url"`
}

func (p *Performer) IsValid() bool {
	if p.Name == "" || p.Genre == "" {
		return false
	}
	return true
}

func (p *Performer) GetNameHash() string {
	if p.Name == "" {
		return ""
	}
	hasher := md5.New()
	hasher.Write([]byte(p.Name))
	return hex.EncodeToString(hasher.Sum(nil)[0:4]) //use only a short hash since the dataset is so small
}
