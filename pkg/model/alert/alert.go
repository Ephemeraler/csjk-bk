package alert

import "time"

type Alerts []*Alerts

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
