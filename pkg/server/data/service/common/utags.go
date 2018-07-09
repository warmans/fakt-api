package common

type UTags struct {
	Username string   `json:"username" db:"username"`
	Values   []string `json:"tags" db:"tags"`
}

func (u *UTags) HasValue(value string) bool {
	for _, val := range u.Values {
		if value == val {
			return true
		}
	}
	return false
}

type UTagsFilter struct {
	Username string `json:"username"`
}
