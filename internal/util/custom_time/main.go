package customtime

import "time"

func TimeToString(date time.Time) string {
	return date.UTC().Format(time.RFC3339)
}

func ParseTime(timeStr string) (*time.Time, error) {
	parsedTime, err := time.Parse(time.RFC3339, timeStr)

	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}
