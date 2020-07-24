package repository

import (
	"time"

	"github.com/jasonblanchard/di-velocity/domain"
)

// T interface for reposutory functions
type T interface {
	DropDailyCounts() error
	GetDailyCounts(start time.Time, end time.Time) (domain.DailyCounts, error)
	IncrementDailyCounter(day time.Time, creatorID string) error
}
