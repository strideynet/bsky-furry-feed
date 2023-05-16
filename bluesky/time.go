package bluesky

import "time"

var timeFormat = "2006-01-02T15:04:05.999999999Z"

func ParseTime(str string) (time.Time, error) {
	t, err := time.Parse(timeFormat, str)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func FormatTime(t time.Time) string {
	return t.UTC().Format(timeFormat)
}
