package cli

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"diskmon/internal/api"
	"diskmon/internal/config"
	"diskmon/internal/health"
	"diskmon/internal/smart"
	"diskmon/internal/storage"

	"github.com/spf13/cobra"
)

func newDaemonCmd(cfg *config.Config, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "daemon",
		Short: "Run diskmon daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			drives, err := resolveDrives(ctx, cfg.Drives, logger)
			if err != nil {
				return err
			}

			db, err := storage.OpenDuckDB(cfg.Database)
			if err != nil {
				return err
			}
			defer db.Close()

			collector := smart.NewCollector(smart.NewExecRunner(), logger)
			evaluator := health.NewEvaluator(health.DefaultRules())

			apiServer := api.NewServer(cfg.WebListen, logger, db)
			errCh := make(chan error, 1)
			go func() {
				if err := apiServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errCh <- err
				}
			}()

			ticker := time.NewTicker(cfg.Interval)
			defer ticker.Stop()

			runCollection := func() {
				results, err := collector.CollectAll(ctx, drives)
				if err != nil {
					logger.Error("collection failed", "error", err)
					return
				}
				for _, res := range results {
					healthResult := evaluator.Evaluate(res.Sample)
					if _, err := db.InsertSample(ctx, res.Info, res.Sample, healthResult); err != nil {
						logger.Error("failed storing sample", "device", res.Info.Device, "error", err)
					}
				}
			}

			runCollection()
			for {
				select {
				case <-ctx.Done():
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					return apiServer.Shutdown(shutdownCtx)
				case err := <-errCh:
					return err
				case <-ticker.C:
					runCollection()
				}
			}
		},
	}
}
