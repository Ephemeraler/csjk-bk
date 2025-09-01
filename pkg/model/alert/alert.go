package alert

import "time"

// Alerts represents a list of Alert pointers.
type Alerts []*Alert

type Alert struct {
	ID         int64
	Status     string
	StartsAt   time.Time
	EndsAt     time.Time
	Responder  string
	Operation  string
	Label      map[string]string
	Annotation map[string]string
}
