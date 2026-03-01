package health

type Rules struct {
	ReallocatedSectorsRed int64
	TemperatureWarnC      int
	PendingSectorsWarn    int64
	WearLevelWarnPct      int64
}

func DefaultRules() Rules {
	return Rules{
		ReallocatedSectorsRed: 16,
		TemperatureWarnC:      55,
		PendingSectorsWarn:    1,
		WearLevelWarnPct:      80,
	}
}
