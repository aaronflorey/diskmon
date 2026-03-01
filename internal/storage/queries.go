//go:build cgo

package storage

import (
	"context"
	"database/sql"
	"time"
)

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

func (d *DuckDB) ListDrives(ctx context.Context) ([]DriveSummary, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT d.id, d.device, d.model, d.serial,
		       COALESCE(h.status, 'UNKNOWN') AS health,
		       s.temperature, s.power_on_hours, d.last_seen_at
		FROM drives d
		LEFT JOIN LATERAL (
			SELECT sample_id, status FROM drive_health dh
			WHERE dh.drive_id = d.id
			ORDER BY sample_id DESC
			LIMIT 1
		) h ON true
		LEFT JOIN LATERAL (
			SELECT temperature, power_on_hours FROM smart_samples ss
			WHERE ss.drive_id = d.id
			ORDER BY id DESC
			LIMIT 1
		) s ON true
		ORDER BY d.device
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []DriveSummary{}
	for rows.Next() {
		var item DriveSummary
		if err := rows.Scan(&item.ID, &item.Device, &item.Model, &item.Serial, &item.Health, &item.Temperature, &item.PowerOnHrs, &item.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (d *DuckDB) GetDrive(ctx context.Context, id int64) (*DriveDetail, error) {
	row := d.db.QueryRowContext(ctx, `
		SELECT d.id, d.device, d.model, d.serial, d.wwn,
		       COALESCE(h.status, 'UNKNOWN'), COALESCE(h.score, 0), COALESCE(h.reasons, ''),
		       s.temperature, s.power_on_hours, s.reallocated_sectors, s.pending_sectors,
		       s.uncorrectable_sectors, s.wear_level, s.collected_at,
		       d.first_seen_at, d.last_seen_at
		FROM drives d
		LEFT JOIN LATERAL (
			SELECT sample_id, status, score, reasons
			FROM drive_health
			WHERE drive_id = d.id
			ORDER BY sample_id DESC
			LIMIT 1
		) h ON true
		LEFT JOIN LATERAL (
			SELECT id, temperature, power_on_hours, reallocated_sectors, pending_sectors, uncorrectable_sectors, wear_level, collected_at
			FROM smart_samples
			WHERE drive_id = d.id
			ORDER BY id DESC
			LIMIT 1
		) s ON true
		WHERE d.id = ?
	`, id)

	var item DriveDetail
	if err := row.Scan(
		&item.ID, &item.Device, &item.Model, &item.Serial, &item.WWN,
		&item.Health, &item.HealthScore, &item.HealthReasons,
		&item.Temperature, &item.PowerOnHours, &item.Reallocated, &item.Pending,
		&item.Uncorrectable, &item.WearLevel, &item.CollectedAt,
		&item.FirstSeen, &item.LastSeen,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (d *DuckDB) DriveHistory(ctx context.Context, id int64, limit int) ([]HistoryPoint, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := d.db.QueryContext(ctx, `
		SELECT collected_at, temperature, power_on_hours, reallocated_sectors,
		       pending_sectors, uncorrectable_sectors, wear_level
		FROM smart_samples
		WHERE drive_id = ?
		ORDER BY collected_at DESC
		LIMIT ?
	`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []HistoryPoint{}
	for rows.Next() {
		var p HistoryPoint
		if err := rows.Scan(&p.CollectedAt, &p.Temperature, &p.PowerOnHours, &p.ReallocatedSectors, &p.PendingSectors, &p.UncorrectableSectors, &p.WearLevel); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DuckDB) DriveAttributes(ctx context.Context, id int64) ([]AttributePoint, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT a.attribute_id, a.name, a.value, a.worst, a.threshold, a.raw
		FROM smart_attributes a
		JOIN smart_samples s ON s.id = a.sample_id
		WHERE s.drive_id = ?
		  AND s.id = (
			SELECT id FROM smart_samples WHERE drive_id = ? ORDER BY id DESC LIMIT 1
		  )
		ORDER BY a.attribute_id
	`, id, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []AttributePoint{}
	for rows.Next() {
		var item AttributePoint
		if err := rows.Scan(&item.AttributeID, &item.Name, &item.Value, &item.Worst, &item.Threshold, &item.Raw); err != nil {
			return nil, err
		}
		item.Status = classifyAttribute(item)
		out = append(out, item)
	}
	return out, rows.Err()
}

func classifyAttribute(a AttributePoint) string {
	if a.Threshold > 0 && a.Value <= a.Threshold {
		return "RED"
	}
	if a.Threshold > 0 && a.Value <= a.Threshold+5 {
		return "YELLOW"
	}
	return "GREEN"
}
