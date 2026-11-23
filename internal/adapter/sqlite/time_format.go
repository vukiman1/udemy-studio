package sqlite

import "time"

func formatTime(moment time.Time) string { return moment.UTC().Format(time.RFC3339Nano) }

func parseTime(text string) (time.Time, error) { return time.Parse(time.RFC3339Nano, text) }
