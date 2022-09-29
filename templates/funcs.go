package templates

import (
	"time"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC822)
}
