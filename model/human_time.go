package model

import (
	"github.com/dustin/go-humanize"
	"strings"
	"time"
)

func HumanTime(unixTime int64) string {
	if unixTime == 0 {
		return ""
	}

	unix := time.Unix(unixTime, 0)

	if time.Since(unix) < 1*time.Minute {
		return "now"
	}

	if time.Since(unix) > 7*24*time.Hour {
		return unix.Format("Jan 2")
	}

	replacements := map[string]string{
		" minutes": "m",
		" minute":  "m",
		" hours":   "h",
		" hour":    "h",
		" days":    "d",
		" day":     "d",
		" ago":     "",
	}

	human := humanize.Time(unix)

	for k, v := range replacements {
		human = strings.Replace(human, k, v, -1)
	}

	return human
}

func LongHumanTime(unixTime int64) string {
	if unixTime == 0 {
		return ""
	}

	unix := time.Unix(unixTime, 0)

	if time.Since(unix) < 1*time.Minute {
		return "now"
	}

	if time.Since(unix) > 7*24*time.Hour {
		return unix.Format("02 Jan 2006")
	}

	replacements := map[string]string{
		" ago": "",
	}

	human := humanize.Time(unix)

	for k, v := range replacements {
		human = strings.Replace(human, k, v, -1)
	}

	return human
}
