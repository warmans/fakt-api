package common

type Venue struct {
	ID      int64      `json:"id"`
	Name    string     `json:"name"`
	Address string     `json:"address"`  //todo rename do location and create a new address field with actual address
	LatLong [2]float64 `json:"lat_long"` //todo
	Info    string     `json:"info"`     //todo
	Img     string     `json:"img"`      //todo
	Links   []string   `json:"link"`     //todo
}

func (v *Venue) IsValid() bool {
	if v.Name == "" {
		return false
	}
	return true
}

