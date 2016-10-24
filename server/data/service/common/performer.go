package common

type Performer struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Info       string            `json:"info"`
	Genre      string            `json:"genre"`
	Home       string            `json:"home"`
	Img        string            `json:"img"`
	ListenURL  string            `json:"listen_url"`
	Activity   string            `json:"activity"`   //todo: e.g. high/medium/low based on number of gigs within last X days
	Popularity float64           `json:"popularity"` //toodo:figure out based on banccamp downloads etc?
	Events     []*Event          `json:"event,omitempty"`
	Links      []*Link           `json:"link,omitempty"`
	Tags       []string          `json:"tag"`
	UTags      []UTags           `json:"utag,omitempty"`
	Images     map[string]string `json:"images"`
}

func (p *Performer) IsValid() bool {
	if p.Name == "" || p.Genre == "" {
		return false
	}
	return true
}
