package health

type Rules struct {
	TemperatureRedC  int
	TemperatureWarnC int
	WearLevelWarnPct int64
}

func DefaultRules() Rules {
	return Rules{
		TemperatureRedC:  55,
		TemperatureWarnC: 50,
		WearLevelWarnPct: 80,
	}
}
