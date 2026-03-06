package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"diskmon/internal/api"
	"diskmon/internal/config"
	"diskmon/internal/health"
	"diskmon/internal/smart"
	"diskmon/internal/storage"

	"github.com/robfig/cron/v3"
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

			deletedRuns, err := db.DeleteIncompleteSmartTestRuns(ctx)
			if err != nil {
				logger.Error("failed cleaning incomplete SMART test runs", "error", err)
			} else if deletedRuns > 0 {
				logger.Info("cleaned incomplete SMART test runs", "deleted", deletedRuns)
			}

			collector := smart.NewCollector(smart.NewExecRunner(), logger)
			evaluator := health.NewEvaluator(health.DefaultRules())
			events := api.NewEventBroker()

			apiServer := api.NewServer(cfg.WebListen, logger, db, events)
			errCh := make(chan error, 1)
			go func() {
				if err := apiServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errCh <- err
				}
			}()

			cronScheduler, err := configureSmartTestCron(ctx, cfg, drives, collector, db, logger, events)
			if err != nil {
				return err
			}
			if cronScheduler != nil {
				cronScheduler.Start()
				defer cronScheduler.Stop()
			}

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
						continue
					}
					events.Publish("sample.inserted", res.Info.Device)
				}
			}

			runCollection()
			for {
				select {
				case <-ctx.Done():
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := apiServer.Shutdown(shutdownCtx); err != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							logger.Warn("graceful shutdown timed out; forcing close")
							_ = apiServer.Close()
							return nil
						}
						return err
					}
					return nil
				case err := <-errCh:
					return err
				case <-ticker.C:
					runCollection()
				}
			}
		},
	}
}

func configureSmartTestCron(
	ctx context.Context,
	cfg *config.Config,
	drives []string,
	collector *smart.Collector,
	db *storage.DuckDB,
	logger *slog.Logger,
	events *api.EventBroker,
) (*cron.Cron, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	scheduler := cron.New(cron.WithParser(parser))
	enabled := false
	inFlight := make(map[string]bool)
	var inFlightMu sync.Mutex

	addJob := func(testType string, expr *string) error {
		if expr == nil {
			return nil
		}
		testType = strings.ToLower(strings.TrimSpace(testType))
		spec := strings.TrimSpace(*expr)
		if _, err := scheduler.AddFunc(spec, func() {
			scheduledAt := time.Now().UTC()
			for _, device := range drives {
				select {
				case <-ctx.Done():
					return
				default:
				}
				testKey := device + ":" + testType
				inFlightMu.Lock()
				if inFlight[testKey] {
					inFlightMu.Unlock()
					logger.Warn("skipping scheduled SMART test; previous run still in progress", "device", device, "type", testType)
					continue
				}
				inFlight[testKey] = true
				inFlightMu.Unlock()

				func() {
					defer func() {
						inFlightMu.Lock()
						delete(inFlight, testKey)
						inFlightMu.Unlock()
					}()

					baselineStatus, baselineMsg := collector.ReadSelfTestResult(ctx, device, testType)
					if baselineStatus == "IN_PROGRESS" {
						logger.Warn("skipping scheduled SMART test; device reports test already in progress", "device", device, "type", testType)
						return
					}

					startedAt := time.Now().UTC()
					output, runErr := collector.RunSelfTest(ctx, device, testType)
					finishedAt := startedAt

					status := "STARTED"
					message := output
					if runErr != nil {
						status = "FAILED"
						message = runErr.Error()
						finishedAt = time.Now().UTC()
						logger.Error("scheduled SMART test failed", "device", device, "type", testType, "error", runErr)
					} else {
						logger.Info("scheduled SMART test triggered", "device", device, "type", testType)
					}

					if _, err := db.InsertSmartTestRun(ctx, smart.DriveInfo{Device: device}, storage.SmartTestRun{
						TestType:    testType,
						ScheduledAt: scheduledAt,
						StartedAt:   startedAt,
						FinishedAt:  finishedAt,
						Status:      status,
						Message:     message,
					}); err != nil {
						logger.Error("failed storing SMART test run", "device", device, "type", testType, "error", err)
					} else {
						events.Publish("test.updated", device)
					}

					if runErr != nil {
						return
					}

					waitFor := collector.ParseSelfTestWait(output)
					if waitFor > 0 {
						waitFor += 10 * time.Second
						select {
						case <-ctx.Done():
							return
						case <-time.After(waitFor):
						}
					}

					finalStatus := "UNKNOWN"
					finalMsg := "self-test result unavailable"
					for i := 0; i < 12; i++ {
						finalStatus, finalMsg = collector.ReadSelfTestResult(ctx, device, testType)
						if finalStatus == "IN_PROGRESS" {
							// Test is still running, keep polling.
						} else if finalStatus == baselineStatus && finalMsg == baselineMsg {
							// Result unchanged from before we started the test;
							// likely a stale entry from a previous run.
							logger.Debug("self-test result unchanged from baseline, still waiting", "device", device, "type", testType, "status", finalStatus)
						} else {
							break
						}
						select {
						case <-ctx.Done():
							return
						case <-time.After(20 * time.Second):
						}
					}

					finalFinishedAt := time.Now().UTC()
					if _, err := db.InsertSmartTestRun(ctx, smart.DriveInfo{Device: device}, storage.SmartTestRun{
						TestType:    testType,
						ScheduledAt: scheduledAt,
						StartedAt:   startedAt,
						FinishedAt:  finalFinishedAt,
						Status:      finalStatus,
						Message:     finalMsg,
					}); err != nil {
						logger.Error("failed storing SMART final test result", "device", device, "type", testType, "error", err)
						return
					} else {
						events.Publish("test.updated", device)
					}
				}()
			}
		}); err != nil {
			return fmt.Errorf("configure collector.tests.%s: %w", testType, err)
		}
		logger.Info("scheduled SMART test enabled", "type", testType, "cron", spec)
		enabled = true
		return nil
	}

	if err := addJob("short", cfg.Tests.Short); err != nil {
		return nil, err
	}
	if err := addJob("long", cfg.Tests.Long); err != nil {
		return nil, err
	}
	if !enabled {
		return nil, nil
	}
	return scheduler, nil
}
