package alerts

import (
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
