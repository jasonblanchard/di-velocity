package repository

import (
	"database/sql"

	"github.com/pkg/errors"
)

// DropDailyCounts Drops velocities table
func DropDailyCounts(db *sql.DB) error {
	rows, err := db.Query("DELETE FROM daily_counts")
	if rows != nil {
		defer rows.Close()
	}
	return errors.Wrap(err, "drop query failed")
}
