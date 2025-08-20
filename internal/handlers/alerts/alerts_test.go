package alerts

import (
	"context"
	"fmt"
	"net/url"
	"testing"
)

func TestGetFiringAlertsFromAlertmanager(t *testing.T) {
	u := &url.URL{
		Scheme:   "http",
		Host:     "192.168.2.35:9093",
		Path:     "/api/v2/alerts",
		RawQuery: "active=true&inhibited=false&silenced=false&unprocessed=false",
	}

	alerts, _ := getFiringAlertsFromAlertmanager(context.Background(), u)
	fmt.Printf("%+v\n", alerts)
}
