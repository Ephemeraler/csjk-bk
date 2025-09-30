package exec

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type ExecCommandFunc func(ctx context.Context, name string, args ...string) *exec.Cmd

type Client struct {
	execCommand ExecCommandFunc
	logger      *slog.Logger
}

func (c *Client) Set(exec ExecCommandFunc, logger *slog.Logger) *Client {
	c.execCommand = exec
	c.logger = logger
	return c
}

func (c *Client) SetUpperThreshsOfOutbandSensor(ctx context.Context, rmu, bmu, id, unr, ucr, unc string) error {
	// TODO: 账号和密码会经常改吗？要不写到配置文件中.
	cmd := c.execCommand(ctx, "smu_transfer_cmd", rmu, "ipmitool", "-I", "lanp", "-H", bmu, "-U", "admin", "-P", "admin", "sensor", "thresh", id, "upper", unc, ucr, unr)
	output, err := cmd.CombinedOutput()
	c.logger.Debug(cmd.String())
	if err != nil {
		c.logger.Error("unable to execute command", "cmd", cmd.String(), "output", output, "err", err)
		return fmt.Errorf(cmd.String())
	}
	return nil
}

func (c *Client) SetLowerThreshsOfOutbandSensor(ctx context.Context, rmu, bmu, id, lnr, lcr, lnc string) error {
	// TODO: 账号和密码会经常改吗？要不写到配置文件中.
	cmd := c.execCommand(ctx, "smu_transfer_cmd", rmu, "ipmitool", "-I", "lanp", "-H", bmu, "-U", "admin", "-P", "admin", "sensor", "thresh", id, "lower", lnr, lcr, lnc)
	output, err := cmd.CombinedOutput()
	c.logger.Debug(cmd.String())
	if err != nil {
		c.logger.Error("unable to execute command", "cmd", cmd.String(), "output", output, "err", err)
		return fmt.Errorf(cmd.String())
	}
	return nil
}

func (c *Client) SetThreshsOfOutbandSensor(ctx context.Context, rmu, bmu, id, which, value string) error {
	// TODO: 账号和密码会经常改吗？要不写到配置文件中.
	cmd := c.execCommand(ctx, "smu_transfer_cmd", rmu, "ipmitool", "-I", "lanp", "-H", bmu, "-U", "admin", "-P", "admin", "sensor", "thresh", id, which, value)
	output, err := cmd.CombinedOutput()
	c.logger.Debug(cmd.String())
	if err != nil {
		c.logger.Error("unable to execute command", "cmd", cmd.String(), "output", output, "err", err)
		return fmt.Errorf(cmd.String())
	}
	return nil
}

func (c *Client) SetInhibitOfOutbandSensor(ctx context.Context, rmu, board, sensor, rule string) error {
	cmd := c.execCommand(ctx, "smu_transfer_cmd", rmu, "set_sensor_filter.sh", board, sensor, rule)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logger.Error("unable to execute command", "cmd", cmd.String(), "output", output, "err", err)
		return fmt.Errorf(cmd.String())
	}

	if strings.Contains(string(output), "set success") {
		return nil
	}

	return fmt.Errorf(string(output))
}

func (c *Client) GetThresholdOfOutbandSensor(ctx context.Context, rmu, bmu string) (Thresholds, error) {
	cmd := c.execCommand(ctx, "smu_transfer_cmd", rmu, "ipmitool", "-I", "lanp", "-H", bmu, "-U", "admin", "-P", "admin", "sensor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logger.Error("unable to execute command", "cmd", cmd.String(), "output", output, "err", err)
		return nil, fmt.Errorf(cmd.String())
	}

	thd, err := parseThresholdOfOutbandSensor(output)
	if err != nil {
		c.logger.Error(fmt.Sprintf("unable to parse output for %s", cmd.String()), "output", output)
	}

	return thd, fmt.Errorf(string(output))
}

type Thresholds []Threshold
type Threshold struct {
	Name    string
	Current string
	Unit    string
	State   string
	Value1  string
	Value2  string
	Value3  string
	Value4  string
	Value5  string
	Value6  string
}

func parseThresholdOfOutbandSensor(content []byte) (Thresholds, error) {
	ths := make(Thresholds, 0)
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		linetext := scanner.Text()
		fields := strings.Split(linetext, "|")
		if len(fields) != 10 {
			return nil, fmt.Errorf("传感器阈值查询命令输出格式与预期不符合, '|' 没有划分10个输出")
		}
		ths = append(ths, Threshold{
			Name:    fields[0],
			Current: fields[1],
			Unit:    fields[2],
			State:   fields[3],
			Value1:  fields[4],
			Value2:  fields[5],
			Value3:  fields[6],
			Value4:  fields[7],
			Value5:  fields[8],
			Value6:  fields[9],
		})
	}

	return ths, nil
}
