package utils

import (
	"time"
	"github.com/charmbracelet/lipgloss"
)

var TIMESTAMP_STYLE lipgloss.Style = lipgloss.NewStyle().Faint(true).PaddingTop(1)

func FormatTimestamp(timestamp string) string {
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	layout := "2006-01-02 15:04:05"
	parsed := parsedTime.Format(layout)

	return TIMESTAMP_STYLE.Render(parsed) + "\n"
}
