package bluesky

import (
	"strings"
	"time"
)

var timeFormat = "2006-01-02T15:04:05.999999999Z"

func ParseTime(str string) (time.Time, error) {
	if strings.HasSuffix(str, "Z") {
		return parseZulu(str)
	} else {
		return parseWithOffset(str)
	}
}

func parseZulu(str string) (time.Time, error) {
	t, err := time.Parse(timeFormat, str)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func parseWithOffset(str string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func FormatTime(t time.Time) string {
	return t.UTC().Format(timeFormat)
}
