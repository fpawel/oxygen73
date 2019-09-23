package app

import (
	"strings"
	"time"
)

func cutErrStr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if strings.Contains(s, ":") {
		return strings.Split(s, ":")[0]
	}
	return s
}

func pause(chDone <-chan struct{}, d time.Duration) {
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			return
		case <-chDone:
			timer.Stop()
			return
		}
	}
}

//func formatTimeAsQuery(t time.Time) string {
//	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
//		t.Format(timeLayout) + "'))"
//}
