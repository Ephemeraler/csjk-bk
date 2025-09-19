package alert

import (
	"csjk-bk/internal/pkg/common/paging"
	"csjk-bk/internal/pkg/response"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AlertsFiringResults struct {
	FiringStatistic FiringStatistic `json:"severity_count"` // 统计信息
	Alerts          Alerts          `json:"alerts"`         // 报警信息
}

type FiringStatistic struct {
	InBand  map[string]int `json:"inband"`  // 带内报警统计信息
	OutBand map[string]int `json:"outband"` // 带外报警统计信息
	Event   map[string]int `json:"event"`   // 时间报警统计信息
}

type Alerts []Alert // 报警信息
type Alert struct {
	ID           int               `json:"id,omitemoty"`        // ID, 活跃报警中无该字段, 历史报警中才会存在, 对应数据库中 alert 表 ID.
	Fingerprint  string            `json:"fingerprint"`         // 报警指纹
	Status       string            `json:"status"`              // 报警状态
	StartsAt     time.Time         `json:"startsat"`            // 报警开始时间
	EndsAt       time.Time         `json:"endsat"`              // 报警结束时间
	Generatorurl string            `json:"generatorurl"`        // 报警产生来源
	Responder    string            `json:"responder,omitemoty"` // 报警处理人
	Operation    string            `json:"operation,omitemoty"` // 报警处理办法
	Lables       map[string]string `json:"labels"`              // 报警标签
	Annotations  map[string]string `json:"annotations"`         // 报警注释
}

// HandlerGetAlertsFiring 从 Alertmanager 中获取实时报警信息.
// @Summary 获取实时报警信息.
// @Description 从报警平台(alertmanager)中获取实时报警信息, 请求参数 class 对应分屏设计. 不分屏时, class=all. 分屏时 class="<对应系统>"
// @Tags 报警, 实时报警
// @Produce json
// @Param class query string false "报警类别, 支持 inband, outband, event, all" Enums(inband, outband, event, all) default(all)
// @Param paging query bool false "启动分页, 当前仅支持强制分页" Enums(true) default(true)
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页报警数量, 最大仅支持100" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.Response{results=AlertsFiringResults}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/alerts/firing [get]
func (rt *Router) HandlerGetAlertsFiring(c *gin.Context) {
	class := c.DefaultQuery("class", "all")
	class = strings.ToLower(class)
	// 默认开启分页，并设置默认页码与每页数量
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.Paging = true
	pq.SetDefaults(1, 20, 100)

	// 预置返回结构
	results := AlertsFiringResults{
		FiringStatistic: FiringStatistic{InBand: map[string]int{}, OutBand: map[string]int{}, Event: map[string]int{}},
		Alerts:          make(Alerts, 0),
	}

	// TODO: 测试阶段使用固定地址（注意修正为有效查询串）
	server, _ := url.Parse("http://192.168.2.35:9093/api/v2/alerts?active=true&inhibited=false&silenced=false&unprocessed=false")
	rt.amClient.SetClient(nil, server, rt.logger)

	// 拉取当前 firing 报警
	alertsFromAlertmanager, err := rt.amClient.GetActiveAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Results: results, Detail: "无法从报警平台获取报警数据"})
		return
	}

	// 按开始时间降序排序
	sort.Slice(alertsFromAlertmanager, func(i, j int) bool {
		return alertsFromAlertmanager[i].StartsAt.After(alertsFromAlertmanager[j].StartsAt)
	})

	switch class {
	case "all":
		for _, alert := range alertsFromAlertmanager {
			alertClass, ok := alert.Labels["class"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('class') noexist", "fingerprint", alert.Fingerprint)
				continue
			}
			alertSeverity, ok := alert.Labels["severity"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('severity') noexist", "fingerprint", alert.Fingerprint)
				continue
			}

			switch strings.ToLower(strings.TrimSpace(alertClass)) {
			case "inband":
				if _, ok := results.FiringStatistic.InBand[alertSeverity]; !ok {
					results.FiringStatistic.InBand[alertSeverity] = 1
				} else {
					results.FiringStatistic.InBand[alertSeverity] += 1
				}
			case "outband":
				if _, ok := results.FiringStatistic.OutBand[alertSeverity]; !ok {
					results.FiringStatistic.OutBand[alertSeverity] = 1
				} else {
					results.FiringStatistic.OutBand[alertSeverity] += 1
				}
			case "event":
				if _, ok := results.FiringStatistic.Event[alertSeverity]; !ok {
					results.FiringStatistic.Event[alertSeverity] = 1
				} else {
					results.FiringStatistic.Event[alertSeverity] += 1
				}
			default:
			}
			results.Alerts = append(results.Alerts, Alert{
				Fingerprint:  alert.Fingerprint,
				Status:       "firing",
				StartsAt:     alert.StartsAt,
				EndsAt:       alert.EndsAt,
				Generatorurl: alert.GeneratorURL,
				Lables:       alert.Labels,
				Annotations:  alert.Annotations,
			})
		}
	case "inband":
		for _, alert := range alertsFromAlertmanager {
			alertClass, ok := alert.Labels["class"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('class') noexist", "fingerprint", alert.Fingerprint)
				continue
			}
			alertSeverity, ok := alert.Labels["severity"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('severity') noexist", "fingerprint", alert.Fingerprint)
				continue
			}

			switch strings.ToLower(strings.TrimSpace(alertClass)) {
			case "inband":
				if _, ok := results.FiringStatistic.InBand[alertSeverity]; !ok {
					results.FiringStatistic.InBand[alertSeverity] = 1
				} else {
					results.FiringStatistic.InBand[alertSeverity] += 1
				}
			default:
			}
			results.Alerts = append(results.Alerts, Alert{
				Fingerprint:  alert.Fingerprint,
				Status:       "firing",
				StartsAt:     alert.StartsAt,
				EndsAt:       alert.EndsAt,
				Generatorurl: alert.GeneratorURL,
				Lables:       alert.Labels,
				Annotations:  alert.Annotations,
			})
		}
	case "outband":
		for _, alert := range alertsFromAlertmanager {
			alertClass, ok := alert.Labels["class"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('class') noexist", "fingerprint", alert.Fingerprint)
				continue
			}
			alertSeverity, ok := alert.Labels["severity"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('severity') noexist", "fingerprint", alert.Fingerprint)
				continue
			}

			switch strings.ToLower(strings.TrimSpace(alertClass)) {
			case "outband":
				if _, ok := results.FiringStatistic.OutBand[alertSeverity]; !ok {
					results.FiringStatistic.OutBand[alertSeverity] = 1
				} else {
					results.FiringStatistic.OutBand[alertSeverity] += 1
				}
			default:
			}
			results.Alerts = append(results.Alerts, Alert{
				Fingerprint:  alert.Fingerprint,
				Status:       "firing",
				StartsAt:     alert.StartsAt,
				EndsAt:       alert.EndsAt,
				Generatorurl: alert.GeneratorURL,
				Lables:       alert.Labels,
				Annotations:  alert.Annotations,
			})
		}
	case "event":
		for _, alert := range alertsFromAlertmanager {
			alertClass, ok := alert.Labels["class"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('class') noexist", "fingerprint", alert.Fingerprint)
				continue
			}
			alertSeverity, ok := alert.Labels["severity"]
			if !ok {
				rt.logger.Warn("ignore to statistic a alert, because of label('severity') noexist", "fingerprint", alert.Fingerprint)
				continue
			}

			switch strings.ToLower(strings.TrimSpace(alertClass)) {
			case "event":
				if _, ok := results.FiringStatistic.Event[alertSeverity]; !ok {
					results.FiringStatistic.Event[alertSeverity] = 1
				} else {
					results.FiringStatistic.Event[alertSeverity] += 1
				}
			default:
			}
			results.Alerts = append(results.Alerts, Alert{
				Fingerprint:  alert.Fingerprint,
				Status:       "firing",
				StartsAt:     alert.StartsAt,
				EndsAt:       alert.EndsAt,
				Generatorurl: alert.GeneratorURL,
				Lables:       alert.Labels,
				Annotations:  alert.Annotations,
			})
		}
	default:
		c.JSON(http.StatusBadRequest, response.Response{Results: results, Detail: "class 参数设定错误, 无该类别报警数据."})
		return
	}
	total := len(results.Alerts)
	var prev, next url.URL
	if pq.Paging {
		if (pq.Page-1)*pq.PageSize > len(results.Alerts) {
			c.JSON(http.StatusBadRequest, response.Response{Results: results, Detail: "页码超出实际最大值"})
			return
		}

		if pq.Page*pq.PageSize > len(results.Alerts) {
			results.Alerts = results.Alerts[(pq.Page-1)*pq.PageSize:]
		} else {
			results.Alerts = results.Alerts[(pq.Page-1)*pq.PageSize : pq.Page*pq.PageSize]
		}
		prev, next = response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	}

	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: results})
}

