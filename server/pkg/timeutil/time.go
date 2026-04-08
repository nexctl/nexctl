package timeutil

import "time"

// NowUTC returns current UTC time.
func NowUTC() time.Time {
	return time.Now().UTC()
}
