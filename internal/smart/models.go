package smart

import "time"

type DriveInfo struct {
	Device string
	Model  string
	Serial string
	WWN    string
}

type SmartAttribute struct {
	AttributeID int
	Name        string
	Value       int
	Worst       int
	Threshold   int
	Raw         string
	RawValue    uint64
	WhenFailed  string // empty means not failed; "past" or "now" means failure
}

type SmartSample struct {
	CollectedAt          time.Time
	Temperature          *int
	PowerOnHours         *int64
	ReallocatedSectors   *int64
	PendingSectors       *int64
	UncorrectableSectors *int64
	UDMACRCErrors        *int64
	ReportedUncorrect    *int64
	CommandTimeout       *int64
	WearLevel            *int64
	FailingNow           bool
	CriticalWarning      bool
	RawJSON              string
	Attributes           []SmartAttribute
}

type CollectResult struct {
	Info   DriveInfo
	Sample SmartSample
}
