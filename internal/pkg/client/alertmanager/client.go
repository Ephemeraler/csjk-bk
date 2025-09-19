package alertmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer   Doer
	server *url.URL
	logger *slog.Logger
}

// SetCilent 仅用于测试阶段使用
func (c *Client) SetClient(Doer Doer, server *url.URL, logger *slog.Logger) {
	c.doer = Doer
	if Doer == nil {
		c.doer = http.DefaultClient
	}
	c.server = server
	c.logger = logger
	if logger == nil {
		c.logger = slog.Default()
	}
}

type Alerts []Alert
type Alert struct {
	Fingerprint  string            `json:"fingerprint"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
}

func (c *Client) GetActiveAlerts(ctx context.Context) (Alerts, error) {
	alerts := make(Alerts, 0)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.server.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request to get active alerts from alertmanager", "err", err, "url", c.server.String())
		return alerts, fmt.Errorf("unable to create request to get active alerts from alertmanager")
	}
	resp, err := c.doer.Do(req)
	if err != nil {
		c.logger.Error("unable to send request to alertmanager", "err", err, "url", c.server.String())
		return alerts, fmt.Errorf("unbale to send request to alertmanager")
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		c.logger.Error("unable to decode response body of alertmanager", "err", err)
		return alerts, fmt.Errorf("unable to decode response body of alertmanager")
	}

	return alerts, nil
}
