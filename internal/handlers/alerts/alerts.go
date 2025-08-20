package alerts

import (
	"context"
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
	"csjk-bk/pkg/common/utils"
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

// NewGetFiringAlertsAllHandler 创建 getFiringAlertsAll Handler.
func NewGetFiringAlertsAllHandler(client *http.Client, address string) alertsapi.GetFiringAlertsAllHandler {
	return alertsapi.GetFiringAlertsAllHandlerFunc(func(params alertsapi.GetFiringAlertsAllParams) middleware.Responder {
		// 从 Alertmanager 获取全部活跃报警信息
		amUrl := &url.URL{
			Scheme:   "http",
			Host:     address,
			Path:     "/api/v2/alerts",
			RawQuery: "active=true&inhibited=false&silenced=false&unprocessed=false",
		}
		amAlerts, err := getFiringAlertsFromAlertmanager(params.HTTPRequest.Context(), amUrl)
		if err != nil {
			return alertsapi.NewGetFiringAlertsAllInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("无法从报警服务平台(Alertmanager)中获取当前活跃报警信息")})
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
				Status:       "firing",
				Startsat:     strfmt.DateTime(a.StartsAt),
				Endsat:       strfmt.DateTime(time.Time{}),
				Generatorurl: "",
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
			CommonResponse: models.CommonResponse{
				Count:    &total,
				Next:     &nextURI,
				Previous: &prevURI,
				Detail:   utils.StringPtr("获取报警信息成功"),
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

// NewGetFiringAlertsClassificationHandler 创建 getFiringAlertsClassification Handler.
func NewGetFiringAlertsClassificationHandler(client *http.Client, address string) alertsapi.GetFiringAlertsClassificationHandler {
	return alertsapi.GetFiringAlertsClassificationHandlerFunc(func(params alertsapi.GetFiringAlertsClassificationParams) middleware.Responder {
		v := url.Values{}
		v.Set("active", "true")
		v.Set("inhibited", "false")
		v.Set("silenced", "false")
		v.Set("unprocessed", "false")
		v.Set("filter", "class="+params.Classification)

		amURL := &url.URL{
			Scheme:   "http",
			Host:     address,
			Path:     "/api/v2/alerts",
			RawQuery: v.Encode(),
		}

		amAlerts, err := getFiringAlertsFromAlertmanager(params.HTTPRequest.Context(), amURL)
		if err != nil {
			return alertsapi.NewGetFiringAlertsClassificationInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("无法从报警服务平台(Alertmanager)中获取当前活跃报警信息")})
		}

		sort.Slice(amAlerts, func(i, j int) bool {
			return amAlerts[i].StartsAt.After(amAlerts[j].StartsAt)
		})

		stats := map[string]map[string]int64{
			params.Classification: {},
		}
		for _, a := range amAlerts {
			severity := a.Labels["severity"]
			stats[params.Classification][severity]++
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
				Status:       "firing",
				Startsat:     strfmt.DateTime(a.StartsAt),
				Endsat:       strfmt.DateTime(time.Time{}),
				Generatorurl: "",
				Labels:       a.Labels,
				Annotaions:   a.Annotations,
			})
		}

		path := params.HTTPRequest.URL.Path
		var nextURI, prevURI strfmt.URI
		if end < len(amAlerts) {
			v := url.Values{}
			v.Set("classification", params.Classification)
			v.Set("page", strconv.FormatInt(params.Page+1, 10))
			v.Set("page_size", strconv.FormatInt(params.PageSize, 10))
			nextURI = strfmt.URI(path + "?" + v.Encode())
		}
		if params.Page > 1 {
			v := url.Values{}
			v.Set("classification", params.Classification)
			v.Set("page", strconv.FormatInt(params.Page-1, 10))
			v.Set("page_size", strconv.FormatInt(params.PageSize, 10))
			prevURI = strfmt.URI(path + "?" + v.Encode())
		}

		payload := &alertsapi.GetFiringAlertsClassificationOKBody{
			CommonResponse: models.CommonResponse{
				Count:    &total,
				Next:     &nextURI,
				Previous: &prevURI,
				Detail:   utils.StringPtr("获取报警信息成功"),
			},
			Results: &alertsapi.GetFiringAlertsClassificationOKBodyGetFiringAlertsClassificationOKBodyAO1Results{
				Statistic: stats,
				Alerts:    alertsModel,
			},
		}

		return alertsapi.NewGetFiringAlertsClassificationOK().WithPayload(payload)
	})
}

type Alerts []Alert

type Alert struct {
	Fingerprint string            `json:"fingerprint"`
	StartsAt    time.Time         `json:"startsAt"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// getFiringAlertsFromAlertmanager 从 Alertmanager 获取全部 Active 报警信息.
func getFiringAlertsFromAlertmanager(ctx context.Context, amUrl *url.URL) (Alerts, error) {
	alerts := make(Alerts, 0)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, amUrl.String(), nil)
	if err != nil {
		return alerts, fmt.Errorf("无法创建请求体: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return alerts, fmt.Errorf("无法执行请求(%s): %w", req.URL.String(), err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return alerts, fmt.Errorf("无法解析响应数据: %w", err)
	}

	return alerts, nil
}
