package smart

import (
	"fmt"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

func ParseSmartJSON(device string, payload []byte) (DriveInfo, SmartSample, error) {
	root := gjson.ParseBytes(payload)
	if root.Type != gjson.JSON {
		return DriveInfo{}, SmartSample{}, fmt.Errorf("decode smartctl json: invalid json payload")
	}

	attrs := parseAttributes(root)
	attrMap := buildAttrMap(attrs)

	info := DriveInfo{
		Device: device,
		Model:  firstNonEmpty(resultString(root.Get("model_name")), resultString(root.Get("model_family")), resultString(root.Get("model_number"))),
		Serial: resultString(root.Get("serial_number")),
		WWN:    resultString(root.Get("wwn.naa")),
	}

	// Temperature: prefer top-level temperature.current, then NVMe log
	temp := firstInt64(root,
		"temperature.current",
		"nvme_smart_health_information_log.temperature",
	)
	if temp != nil && *temp > 200 {
		c := int64(float64(*temp) - 273.15)
		temp = &c
	}

	// Critical counters: use attribute ID lookup only (never array index)
	realloc := attrRawValueByID(attrMap, 5)
	pending := attrRawValueByID(attrMap, 197)
	uncorr := attrRawValueByID(attrMap, 198)
	udmaCRC := attrRawValueByID(attrMap, 199)
	reportedUncorr := attrRawValueByID(attrMap, 187)
	cmdTimeout := attrRawValueByID(attrMap, 188)

	// NVMe fallback for uncorrectable
	if uncorr == nil {
		uncorr = firstInt64(root, "nvme_smart_health_information_log.media_errors")
	}

	// Wear level (NVMe)
	wear := firstInt64(root,
		"nvme_smart_health_information_log.percentage_used",
	)

	criticalWarning := false
	if cw := firstInt64(root, "nvme_smart_health_information_log.critical_warning"); cw != nil && *cw > 0 {
		criticalWarning = true
	}

	failingNow := false
	if passed := root.Get("smart_status.passed"); passed.Exists() &&
		(passed.Type == gjson.True || passed.Type == gjson.False) {
		failingNow = !passed.Bool()
	}

	sample := SmartSample{
		CollectedAt:          time.Now().UTC(),
		Temperature:          intPtrFrom64(temp),
		PowerOnHours:         firstInt64(root, "power_on_time.hours", "power_on_hours"),
		ReallocatedSectors:   realloc,
		PendingSectors:       pending,
		UncorrectableSectors: uncorr,
		UDMACRCErrors:        udmaCRC,
		ReportedUncorrect:    reportedUncorr,
		CommandTimeout:       cmdTimeout,
		WearLevel:            wear,
		FailingNow:           failingNow,
		CriticalWarning:      criticalWarning,
		RawJSON:              string(payload),
		Attributes:           attrs,
	}

	return info, sample, nil
}

func parseAttributes(root gjson.Result) []SmartAttribute {
	table := root.Get("ata_smart_attributes.table")
	if !table.Exists() || !table.IsArray() {
		return nil
	}
	entries := table.Array()
	attrs := make([]SmartAttribute, 0, len(entries))
	for _, item := range entries {
		rawValue := item.Get("raw.value").Uint()
		rawStr := strings.TrimSpace(resultString(item.Get("raw.string")))
		if rawStr == "" {
			rawStr = fmt.Sprintf("%d", rawValue)
		}

		attr := SmartAttribute{
			AttributeID: int(item.Get("id").Int()),
			Name:        resultString(item.Get("name")),
			Value:       int(item.Get("value").Int()),
			Worst:       int(item.Get("worst").Int()),
			Threshold:   int(item.Get("thresh").Int()),
			Raw:         rawStr,
			RawValue:    rawValue,
			WhenFailed:  resultString(item.Get("when_failed")),
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// buildAttrMap indexes attributes by their ID for O(1) lookup.
func buildAttrMap(attrs []SmartAttribute) map[int]SmartAttribute {
	m := make(map[int]SmartAttribute, len(attrs))
	for _, a := range attrs {
		m[a.AttributeID] = a
	}
	return m
}

// attrRawValueByID looks up an attribute by ID and returns its raw value.
func attrRawValueByID(m map[int]SmartAttribute, id int) *int64 {
	attr, ok := m[id]
	if !ok {
		return nil
	}
	v := int64(attr.RawValue)
	return &v
}

func firstInt64(root gjson.Result, paths ...string) *int64 {
	for _, path := range paths {
		res := root.Get(path)
		if !res.Exists() || res.Type != gjson.Number {
			continue
		}
		v := res.Int()
		return &v
	}
	return nil
}

func resultString(result gjson.Result) string {
	if result.Type != gjson.String {
		return ""
	}
	return result.String()
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
