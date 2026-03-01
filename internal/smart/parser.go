package smart

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type rawSmartData map[string]any

func ParseSmartJSON(device string, payload []byte) (DriveInfo, SmartSample, error) {
	var data rawSmartData
	if err := json.Unmarshal(payload, &data); err != nil {
		return DriveInfo{}, SmartSample{}, fmt.Errorf("decode smartctl json: %w", err)
	}

	attrs := parseAttributes(data)

	info := DriveInfo{
		Device: device,
		Model:  firstNonEmpty(strAt(data, "model_name"), strAt(data, "model_family"), strAt(data, "model_number")),
		Serial: strAt(data, "serial_number"),
		WWN:    strAt(data, "wwn", "naa"),
	}

	temp := firstInt64(data,
		[]string{"temperature", "current"},
		[]string{"nvme_smart_health_information_log", "temperature"},
	)
	if temp != nil && *temp > 200 {
		c := int64(float64(*temp) - 273.15)
		temp = &c
	}

	pending := firstInt64(data, []string{"ata_smart_attributes", "table", "197", "raw", "value"})
	realloc := firstInt64(data, []string{"ata_smart_attributes", "table", "5", "raw", "value"})
	uncorr := firstInt64(data,
		[]string{"ata_smart_attributes", "table", "198", "raw", "value"},
		[]string{"nvme_smart_health_information_log", "media_errors"},
	)
	wear := firstInt64(data,
		[]string{"nvme_smart_health_information_log", "percentage_used"},
	)

	if realloc == nil {
		realloc = attrRawByID(attrs, 5)
	}
	if pending == nil {
		pending = attrRawByID(attrs, 197)
	}
	if uncorr == nil {
		uncorr = attrRawByID(attrs, 198)
	}

	criticalWarning := false
	if cw := firstInt64(data, []string{"nvme_smart_health_information_log", "critical_warning"}); cw != nil && *cw > 0 {
		criticalWarning = true
	}

	failingNow := false
	if v := nested(data, "smart_status", "passed"); v != nil {
		if passed, ok := v.(bool); ok {
			failingNow = !passed
		}
	}

	sample := SmartSample{
		CollectedAt:          time.Now().UTC(),
		Temperature:          intPtrFrom64(temp),
		PowerOnHours:         firstInt64(data, []string{"power_on_time", "hours"}, []string{"power_on_hours"}),
		ReallocatedSectors:   realloc,
		PendingSectors:       pending,
		UncorrectableSectors: uncorr,
		WearLevel:            wear,
		FailingNow:           failingNow,
		CriticalWarning:      criticalWarning,
		RawJSON:              string(payload),
		Attributes:           attrs,
	}

	return info, sample, nil
}

func parseAttributes(data rawSmartData) []SmartAttribute {
	table, ok := nested(data, "ata_smart_attributes", "table").([]any)
	if !ok {
		return nil
	}
	attrs := make([]SmartAttribute, 0, len(table))
	for _, item := range table {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		attr := SmartAttribute{
			AttributeID: int(numberAt(m, "id")),
			Name:        strAt(m, "name"),
			Value:       int(numberAt(m, "value")),
			Worst:       int(numberAt(m, "worst")),
			Threshold:   int(numberAt(m, "thresh")),
			Raw:         fmt.Sprintf("%v", nested(m, "raw", "string")),
		}
		if strings.TrimSpace(attr.Raw) == "<nil>" || strings.TrimSpace(attr.Raw) == "" {
			if v, ok := getNumber(m, "raw", "value"); ok {
				attr.Raw = fmt.Sprintf("%.0f", v)
			}
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

func attrRawByID(attrs []SmartAttribute, id int) *int64 {
	for _, attr := range attrs {
		if attr.AttributeID == id {
			if v, err := strconv.ParseInt(strings.Fields(attr.Raw)[0], 10, 64); err == nil {
				return &v
			}
			vv := int64(attr.Value)
			return &vv
		}
	}
	return nil
}

func firstInt64(data rawSmartData, paths ...[]string) *int64 {
	for _, p := range paths {
		if v, ok := getNumber(data, p...); ok {
			vv := int64(v)
			return &vv
		}
	}
	return nil
}

func getNumber(data map[string]any, keys ...string) (float64, bool) {
	v := nested(data, keys...)
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func numberAt(data map[string]any, keys ...string) float64 {
	v, _ := getNumber(data, keys...)
	return v
}

func strAt(data map[string]any, keys ...string) string {
	v := nested(data, keys...)
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func nested(data map[string]any, keys ...string) any {
	var cur any = data
	for _, k := range keys {
		switch n := cur.(type) {
		case map[string]any:
			cur = n[k]
		case []any:
			idx := -1
			_, _ = fmt.Sscanf(k, "%d", &idx)
			if idx < 0 || idx >= len(n) {
				return nil
			}
			cur = n[idx]
		default:
			return nil
		}
	}
	return cur
}

func intPtrFrom64(v *int64) *int {
	if v == nil {
		return nil
	}
	o := int(*v)
	return &o
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
