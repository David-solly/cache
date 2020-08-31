package cache

import "time"

//Create type
type Record struct {
	Key   string `json:"deveui,omitempty"`
	Value string `json:"shortcode,omitempty"`
}

type RecordExpirer struct {
	Key     string
	Value   string
	Timeout time.Duration
}

func (eo *RecordExpirer) GetExpiringRecord() (k, v string, d time.Duration) {
	return eo.Key, eo.Value, eo.Timeout
}

type Expirer interface {
	GetExpiringRecord() (k, v string, d time.Duration)
}
