//go:build !cgo

package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"diskmon/internal/health"
	"diskmon/internal/smart"
)

var ErrCGODisabled = errors.New("duckdb storage requires cgo; rebuild with CGO_ENABLED=1 and a linux cross-compiler")

type DuckDB struct{}

type DriveSummary struct {
	ID          int64      `json:"id"`
	Device      string     `json:"device"`
	Model       string     `json:"model"`
	Serial      string     `json:"serial"`
	Health      string     `json:"health"`
	Temperature *int       `json:"temperature"`
	PowerOnHrs  *int64     `json:"power_on_hours"`
	LastSeen    *time.Time `json:"last_seen"`
}

type DriveDetail struct {
	ID            int64      `json:"id"`
	Device        string     `json:"device"`
	Model         string     `json:"model"`
	Serial        string     `json:"serial"`
	WWN           string     `json:"wwn"`
	Health        string     `json:"health"`
	HealthScore   int        `json:"health_score"`
	HealthReasons string     `json:"health_reasons"`
	Temperature   *int       `json:"temperature"`
	PowerOnHours  *int64     `json:"power_on_hours"`
	Reallocated   *int64     `json:"reallocated_sectors"`
	Pending       *int64     `json:"pending_sectors"`
	Uncorrectable *int64     `json:"uncorrectable_sectors"`
	WearLevel     *int64     `json:"wear_level"`
	CollectedAt   *time.Time `json:"collected_at"`
	FirstSeen     *time.Time `json:"first_seen"`
	LastSeen      *time.Time `json:"last_seen"`
}

type HistoryPoint struct {
	CollectedAt          time.Time `json:"collected_at"`
	Temperature          *int      `json:"temperature"`
	PowerOnHours         *int64    `json:"power_on_hours"`
	ReallocatedSectors   *int64    `json:"reallocated_sectors"`
	PendingSectors       *int64    `json:"pending_sectors"`
	UncorrectableSectors *int64    `json:"uncorrectable_sectors"`
	WearLevel            *int64    `json:"wear_level"`
}

type AttributePoint struct {
	AttributeID int    `json:"attribute_id"`
	Name        string `json:"name"`
	Value       int    `json:"value"`
	Worst       int    `json:"worst"`
	Threshold   int    `json:"threshold"`
	Raw         string `json:"raw"`
	Status      string `json:"status"`
}

func OpenDuckDB(path string) (*DuckDB, error) {
	return nil, ErrCGODisabled
}

func (d *DuckDB) Close() error {
	return ErrCGODisabled
}

func (d *DuckDB) Conn(ctx context.Context) (*sql.Conn, error) {
	return nil, ErrCGODisabled
}

func (d *DuckDB) InsertSample(ctx context.Context, info smart.DriveInfo, sample smart.SmartSample, result health.Result) (int64, error) {
	return 0, ErrCGODisabled
}

func (d *DuckDB) ListDrives(ctx context.Context) ([]DriveSummary, error) {
	return nil, ErrCGODisabled
}

func (d *DuckDB) GetDrive(ctx context.Context, id int64) (*DriveDetail, error) {
	return nil, ErrCGODisabled
}

func (d *DuckDB) DriveHistory(ctx context.Context, id int64, limit int) ([]HistoryPoint, error) {
	return nil, ErrCGODisabled
}

func (d *DuckDB) DriveAttributes(ctx context.Context, id int64) ([]AttributePoint, error) {
	return nil, ErrCGODisabled
}
