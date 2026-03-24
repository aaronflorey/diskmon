//go:build cgo

package storage

import (
	"context"
	"database/sql"
	"time"
)

func (d *DuckDB) GetNotificationState(ctx context.Context, driveID int64, notificationName string) (*NotificationState, error) {
	row := d.db.QueryRowContext(ctx, `
		SELECT drive_id, notification_name, state, updated_at
		FROM notification_state
		WHERE drive_id = ? AND notification_name = ?
	`, driveID, notificationName)

	var out NotificationState
	if err := row.Scan(&out.DriveID, &out.NotificationName, &out.State, &out.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (d *DuckDB) UpsertNotificationState(ctx context.Context, driveID int64, notificationName string, state string, updatedAt time.Time) error {
	conn, err := d.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO notification_state (drive_id, notification_name, state, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (drive_id, notification_name) DO UPDATE SET
			state = EXCLUDED.state,
			updated_at = EXCLUDED.updated_at
	`, driveID, notificationName, state, updatedAt); err != nil {
		return err
	}

	return tx.Commit()
}
