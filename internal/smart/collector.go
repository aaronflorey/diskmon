package smart

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, device string) ([]byte, error)
}

type ExecRunner struct{}

func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

func (r *ExecRunner) Run(ctx context.Context, device string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "smartctl", "-a", "-j", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl failed for %s: %w (%s)", device, err, string(out))
	}
	return out, nil
}

func (r *ExecRunner) Discover(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "smartctl", "--scan-open")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl scan failed: %w (%s)", err, string(out))
	}

	seen := map[string]struct{}{}
	devices := make([]string, 0)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		device := fields[0]
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		if _, ok := seen[device]; ok {
			continue
		}
		seen[device] = struct{}{}
		devices = append(devices, device)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("smartctl scan returned no devices")
	}
	return devices, nil
}

type Collector struct {
	runner Runner
	log    *slog.Logger
}

func NewCollector(runner Runner, logger *slog.Logger) *Collector {
	return &Collector{runner: runner, log: logger}
}

func (c *Collector) CollectAll(ctx context.Context, devices []string) ([]CollectResult, error) {
	results := make([]CollectResult, 0, len(devices))
	for _, d := range devices {
		res, err := c.CollectOne(ctx, d)
		if err != nil {
			c.log.Warn("smart collection failed", "device", d, "error", err)
			continue
		}
		results = append(results, res)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no successful collection results")
	}
	return results, nil
}

func (c *Collector) CollectOne(ctx context.Context, device string) (CollectResult, error) {
	raw, err := c.runner.Run(ctx, device)
	if err != nil {
		return CollectResult{}, err
	}
	info, sample, err := ParseSmartJSON(device, raw)
	if err != nil {
		return CollectResult{}, err
	}
	if info.Device == "" {
		info.Device = device
	}
	return CollectResult{Info: info, Sample: sample}, nil
}

func DiscoverDevices(ctx context.Context) ([]string, error) {
	return NewExecRunner().Discover(ctx)
}
