package time

import (
	"encoding/json"
	"time"
)

// Time wraps stdlib time.Time to customize JSON marshaling.
// When zero, it marshals to an empty string ""; otherwise RFC3339.
type Time time.Time

// MarshalJSON renders zero time as "" and non-zero in RFC3339 format.
func (t Time) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte(`""`), nil
	}

	return json.Marshal((time.Time)(t))
}
