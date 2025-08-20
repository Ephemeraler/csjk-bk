package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Client struct {
	db       *sql.DB
	jobTable string
	prefix   string
}

// slurm_job
// slurm_node

// Table resource_node
//
//	id serial PRIMARY KEY
//	name text NOT NULL
//	cluster text NOT NULL
//	socket  INTEGER NOT NULL
//	cores   INTEGER NOT NULL
//	mem     INTEGER NOT NULL
//	created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
//	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
func (c *Client) GetResourceOverview(ctx context.Context) {

}

// GetTotalJobCount 返回所有作业的数量.
func (c *Client) GetTotalJobCount(ctx context.Context) (int, error) {
	var count int
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM slurm_job").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GetJobCountByState 返回状态为 state 的作业数量.
func (c *Client) GetJobCountByState(ctx context.Context, state []int) (int, error) {
	var count int
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM slurm_job WHERE state IN (?)", state).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

type Sample struct {
	Value     float64
	Timestamp int64
}

// StatisticJobSubmit 作业提交趋势
func (c *Client) StatisticJobSubmit(ctx context.Context, start, end time.Time, step time.Duration, partition string) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		SELECT 
			FLOOR(time_submit / %d) * %d AS ts,
			COUNT(*) AS value
		FROM %s
		WHERE
			time_submit BETWEEN ? AND ?
			AND
			partition = ?
		GROUP BY ts
		ORDER BY ts;

	`, int(step.Seconds()), int(step.Seconds()), c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition)
	if err != nil {
		return samples, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ts, val int64
		if err := rows.Scan(&ts, &val); err != nil {
			return samples, fmt.Errorf("数据读取失败: %w", err)
		}
		samples = append(samples, Sample{Value: float64(val), Timestamp: ts})
	}

	return samples, nil
}

// StatisticJobCountByState 按照状态统计作业数量
func (c *Client) StatisticJobCountByState(ctx context.Context, state []int64, start, end time.Time, partition string) (map[int64]int64, error) {
	rlt := make(map[int64]int64)

	stmt := fmt.Sprintf(`
		SELECT COUNT(*), state FROM %s
		WHERE
			time_start BETWEEN ? AND ?
			AND
			partition = ?
			AND
			state IN (?)
		GROUP BY
			state
	`, c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, state)
	if err != nil {
		return rlt, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var count, state int64
		if err := rows.Scan(&count, &state); err != nil {
			return rlt, fmt.Errorf("无法读取数据: %w", err)
		}
		rlt[state] = count
	}
	return rlt, nil
}

// StatisticJobTimeByState 按照状态统计作业占用总时间
func (c *Client) StatisticJobTimeByState(ctx context.Context, state []int64, start, end time.Time, partition string) (map[int64]int64, error) {
	rlt := make(map[int64]int64)

	stmt := fmt.Sprintf(`
		SELECT
  			state,
  			SUM(time_end - time_start) AS value
		FROM
  			%s
		WHERE
    		time_start BETWEEN ? AND ?
    		AND partition = ?
    		AND state IN (?)
		GROUP BY
  			state
	`, c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, state)
	if err != nil {
		return rlt, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var state, value int64
		if err := rows.Scan(&state, &value); err != nil {
			return rlt, fmt.Errorf("无法读取数据: %w", err)
		}
		rlt[state] = value
	}
	return rlt, nil
}

// StatisticJobCountByUserTopK 按照用户统计作业数量 Top K
func (c *Client) StatisticJobCountByUserTopK(ctx context.Context, start, end time.Time, partition string, k int64) (map[string]int64, error) {
	rlt := make(map[string]int64)

	stmt := fmt.Sprintf(`
		SELECT
  			id_user,
  			COUNT(*) AS total
		FROM %s
		WHERE 
  			time_start BETWEEN ? AND ?
  			AND partition = ?
		GROUP BY id_user
		ORDER BY total DESC
		LIMIT ?
	`, c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, k)
	if err != nil {
		return rlt, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user string
		var count int64
		if err := rows.Scan(&user, &count); err != nil {
			return rlt, fmt.Errorf("无法读取数据: %w", err)
		}
		rlt[user] = count
	}
	return rlt, nil
}

// StatisticJobTimeByUserTopK 按照用户统计作业占用总时间 Top K
func (c *Client) StatisticJobTimeByUserTopK(ctx context.Context, start, end time.Time, partition string, k int64) (map[string]int64, error) {
	rlt := make(map[string]int64)

	stmt := fmt.Sprintf(`
		SELECT
			id_user,
			SUM(time_end - time_start) AS total
		FROM %s
		WHERE
			time_start BETWEEN ? AND ?
			AND partition = ?
		GROUP BY id_user
		ORDER BY total DESC
		LIMIT ?
	`, c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, k)
	if err != nil {
		return rlt, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user string
		var total int64
		if err := rows.Scan(&user, &total); err != nil {
			return rlt, fmt.Errorf("无法读取数据: %w", err)
		}
		rlt[user] = total
	}
	return rlt, nil
}

// StatisticJobCountByJobnameTopk 按照作业名称统计作业数量 Top K
func (c *Client) StatisticJobCountByJobnameTopk(ctx context.Context, start, end time.Time, partition string, k int64) (map[string]int64, error) {
	rlt := make(map[string]int64)

	stmt := fmt.Sprintf(`
		SELECT
			job_name,
			COUNT(*) AS total
		FROM %s
		WHERE
			time_start BETWEEN ? AND ?
			AND partition = ?
		GROUP BY job_name
		ORDER BY total DESC
		LIMIT ?
	`, c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, k)
	if err != nil {
		return rlt, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var jobName string
		var count int64
		if err := rows.Scan(&jobName, &count); err != nil {
			return rlt, fmt.Errorf("无法读取数据: %w", err)
		}
		rlt[jobName] = count
	}
	return rlt, nil
}

// StatisticJobTimeByJobnameTopk 按照作业名称统计作业占用总时间 Top K
func (c *Client) StatisticJobTimeByJobnameTopk(ctx context.Context, start, end time.Time, partition string, k int64) (map[string]int64, error) {
	rlt := make(map[string]int64)

	stmt := fmt.Sprintf(`
		SELECT
			job_name,
			SUM(time_end - time_start) AS total
		FROM %s
		WHERE
			time_start BETWEEN ? AND ?
			AND partition = ?
		GROUP BY job_name
		ORDER BY total DESC
		LIMIT ?
	`, c.jobTable)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, k)
	if err != nil {
		return rlt, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var jobName string
		var total int64
		if err := rows.Scan(&jobName, &total); err != nil {
			return rlt, fmt.Errorf("无法读取数据: %w", err)
		}
		rlt[jobName] = total
	}
	return rlt, nil
}

// StatisticJobCountByCPUS 按照 CPU 数量统计作业数目.
func (c *Client) StatisticJobCountByCPUS(ctx context.Context, start, end time.Time, partition string, min, max int64) (int64, error) {
	stmt := fmt.Sprintf(`
		SELECT
			COUNT(*) AS total
		FROM %s
		WHERE
			time_start BETWEEN ? AND ?
			AND partition = ?
			AND cpus_req BETWEEN ? AND ?
	`, c.jobTable)

	var total int64
	if err := c.db.QueryRowContext(ctx, stmt, start, end, partition, min, max).Scan(&total); err != nil {
		return 0, fmt.Errorf("无法查询数据: %w", err)
	}
	return total, nil
}

// StatisticJobTimeByCPUS 按照 CPU 数量统计作业运行时长.
func (c *Client) StatisticJobTimeByCPUS(ctx context.Context, start, end time.Time, partition string, min, max int64) (int64, error) {
	stmt := fmt.Sprintf(`
		SELECT
			SUM(time_end - time_start) AS total
		FROM %s
		WHERE
			time_start BETWEEN ? AND ?
			AND partition = ?
			AND cpus_req BETWEEN ? AND ?
	`, c.jobTable)

	var total int64
	if err := c.db.QueryRowContext(ctx, stmt, start, end, partition, min, max).Scan(&total); err != nil {
		return 0, fmt.Errorf("无法查询数据: %w", err)
	}
	return total, nil
}

// StatisticJobCountByRuntime 按照作业运行时长统计作业数目, min, max 单位为秒.
func (c *Client) StatisticJobCountByRuntime(ctx context.Context, start, end time.Time, partition string, min, max int64) (int64, error) {
	stmt := fmt.Sprintf(`
		SELECT
			COUNT(*) AS total
		FROM %s
		WHERE
			time_end > time_start
			AND
			(time_end - time_start) BETWEEN ? AND ?
			AND time_start BETWEEN ? AND ?
			AND partition = ?

	`, c.jobTable)

	var total int64
	if err := c.db.QueryRowContext(ctx, stmt, min, max, start, end, partition).Scan(&total); err != nil {
		return 0, fmt.Errorf("无法查询数据: %w", err)
	}
	return total, nil
}

// StatisticJobCountByRuntime 按照作业运行时长统计作业运行时长, min, max 单位为秒.
func (c *Client) StatisticJobTimeByRuntime(ctx context.Context, start, end time.Time, partition string, min, max int64) (int64, error) {
	stmt := fmt.Sprintf(`
		SELECT
			SUM(time_end - time_start) AS total
		FROM %s
		WHERE
			time_end > time_start
			AND
			(time_end - time_start) BETWEEN ? AND ?
			AND time_start BETWEEN ? AND ?
			AND partition = ?
	`, c.jobTable)

	var total int64
	if err := c.db.QueryRowContext(ctx, stmt, min, max, start, end, partition).Scan(&total); err != nil {
		return 0, fmt.Errorf("无法查询数据: %w", err)
	}
	return total, nil
}

// StatisticJobCountBySubmittime 根据提交时间统计作业数量.
func (c *Client) StatisticJobCountBySubmittime() {}

func (c *Client) StatisticSchedulerSystemAllocUtilization(ctx context.Context, start, end time.Time) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		SELECT
  			time_start AS time,
  			alloc_secs/count
		FROM
  			%s_usage_hour_table
		WHERE 
  			time_start BETWEEN ? AND ?
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticSchedulerSystemDownUtilization(ctx context.Context, start, end time.Time) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		SELECT
  			time_start AS time,
  			down_secs/count
		FROM
  			%s_usage_hour_table
		WHERE 
  			time_start BETWEEN ? AND ?
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticSchedulerSystemPDownUtilization(ctx context.Context, start, end time.Time) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		SELECT
  			time_start AS time,
  			pdown_secs/count
		FROM
  			%s_usage_hour_table
		WHERE 
  			time_start BETWEEN ? AND ?
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticSchedulerSystemIdleUtilization(ctx context.Context, start, end time.Time) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		SELECT
  			time_start AS time,
  			idle_secs/count
		FROM
  			%s_usage_hour_table
		WHERE 
  			time_start BETWEEN ? AND ?
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticSchedulerSystemResvUtilization(ctx context.Context, start, end time.Time) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		SELECT
  			time_start AS time,
  			resv_secs/count
		FROM
  			%s_usage_hour_table
		WHERE 
  			time_start BETWEEN ? AND ?
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticSchedulerSystemThroughput(ctx context.Context) {}

// StatisticSchedulerSystemAvgWaittime 作业平均等待时间趋势.
func (c *Client) StatisticSchedulerSystemAvgWaittime(ctx context.Context, start, end time.Time, partition string, min, max int64) ([]Sample, error) {
	samples := make([]Sample, 0)

	stmt := fmt.Sprintf(`
		WITH RECURSIVE hours AS (
  			SELECT FLOOR(%d / 3600) * 3600 AS time
			UNION ALL
			SELECT time + 3600
			FROM hours
			WHERE time + 3600 <= FLOOR(%d / 3600) * 3600
		)
		SELECT
			h.time,
			COALESCE(AVG(j.value), 0)
		FROM hours h
		LEFT JOIN (
		SELECT
			FLOOR(time_start / 3600) * 3600 AS time,
			(time_suspended + time_start - time_submit) AS value
		FROM %s_job_table
		WHERE 
			time_start BETWEEN ? AND ?
			AND partition = ?
			AND cpus_req BETWEEN ? AND ?
		) j
		ON h.time = j.time
		GROUP BY h.time
		ORDER BY h.time
	`, start.Unix(), end.Unix(), c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, min, max)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticSchedulerSystemTurnoverTime(ctx context.Context) {}

// 异常作业结束时间分时图

// 异常作业趋势
func (c *Client) StatisticJobCountByExpection(ctx context.Context, state []int64, start, end time.Time, partition string) ([]Sample, error) {
	samples := make([]Sample, 0)
	stmt := fmt.Sprintf(`
		WITH RECURSIVE hours AS (
  			SELECT FLOOR(%d / 3600) * 3600 AS time
			UNION ALL
			SELECT time + 3600
			FROM hours
			WHERE time + 3600 <= FLOOR(%d / 3600) * 3600
		),
		agg AS (
  			SELECT
    			FLOOR(time_start / 3600) * 3600 AS time,
				COUNT(*) AS value
			FROM %s_job_table
			WHERE
				time_start BETWEEN ? AND ?
				AND partition = ?
				AND state IN (?)
			GROUP BY time
		)
		SELECT
			h.time,
			COALESCE(a.value, 0) AS total
		FROM hours h
		LEFT JOIN agg a ON h.time = a.time
		ORDER BY h.time
	`, start.Unix(), end.Unix(), c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, state)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Timestamp, &sample.Value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (c *Client) StatisticUserCountByJobException(ctx context.Context, state []int64, start, end time.Time, partition string) (map[int64]map[int64]int64, error) {
	rlt := make(map[int64]map[int64]int64, 0)

	stmt := fmt.Sprintf(`
		SELECT
			id_user,
			state,
			COUNT(*) AS value
		FROM %s_job_table
		WHERE
		time_start BETWEEN ? AND ?
		AND partition = ?
		AND state IN (?)
		GROUP BY id_user, state
		ORDER BY id_user, state
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, state)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var idUser, state int64
		var value int64
		if err := rows.Scan(&idUser, &state, &value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		if rlt[idUser] == nil {
			rlt[idUser] = make(map[int64]int64)
		}
		rlt[idUser][state] = value
	}

	return rlt, nil
}

func (c *Client) StatisticJobCountByJobException(ctx context.Context, state []int64, start, end time.Time, partition string) (map[string]map[int64]int64, error) {
	rlt := make(map[string]map[int64]int64, 0)

	stmt := fmt.Sprintf(`
		SELECT
			job_name,
			state,
			COUNT(*) AS value
		FROM %s_job_table
		WHERE
		time_start BETWEEN ? AND ?
		AND partition = ?
		AND state IN (?)
		GROUP BY job_name, state
		ORDER BY job_name, state
	`, c.prefix)

	rows, err := c.db.QueryContext(ctx, stmt, start, end, partition, state)
	if err != nil {
		return nil, fmt.Errorf("无法查询数据: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var jobName string
		var value, state int64
		if err := rows.Scan(&jobName, &state, &value); err != nil {
			return nil, fmt.Errorf("无法读取数据: %w", err)
		}
		if rlt[jobName] == nil {
			rlt[jobName] = make(map[int64]int64)
		}
		rlt[jobName][state] = value
	}

	return rlt, nil
}
