//go:build cgo

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"diskmon/internal/health"
	"diskmon/internal/smart"
)

func (d *DuckDB) InsertSample(ctx context.Context, info smart.DriveInfo, sample smart.SmartSample, result health.Result) (int64, error) {
	conn, err := d.db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	driveID, err := upsertDrive(ctx, tx, info, sample.CollectedAt)
	if err != nil {
		return 0, err
	}

	sampleID, err := insertSmartSample(ctx, tx, driveID, sample)
	if err != nil {
		return 0, err
	}

	if err := insertAttributes(ctx, tx, sampleID, sample.Attributes); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO drive_health (drive_id, sample_id, status, score, reasons) VALUES (?, ?, ?, ?, ?)`,
		driveID, sampleID, result.Status, result.Score, strings.Join(result.Reasons, "; ")); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return sampleID, nil
}

func upsertDrive(ctx context.Context, tx *sql.Tx, info smart.DriveInfo, seenAt time.Time) (int64, error) {
	var driveID int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM drives WHERE device = ?`, info.Device).Scan(&driveID)
	if err == nil {
		_, err = tx.ExecContext(ctx,
			`UPDATE drives SET model = ?, serial = ?, wwn = ?, last_seen_at = ? WHERE id = ?`,
			info.Model, info.Serial, info.WWN, seenAt, driveID)
		return driveID, err
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	if err := tx.QueryRowContext(ctx, `SELECT nextval('seq_drives')`).Scan(&driveID); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO drives (id, device, model, serial, wwn, first_seen_at, last_seen_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		driveID, info.Device, info.Model, info.Serial, info.WWN, seenAt, seenAt); err != nil {
		return 0, err
	}

	return driveID, nil
}

func insertSmartSample(ctx context.Context, tx *sql.Tx, driveID int64, sample smart.SmartSample) (int64, error) {
	var sampleID int64
	if err := tx.QueryRowContext(ctx, `SELECT nextval('seq_samples')`).Scan(&sampleID); err != nil {
		return 0, err
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO smart_samples (
			id, drive_id, collected_at, temperature, power_on_hours,
			reallocated_sectors, pending_sectors, uncorrectable_sectors, wear_level, raw_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		sampleID,
		driveID,
		sample.CollectedAt,
		nullInt(sample.Temperature),
		nullInt64(sample.PowerOnHours),
		nullInt64(sample.ReallocatedSectors),
		nullInt64(sample.PendingSectors),
		nullInt64(sample.UncorrectableSectors),
		nullInt64(sample.WearLevel),
		sample.RawJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("insert smart sample: %w", err)
	}
	return sampleID, nil
}

func insertAttributes(ctx context.Context, tx *sql.Tx, sampleID int64, attrs []smart.SmartAttribute) error {
	for _, attr := range attrs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO smart_attributes (sample_id, attribute_id, name, value, worst, threshold, raw) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			sampleID, attr.AttributeID, attr.Name, attr.Value, attr.Worst, attr.Threshold, attr.Raw)
		if err != nil {
			return err
		}
	}
	return nil
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}
