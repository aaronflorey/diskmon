//go:build cgo

package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestGetNotificationStateMissing(t *testing.T) {
	db := openTestDuckDB(t)
	t.Cleanup(func() { _ = db.Close() })

	got, err := db.GetNotificationState(context.Background(), 1, "discord-primary")
	if err != nil {
		t.Fatalf("GetNotificationState returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil state for missing row, got %+v", *got)
	}
}

func TestUpsertNotificationStateCreatesAndUpdatesSingleRow(t *testing.T) {
	db := openTestDuckDB(t)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	insertTestDrive(t, db, 7, "/dev/sda", time.Date(2026, 3, 20, 1, 0, 0, 0, time.UTC))

	firstAt := time.Date(2026, 3, 20, 2, 0, 0, 0, time.UTC)
	if err := db.UpsertNotificationState(ctx, 7, "discord-primary", "PASS", firstAt); err != nil {
		t.Fatalf("first UpsertNotificationState returned error: %v", err)
	}

	got, err := db.GetNotificationState(ctx, 7, "discord-primary")
	if err != nil {
		t.Fatalf("GetNotificationState returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected state row after first upsert, got nil")
	}
	if got.State != "PASS" {
		t.Fatalf("expected PASS state, got %q", got.State)
	}
	if !got.UpdatedAt.Equal(firstAt) {
		t.Fatalf("expected updated_at %s, got %s", firstAt, got.UpdatedAt)
	}

	secondAt := firstAt.Add(5 * time.Minute)
	if err := db.UpsertNotificationState(ctx, 7, "discord-primary", "FAIL", secondAt); err != nil {
		t.Fatalf("second UpsertNotificationState returned error: %v", err)
	}

	got, err = db.GetNotificationState(ctx, 7, "discord-primary")
	if err != nil {
		t.Fatalf("GetNotificationState after update returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected state row after update, got nil")
	}
	if got.State != "FAIL" {
		t.Fatalf("expected FAIL state after update, got %q", got.State)
	}
	if !got.UpdatedAt.Equal(secondAt) {
		t.Fatalf("expected updated_at %s after update, got %s", secondAt, got.UpdatedAt)
	}

	var count int
	if err := db.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notification_state WHERE drive_id = ? AND notification_name = ?`,
		7, "discord-primary",
	).Scan(&count); err != nil {
		t.Fatalf("count query returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 row for key after repeated upserts, got %d", count)
	}
}

func openTestDuckDB(t *testing.T) *DuckDB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "storage-test.duckdb")
	db, err := OpenDuckDB(path)
	if err != nil {
		t.Fatalf("OpenDuckDB(%q) error: %v", path, err)
	}
	return db
}

func insertTestDrive(t *testing.T, db *DuckDB, id int64, device string, seenAt time.Time) {
	t.Helper()

	_, err := db.db.Exec(`
		INSERT INTO drives (id, device, model, serial, wwn, first_seen_at, last_seen_at)
		VALUES (?, ?, '', '', '', ?, ?)
	`, id, device, seenAt, seenAt)
	if err != nil {
		t.Fatalf("insert test drive: %v", err)
	}
}
