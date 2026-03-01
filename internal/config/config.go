package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ConfigPath string
	Database   string
	Interval   time.Duration
	Drives     []string
	WebListen  string
	LogLevel   string
}

func Default() *Config {
	return &Config{
		ConfigPath: "",
		Database:   "diskmon.duckdb",
		Interval:   60 * time.Second,
		Drives:     []string{},
		WebListen:  "0.0.0.0:8976",
		LogLevel:   "INFO",
	}
}

func Load() (*Config, error) {
	return LoadFromPath("")
}

func LoadFromPath(path string) (*Config, error) {
	cfg := Default()

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("DISKMON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	v.SetDefault("database", cfg.Database)
	v.SetDefault("collector.interval", cfg.Interval)
	v.SetDefault("collector.drives", cfg.Drives)
	v.SetDefault("web.listen", cfg.WebListen)
	v.SetDefault("log.level", cfg.LogLevel)

	_ = v.BindEnv("database", "DISKMON_DATABASE")
	_ = v.BindEnv("collector.interval", "DISKMON_INTERVAL")
	_ = v.BindEnv("collector.drives", "DISKMON_DRIVES")
	_ = v.BindEnv("web.listen", "DISKMON_WEB_LISTEN")
	_ = v.BindEnv("log.level", "DISKMON_LOG_LEVEL")

	if path != "" {
		cfg.ConfigPath = path
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("diskmon")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg.Database = v.GetString("database")
	cfg.Interval = v.GetDuration("collector.interval")
	cfg.Drives = v.GetStringSlice("collector.drives")
	cfg.WebListen = v.GetString("web.listen")
	cfg.LogLevel = strings.ToUpper(v.GetString("log.level"))

	return cfg, nil
}

func ApplyFlagOverrides(cfg *Config, flags *pflag.FlagSet) {
	if flags == nil {
		return
	}

	if flags.Changed("config") {
		cfg.ConfigPath, _ = flags.GetString("config")
	}
	if flags.Changed("db") {
		cfg.Database, _ = flags.GetString("db")
	}
	if flags.Changed("interval") {
		cfg.Interval, _ = flags.GetDuration("interval")
	}
	if flags.Changed("web-listen") {
		cfg.WebListen, _ = flags.GetString("web-listen")
	}
	if flags.Changed("drives") {
		cfg.Drives, _ = flags.GetStringSlice("drives")
	}
	if flags.Changed("log-level") {
		cfg.LogLevel, _ = flags.GetString("log-level")
	}
}

func (c *Config) Validate() error {
	if c.Database == "" {
		return fmt.Errorf("database path is required")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be greater than zero")
	}
	if c.WebListen == "" {
		return fmt.Errorf("web listen address is required")
	}
	return nil
}
