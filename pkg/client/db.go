package client

import (
	"context"
	"csjk-bk/pkg/client/postgres"
	"csjk-bk/pkg/model/alert"
)

type DB interface {
	GetAlerts(ctx context.Context, cond postgres.Conditions) (alert.Alerts, error)
}
