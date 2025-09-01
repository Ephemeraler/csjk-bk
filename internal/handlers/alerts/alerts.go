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
	"csjk-bk/pkg/client/postgres"
	"csjk-bk/pkg/common/utils"
	alertsapi "csjk-bk/restapi/operations/alert"
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

// NewGetFiringAlertsHandler 创建 getFiringAlertsAll Handler.
func NewGetFiringAlertsHandler(client *http.Client, address string) alertsapi.GetFiringAlertsHandler {
	return alertsapi.GetFiringAlertsHandlerFunc(func(params alertsapi.GetFiringAlertsParams) middleware.Responder {
		// 从 Alertmanager 获取全部活跃报警信息
		amUrl := &url.URL{
			Scheme:   "http",
			Host:     address,
			Path:     "/api/v2/alerts",
			RawQuery: "active=true&inhibited=false&silenced=false&unprocessed=false",
		}
		if len(params.Filter) != 0 {
			q := amUrl.Query()
			for _, v := range params.Filter {
				q.Add("filter", v)
			}
			amUrl.RawQuery = q.Encode()
		}

		amAlerts, err := getFiringAlertsFromAlertmanager(params.HTTPRequest.Context(), amUrl)
		if err != nil {
			return alertsapi.NewGetFiringAlertsInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("无法从报警服务平台(Alertmanager)中获取当前活跃报警信息")})
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

		payload := &alertsapi.GetFiringAlertsOKBody{
			CommonResponse: models.CommonResponse{
				Count:    &total,
				Next:     &nextURI,
				Previous: &prevURI,
				Detail:   utils.StringPtr("获取报警信息成功"),
			},
			Results: &alertsapi.GetFiringAlertsOKBodyGetFiringAlertsOKBodyAO1Results{
				SeverityCount: &alertsapi.GetFiringAlertsOKBodyGetFiringAlertsOKBodyAO1ResultsSeverityCount{
					Inband:  stats.Inband,
					Outband: stats.Outband,
					Event:   stats.Event,
				},
				Alerts: alertsModel,
			},
		}

		return alertsapi.NewGetFiringAlertsOK().WithPayload(payload)
	})
}

func NewPostAlertHistory(client *postgres.Client) alertsapi.PostAlertHistoryHandler {
	return alertsapi.PostAlertHistoryHandlerFunc(func(pahp alertsapi.PostAlertHistoryParams) middleware.Responder {
		ctx := pahp.HTTPRequest.Context()

		start := time.Time(pahp.Body.Start)
		end := time.Time(pahp.Body.End)
		if !start.IsZero() && !end.IsZero() && start.After(end) {
			return alertsapi.NewPostAlertHistoryBadRequest().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("开始时间不能晚于结束时间")})
		}

		alerts, err := client.GetAlerts(ctx, start, end, pahp.Body.Status, pahp.Body.Labels, pahp.Body.Annotations, pahp.Body.Page, pahp.Body.PageSize)
		if err != nil {
			return alertsapi.NewPostAlertHistoryInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("获取报警历史信息失败")})
		}

		var alertsModel models.Alerts
		for _, a := range alerts {
			alertsModel = append(alertsModel, &models.Alert{
				ID:         a.ID,
				Status:     a.Status,
				Startsat:   strfmt.DateTime(a.StartsAt),
				Endsat:     strfmt.DateTime(a.EndsAt),
				Responder:  a.Responder,
				Operation:  a.Operation,
				Labels:     a.Label,
				Annotaions: a.Annotation,
			})
		}

		total := int64(len(alerts))
		var emptyURI strfmt.URI
		payload := &alertsapi.PostAlertHistoryOKBody{
			CommonResponse: models.CommonResponse{
				Count:    &total,
				Next:     &emptyURI,
				Previous: &emptyURI,
				Detail:   utils.StringPtr("获取报警信息成功"),
			},
			Alerts: alertsModel,
		}

		return alertsapi.NewPostAlertHistoryOK().WithPayload(payload)
	})
}

func NewGetAlertLabelNames(client *postgres.Client) alertsapi.GetAlertLabelNamesHandler {
	return alertsapi.GetAlertLabelNamesHandlerFunc(func(galnp alertsapi.GetAlertLabelNamesParams) middleware.Responder {
		ctx := galnp.HTTPRequest.Context()

		labels, err := client.GetAlertLabels(ctx)
		if err != nil {
			return alertsapi.NewGetAlertLabelNamesInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("获取报警标签名称失败")})
		}

		total := int64(len(labels))
		var emptyURI strfmt.URI
		payload := &alertsapi.GetAlertLabelNamesOKBody{
			CommonResponse: models.CommonResponse{
				Count:    &total,
				Next:     &emptyURI,
				Previous: &emptyURI,
				Detail:   utils.StringPtr("获取报警标签名称成功"),
			},
			Results: labels,
		}

		return alertsapi.NewGetAlertLabelNamesOK().WithPayload(payload)
	})
}

func NewGetAlertLabelValues(client *postgres.Client) alertsapi.GetAlertLabelValuesHandlerFunc {
	return alertsapi.GetAlertLabelValuesHandlerFunc(func(galp alertsapi.GetAlertLabelValuesParams) middleware.Responder {
		ctx := galp.HTTPRequest.Context()

		values, err := client.GetAlertLabelValues(ctx, galp.Labelname)
		if err != nil {
			return alertsapi.NewGetAlertLabelValuesInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("获取报警标签值失败")})
		}

		total := int64(len(values))
		var emptyURI strfmt.URI
		payload := &alertsapi.GetAlertLabelValuesOKBody{
			CommonResponse: models.CommonResponse{
				Count:    &total,
				Next:     &emptyURI,
				Previous: &emptyURI,
				Detail:   utils.StringPtr("获取报警标签值成功"),
			},
			Results: values,
		}

		return alertsapi.NewGetAlertLabelValuesOK().WithPayload(payload)
	})
}

// NewPutAlertOperation 更新报警处理人和处理方式.
func NewPutAlertOperation(client *postgres.Client) alertsapi.PutAlertOperationHandler {
	return alertsapi.PutAlertOperationHandlerFunc(func(pao alertsapi.PutAlertOperationParams) middleware.Responder {
		ctx := pao.HTTPRequest.Context()

		if err := client.SetOperation(ctx, *pao.Body.Ids, *pao.Body.Operation, *pao.Body.Responder); err != nil {
			return alertsapi.NewPutAlertOperationInternalServerError().WithPayload(&models.StandardResponse{Detail: utils.StringPtr("更新报警处理信息失败")})
		}

		var emptyURI strfmt.URI
		count := int64(-1)
		payload := &models.StandardResponse{
			Count:    &count,
			Next:     &emptyURI,
			Previous: &emptyURI,
			Results:  struct{}{},
			Detail:   utils.StringPtr("更新报警处理信息成功"),
		}

		return alertsapi.NewPutAlertOperationOK().WithPayload(payload)
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
