package config

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func TestOptionalString(t *testing.T) {
	if got := optionalString("   "); got != nil {
		t.Fatalf("expected nil for whitespace, got %v", *got)
	}
	got := optionalString("  */5 * * * *  ")
	if got == nil || *got != "*/5 * * * *" {
		t.Fatalf("expected trimmed cron string, got %v", got)
	}
}

func TestValidate(t *testing.T) {
	validShort := "*/5 * * * *"
	validLong := "0 2 * * *"

	cfg := Default()
	cfg.Tests.Short = &validShort
	cfg.Tests.Long = &validLong
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	cfg2 := Default()
	cfg2.Database = ""
	if err := cfg2.Validate(); err == nil {
		t.Fatal("expected error for empty database")
	}

	cfg3 := Default()
	cfg3.Interval = 0
	if err := cfg3.Validate(); err == nil {
		t.Fatal("expected error for non-positive interval")
	}

	cfg4 := Default()
	cfg4.WebListen = ""
	if err := cfg4.Validate(); err == nil {
		t.Fatal("expected error for empty web listen")
	}

	cfg5 := Default()
	bad := "not-a-cron"
	cfg5.Tests.Short = &bad
	if err := cfg5.Validate(); err == nil {
		t.Fatal("expected error for invalid short cron")
	}
}

func TestApplyFlagOverrides(t *testing.T) {
	cfg := Default()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("config", cfg.ConfigPath, "")
	flags.String("db", cfg.Database, "")
	flags.Duration("interval", cfg.Interval, "")
	flags.String("web-listen", cfg.WebListen, "")
	flags.StringSlice("drives", cfg.Drives, "")
	flags.String("log-level", cfg.LogLevel, "")

	if err := flags.Set("db", "/tmp/test.duckdb"); err != nil {
		t.Fatalf("set db flag: %v", err)
	}
	if err := flags.Set("interval", "30s"); err != nil {
		t.Fatalf("set interval flag: %v", err)
	}
	if err := flags.Set("drives", "/dev/sda,/dev/sdb"); err != nil {
		t.Fatalf("set drives flag: %v", err)
	}

	ApplyFlagOverrides(cfg, flags)

	if cfg.Database != "/tmp/test.duckdb" {
		t.Fatalf("database override not applied: %q", cfg.Database)
	}
	if cfg.Interval != 30*time.Second {
		t.Fatalf("interval override not applied: %v", cfg.Interval)
	}
	if len(cfg.Drives) != 2 {
		t.Fatalf("drives override not applied: %v", cfg.Drives)
	}
}

