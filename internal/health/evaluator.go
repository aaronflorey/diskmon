package health

import "diskmon/internal/smart"

type Status string

const (
	StatusGreen   Status = "GREEN"
	StatusYellow  Status = "YELLOW"
	StatusRed     Status = "RED"
	StatusUnknown Status = "UNKNOWN"
)

type Result struct {
	Status  Status
	Score   int
	Reasons []string
}

type Evaluator struct {
	rules Rules
}

func NewEvaluator(rules Rules) *Evaluator {
	return &Evaluator{rules: rules}
}

func (e *Evaluator) Evaluate(s smart.SmartSample) Result {
	reasons := make([]string, 0, 4)

	if s.FailingNow {
		reasons = append(reasons, "drive reports failing_now")
	}
	if s.ReallocatedSectors != nil && *s.ReallocatedSectors > e.rules.ReallocatedSectorsRed {
		reasons = append(reasons, "reallocated sectors exceeded threshold")
	}
	if s.UncorrectableSectors != nil && *s.UncorrectableSectors > 0 {
		reasons = append(reasons, "uncorrectable sectors detected")
	}
	if s.CriticalWarning {
		reasons = append(reasons, "nvme critical warning present")
	}
	if len(reasons) > 0 {
		return Result{Status: StatusRed, Score: 20, Reasons: reasons}
	}

	if s.Temperature != nil && *s.Temperature > e.rules.TemperatureWarnC {
		reasons = append(reasons, "temperature above warning threshold")
	}
	if s.PendingSectors != nil && *s.PendingSectors >= e.rules.PendingSectorsWarn {
		reasons = append(reasons, "pending sectors detected")
	}
	if s.WearLevel != nil && *s.WearLevel >= e.rules.WearLevelWarnPct {
		reasons = append(reasons, "wear level degraded")
	}
	if len(reasons) > 0 {
		return Result{Status: StatusYellow, Score: 65, Reasons: reasons}
	}

	if s.Temperature == nil && s.PowerOnHours == nil && len(s.Attributes) == 0 {
		return Result{Status: StatusUnknown, Score: 0, Reasons: []string{"insufficient SMART data"}}
	}

	return Result{Status: StatusGreen, Score: 95, Reasons: []string{"no warnings triggered"}}
}
