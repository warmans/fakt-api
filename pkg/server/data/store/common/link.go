package common

type Link struct {
	URI  string `json:"uri" db:"link"`
	Type string `json:"type" db:"link_type"`
	Text string `json:"text" db:"link_description"`
}
