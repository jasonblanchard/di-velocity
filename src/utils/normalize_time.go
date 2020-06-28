package utils

import "time"

// NormalizeTime returns nearest day instant from arbitrary time
func NormalizeTime(initialTime time.Time) time.Time {
	year := initialTime.Year()
	month := initialTime.Month()
	day := initialTime.Day()
	normalizedTime := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return normalizedTime
}
