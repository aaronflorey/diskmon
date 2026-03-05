package smart

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Runner interface {
	Run(ctx context.Context, device string) ([]byte, error)
	RunSelfTest(ctx context.Context, device string, testType string) ([]byte, error)
	RunSelfTestLog(ctx context.Context, device string) ([]byte, error)
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

func (r *ExecRunner) RunSelfTest(ctx context.Context, device string, testType string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "smartctl", "-t", testType, device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl %s test failed for %s: %w (%s)", testType, device, err, string(out))
	}
	return out, nil
}

func (r *ExecRunner) RunSelfTestLog(ctx context.Context, device string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "smartctl", "-l", "selftest", "-j", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl selftest log failed for %s: %w (%s)", device, err, string(out))
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

func (c *Collector) RunSelfTest(ctx context.Context, device string, testType string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(testType))
	if normalized != "short" && normalized != "long" {
		return "", fmt.Errorf("unsupported SMART test type %q", testType)
	}

	out, err := c.runner.RunSelfTest(ctx, device, normalized)
	if err != nil {
		return "", err
	}
	return summarizeSelfTestStartOutput(string(out), normalized), nil
}

func (c *Collector) ParseSelfTestWait(output string) time.Duration {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(strings.ToLower(line), "please wait") {
			continue
		}
		fields := strings.Fields(line)
		for i, token := range fields {
			if strings.EqualFold(token, "wait") && i+1 < len(fields) {
				if n, err := strconv.Atoi(strings.TrimSpace(fields[i+1])); err == nil && n >= 0 {
					return time.Duration(n) * time.Minute
				}
			}
		}
	}
	return 0
}

func (c *Collector) ReadSelfTestResult(ctx context.Context, device string, testType string) (string, string) {
	raw, err := c.runner.RunSelfTestLog(ctx, device)
	if err != nil {
		return "UNKNOWN", err.Error()
	}
	status, msg, ok := parseSelfTestResult(raw, testType)
	if !ok {
		return "UNKNOWN", strings.TrimSpace(string(raw))
	}
	return status, msg
}

func parseSelfTestResult(payload []byte, testType string) (string, string, bool) {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", "", false
	}
	root, ok := data["ata_smart_self_test_log"].(map[string]any)
	if !ok {
		return "", "", false
	}
	standard, ok := root["standard"].(map[string]any)
	if !ok {
		return "", "", false
	}
	table, ok := standard["table"].([]any)
	if !ok || len(table) == 0 {
		return "", "", false
	}
	for _, row := range table {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		typeStr := strings.ToLower(strAt(m, "type", "string"))
		if !testTypeMatches(typeStr, testType) {
			continue
		}
		statusStr := strAt(m, "status", "string")
		if statusStr == "" {
			statusStr = "unknown"
		}
		lower := strings.ToLower(statusStr)
		switch {
		case strings.Contains(lower, "without error"):
			return "PASSED", statusStr, true
		case strings.Contains(lower, "in progress"):
			return "IN_PROGRESS", statusStr, true
		case strings.Contains(lower, "aborted"):
			return "FAILED", statusStr, true
		case strings.Contains(lower, "completed"):
			return "FAILED", statusStr, true
		default:
			return "UNKNOWN", statusStr, true
		}
	}
	return "", "", false
}

func testTypeMatches(typeStr string, wanted string) bool {
	typeStr = strings.ToLower(typeStr)
	wanted = strings.ToLower(strings.TrimSpace(wanted))
	switch wanted {
	case "short":
		return strings.Contains(typeStr, "short")
	case "long":
		return strings.Contains(typeStr, "extended") || strings.Contains(typeStr, "long")
	default:
		return false
	}
}

func summarizeSelfTestStartOutput(raw string, testType string) string {
	lines := strings.Split(raw, "\n")
	waitLine := ""
	completeLine := ""
	startedLine := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "testing has begun"):
			startedLine = line
		case strings.Contains(lower, "please wait"):
			waitLine = line
		case strings.Contains(lower, "test will complete after"):
			completeLine = line
		}
	}

	parts := []string{fmt.Sprintf("SMART %s self-test started.", testType)}
	if startedLine != "" {
		parts = append(parts, startedLine)
	}
	if waitLine != "" {
		parts = append(parts, waitLine)
	}
	if completeLine != "" {
		parts = append(parts, completeLine)
	}

	msg := strings.Join(parts, " ")
	msg = strings.Join(strings.Fields(msg), " ")
	return strings.TrimSpace(msg)
}

func DiscoverDevices(ctx context.Context) ([]string, error) {
	return NewExecRunner().Discover(ctx)
}
