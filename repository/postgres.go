package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jasonblanchard/di-velocity/domain"
	"github.com/pkg/errors"
)

// Postgres postgres repository
type Postgres struct {
	Connection *sql.DB
}

// NewPostgres creates a new Postgres repository
func NewPostgres(user string, password string, dbname string, host string) (*Postgres, error) {
	postgres := &Postgres{}

	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", user, password, host, dbname)
	connection, err := sql.Open("postgres", connStr)
	// Ensure a healthy DB connection
	err = connection.Ping()
	if err != nil {
		return postgres, errors.Wrap(err, "Database connetion failed")
	}
	postgres.Connection = connection
	return postgres, nil
}

// DropDailyCounts Drops velocities table
func (p *Postgres) DropDailyCounts() error {
	rows, err := p.Connection.Query("DELETE FROM daily_counts")
	if rows != nil {
		defer rows.Close()
	}
	return errors.Wrap(err, "drop query failed")
}

// GetDailyCounts returns velocity score for each day between start and end (inclusive)
func (p *Postgres) GetDailyCounts(creatorID string, start time.Time, end time.Time) (domain.DailyCounts, error) {
	rows, err := p.Connection.Query(`
SELECT count, day, creator_id
FROM daily_counts
WHERE creator_id = $1
AND day >= $2
AND day <= $3
`, creatorID, start, end)
	if err != nil {
		return nil, errors.Wrap(err, "db query failed")
	}
	defer rows.Close()

	var dailyCounts = make(domain.DailyCounts, 0)

	for rows.Next() {
		dailyCount := domain.DailyCount{}
		if err := rows.Scan(&dailyCount.Count, &dailyCount.Day, &dailyCount.CreatorID); err != nil {
			return nil, errors.Wrap(err, "scan rows failed")
		}
		dailyCounts = append(dailyCounts, dailyCount)
	}

	return dailyCounts, nil
}

// IncrementDailyCounter Creates or increments a velocity event
func (p *Postgres) IncrementDailyCounter(day time.Time, creatorID string) error {
	rows, err := p.Connection.Query("SELECT id, count FROM daily_counts WHERE day = $1 AND creator_id = $2", day, creatorID)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "select query failed")
	}

	if rows.Next() {
		var id int
		var count int
		rows.Scan(&id, &count)
		count = count + 1
		_, err = p.Connection.Query("UPDATE daily_counts SET count = $1 WHERE id = $2", count, id)
		if err != nil {
			return errors.Wrap(err, "update failed")
		}
		return nil
	}

	_, err = p.Connection.Query("INSERT INTO daily_counts (day, count, creator_id) VALUES ($1, $2, $3)", day, 1, creatorID)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}

	return nil
}