// HandlerGetAlertsHistory 从报警历史数据库中获取报警信息.
// @Summary 获取历史报警信息.
// @Description 从数据库中获取历史报警信息.
// @Tags 报警, 历史报警
// @Produce json
// @Param status query []string false "报警状态" collectionFormat(multi) default(firing,resolved)
// @Param start query string false "报警发生时间范围的开始时间, 需符合 RFC3339 时间格式, 默认为当前时间 - 1小时"
// @Param end query string false "报警发生时间范围的结束时间, 需符合 RFC3339 时间格式, 默认为当前时间"
// @Param labels query []string false "报警标签" collectionFormat(multi)
// @Param annotations query []string false "报警注释" collectionFormat(multi)
// @Param paging query bool false "启动分页, 当前仅支持开启模式" Enums(true) default(true)
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页报警数量, 最大仅支持100" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.Response{results=postgres.Alerts}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/alerts/history [get]
func (rt *Router) HandlerGetAlertsHistory(c *gin.Context) {
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.Paging = true
	pq.SetDefaults(1, 20, 100)

	start := c.DefaultQuery("start", time.Now().Add(-1*time.Hour).Format(time.RFC3339Nano))
	from, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "start 参数格式错误"})
		return
	}
	end := c.DefaultQuery("end", time.Now().Format(time.RFC3339Nano))
	to, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "end 参数格式错误"})
		return
	}
	status := c.QueryArray("status")
	for _, value := range status {
		v := strings.ToLower(strings.TrimSpace(value))
		if v != "firing" || v != "resolved" {
			c.JSON(http.StatusBadRequest, response.Response{Detail: "status 值字符串错误"})
			return
		}
	}
	labels := c.QueryArray("labels")
	condLabels := make(map[string][]string)
	for _, label := range labels {
		ss := strings.Split(label, "=")
		if len(ss) != 2 {
			c.JSON(http.StatusBadRequest, response.Response{Detail: fmt.Sprintf("labels 参数错误: %s", label)})
			return
		}
		if _, ok := condLabels[strings.TrimSpace(ss[0])]; !ok {
			condLabels[strings.TrimSpace(ss[0])] = make([]string, 0)
		}
		condLabels[strings.TrimSpace(ss[0])] = append(condLabels[strings.TrimSpace(ss[0])], ss[1])
	}

	annotations := c.QueryArray("annotations")
	condAnnotations := make(map[string][]string)
	for _, annotation := range annotations {
		ss := strings.Split(annotation, "=")
		if len(ss) != 2 {
			c.JSON(http.StatusBadRequest, response.Response{Detail: fmt.Sprintf("annotations 参数错误: %s", annotation)})
			return
		}
		if _, ok := condAnnotations[strings.TrimSpace(ss[0])]; !ok {
			condAnnotations[strings.TrimSpace(ss[0])] = make([]string, 0)
		}
		condAnnotations[strings.TrimSpace(ss[0])] = append(condAnnotations[strings.TrimSpace(ss[0])], ss[1])
	}

	rt.logger.Debug("query parameter", "start", start, "from", from)
	rt.logger.Debug("query parameter", "end", end, "to", to)
	rt.logger.Debug("query parameter", "status", status)
	rt.logger.Debug("query parameter", "labels", strings.Join(labels, ","), "condLabels", condLabels)
	rt.logger.Debug("query parameter", "annotations", strings.Join(annotations, ","), "condAnnotations", condAnnotations)

	alerts, total, err := rt.db.GetAlerts(c.Request.Context(), from, to, status, condLabels, condAnnotations, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "数据查询异常"})
		return
	}
	var prev, next url.URL
	prev, next = response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: alerts})
}
