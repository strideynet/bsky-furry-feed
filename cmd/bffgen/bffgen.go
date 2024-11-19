package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

type FeedConfig struct {
	Name     string
	BaseType string // "new" or "hot"
}

func ConditionNotHidden() string {
	return "cp.is_hidden = FALSE"
}

func ConditionHasHashtag(hashtags []string) string {
	str := ""
	for i, h := range hashtags {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("'%s'", h)
	}
	return fmt.Sprintf(`ARRAY[%s] && cp.hashtags`, str)
}

func run() error {

	const baseType = "new"
	const feedName = "FurryNew"
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("-- name: Get%sFeed :many\n", feedName))
	if baseType == "new" {
		conditions := []string{
			ConditionNotHidden(),
			ConditionHasHashtag([]string{"furry"}),
		}

		for _, c := range conditions {
			sb.WriteString(fmt.Sprintf("AND %s\n", c))
		}
	}

	f, err := os.Create(".//store/queries/" + fmt.Sprintf("feed_%s.gen.sql", feedName))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	f.WriteString(sb.String())
	return nil
}
