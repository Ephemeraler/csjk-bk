package alerts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"csjk-bk/models"
	alertsapi "csjk-bk/restapi/operations/alerts"
)

type amAlert struct {
	Fingerprint string `json:"fingerprint"`
	Status      struct {
		State string `json:"state"`
	} `json:"status"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
}

// NewGetFiringAlertsAllHandler creates handler for retrieving firing alerts.
func NewGetFiringAlertsAllHandler(client *http.Client, alertManagerURL string) alertsapi.GetFiringAlertsAllHandler {
	return alertsapi.GetFiringAlertsAllHandlerFunc(func(params alertsapi.GetFiringAlertsAllParams) middleware.Responder {
		reqURL := fmt.Sprintf("%s/api/v2/alerts?active=true&inhibited=false&silenced=false&unprocessed=false", alertManagerURL)
		req, err := http.NewRequestWithContext(params.HTTPRequest.Context(), http.MethodGet, reqURL, nil)
		if err != nil {
			return alertsapi.NewGetFiringAlertsAllInternalServerError().WithPayload(&models.StandardResponse{Detail: err.Error()})
		}

		resp, err := client.Do(req)
		if err != nil {
			return alertsapi.NewGetFiringAlertsAllInternalServerError().WithPayload(&models.StandardResponse{Detail: err.Error()})
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return alertsapi.NewGetFiringAlertsAllInternalServerError().WithPayload(&models.StandardResponse{Detail: fmt.Sprintf("alertmanager status %s", resp.Status)})
		}

		var amAlerts []amAlert
		if err := json.NewDecoder(resp.Body).Decode(&amAlerts); err != nil {
			return alertsapi.NewGetFiringAlertsAllInternalServerError().WithPayload(&models.StandardResponse{Detail: err.Error()})
		}

		sort.Slice(amAlerts, func(i, j int) bool {
			return amAlerts[i].StartsAt.After(amAlerts[j].StartsAt)
		})

		stats := struct {
			Inband  map[string]int64
			Outband map[string]int64
			Event   map[string]int64
		}{
			Inband:  make(map[string]int64),
			Outband: make(map[string]int64),
			Event:   make(map[string]int64),
		}

		for _, a := range amAlerts {
			class := a.Labels["class"]
			severity := a.Labels["severity"]
			switch class {
			case "inband":
				stats.Inband[severity]++
			case "outband":
				stats.Outband[severity]++
			case "event":
				stats.Event[severity]++
			}
		}

		total := int64(len(amAlerts))
		start := int((params.Page - 1) * params.PageSize)
		end := start + int(params.PageSize)
		if start > len(amAlerts) {
			start = len(amAlerts)
		}
		if end > len(amAlerts) {
			end = len(amAlerts)
		}

		var alertsModel models.Alerts
		for _, a := range amAlerts[start:end] {
			alertsModel = append(alertsModel, &models.Alert{
				Fingerprint:  a.Fingerprint,
				Status:       a.Status.State,
				Startsat:     strfmt.DateTime(a.StartsAt),
				Endsat:       strfmt.DateTime(a.EndsAt),
				Generatorurl: a.GeneratorURL,
				Labels:       a.Labels,
				Annotaions:   a.Annotations,
			})
		}

		path := params.HTTPRequest.URL.Path
		var nextURI, prevURI strfmt.URI
		if end < len(amAlerts) {
			v := url.Values{}
			v.Set("page", strconv.FormatInt(params.Page+1, 10))
			v.Set("page_size", strconv.FormatInt(params.PageSize, 10))
			nextURI = strfmt.URI(path + "?" + v.Encode())
		}
		if params.Page > 1 {
			v := url.Values{}
			v.Set("page", strconv.FormatInt(params.Page-1, 10))
			v.Set("page_size", strconv.FormatInt(params.PageSize, 10))
			prevURI = strfmt.URI(path + "?" + v.Encode())
		}

		payload := &alertsapi.GetFiringAlertsAllOKBody{
			StandardResponse: models.StandardResponse{
				Count:    &total,
				Next:     nextURI,
				Previous: prevURI,
			},
			Results: &alertsapi.GetFiringAlertsAllOKBodyGetFiringAlertsAllOKBodyAO1Results{
				Statistic: &alertsapi.GetFiringAlertsAllOKBodyGetFiringAlertsAllOKBodyAO1ResultsStatistic{
					Inband:  stats.Inband,
					Outband: stats.Outband,
					Event:   stats.Event,
				},
				Alerts: alertsModel,
			},
		}

		return alertsapi.NewGetFiringAlertsAllOK().WithPayload(payload)
	})
}
