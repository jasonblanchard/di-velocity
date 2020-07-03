package op

import (
	"database/sql"
)

// DropDailyCounts Drops velocities table
func DropDailyCounts(db *sql.DB) error {
	rows, err := db.Query("DELETE FROM daily_counts")
	if rows != nil {
		defer rows.Close()
	}
	return err
}
