package notification

import (
	"fmt"
	"strings"

	"diskmon/internal/health"
)

func FormatMessage(driveID string, result health.Result) (subject, body string) {
	subject = fmt.Sprintf("diskmon %s: %s", driveID, result.Status)

	var b strings.Builder
	fmt.Fprintf(&b, "Drive: %s\n", driveID)
	fmt.Fprintf(&b, "Status: %s\n", result.Status)
	fmt.Fprintf(&b, "Score: %d", result.Score)

	if len(result.Reasons) > 0 {
		b.WriteString("\nReasons:")
		for _, reason := range result.Reasons {
			b.WriteString("\n- ")
			b.WriteString(reason)
		}
	}

	return subject, b.String()
}
