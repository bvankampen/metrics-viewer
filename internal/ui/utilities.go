package ui

import (
	"fmt"
	"time"
)

func ToHumanAgo(input interface{}, currentUnixTime int64) (interface{}, error) {
	timestamp, ok := input.(int64)
	if !ok {
		return nil, fmt.Errorf("value is not a valid timestamp: %v", input)
	}
	duration := time.Duration(currentUnixTime-timestamp) * time.Second
	return HumanDuration(duration), nil
}

func HumanDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%d minutes ago", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%d hours ago", seconds/3600)
	case seconds < 2592000:
		return fmt.Sprintf("%d days ago", seconds/86400)
	case seconds < 31536000:
		return fmt.Sprintf("%d months ago", seconds/2592000)
	default:
		return fmt.Sprintf("%d years ago", seconds/31536000)
	}
}

// ToISO8601 converts a Unix timestamp to an ISO 8601 formatted string.
func ToISO8601(unixTime int64) string {
	// Convert the Unix timestamp to a time.Time object
	t := time.Unix(unixTime, 0)
	// Format the time in ISO 8601 format
	return t.Format(time.RFC3339)
}
