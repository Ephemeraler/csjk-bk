package postgres

import (
	"context"
	"testing"
	"time"
)

func TestGetAlerts(t *testing.T) {
	client := Client{}
	cond := Conditions{
		Status: []string{"firing"},
		Start:  time.Now().Add(-1 * time.Hour),
		End:    time.Now(),
		Labels: map[string][]string{
			"severity": {"debug", "info"},
			"cluster":  {"cluster-1", "cluster-2"},
		},
	}
	client.GetAlerts(context.Background(), cond)
}
