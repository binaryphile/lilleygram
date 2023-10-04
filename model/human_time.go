package model

import (
	"github.com/dustin/go-humanize"
	"time"
)

func HumanTime(unixTime int64) string {
	if unixTime == 0 {
		return ""
	}

	unix := time.Unix(unixTime, 0)

	if time.Since(unix) > 48*time.Hour {
		return unix.Format("02 Jan 2006 03:04PM")
	}

	return humanize.Time(unix)
}
