//go:build cgo

package storage

import "context"

func (d *DuckDB) DeleteIncompleteSmartTestRuns(ctx context.Context) (int64, error) {
	res, err := d.db.ExecContext(ctx, `
		DELETE FROM smart_test_runs
		WHERE UPPER(status) NOT IN ('FAILED', 'PASSED', 'SUCCESS', 'COMPLETED')
	`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
