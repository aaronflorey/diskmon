//go:build cgo

package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

//go:embed schema.sql
var schemaSQL string

type DuckDB struct {
	db *sql.DB
}

func OpenDuckDB(path string) (*DuckDB, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &DuckDB{db: db}, nil
}

func (d *DuckDB) Close() error {
	return d.db.Close()
}

func (d *DuckDB) Conn(ctx context.Context) (*sql.Conn, error) {
	return d.db.Conn(ctx)
}

func (d *DuckDB) Ready(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
