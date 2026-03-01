package cli

import (
	"context"
	"log/slog"

	"diskmon/internal/smart"
)

func resolveDrives(ctx context.Context, configured []string, logger *slog.Logger) ([]string, error) {
	drives := normalizeDrives(configured)
	if len(drives) > 0 {
		return drives, nil
	}

	discovered, err := smart.DiscoverDevices(ctx)
	if err != nil {
		return nil, err
	}
	logger.Info("auto-discovered drives", "count", len(discovered), "drives", discovered)
	return discovered, nil
}
