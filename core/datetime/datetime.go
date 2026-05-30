package datetime

import (
	"sync"
	"time"
)

const Layout = "2006-01-02 15:04:05"

type DateTime time.Time

var (
	locationMu sync.RWMutex
	location   = defaultLocation()
)

func SetLocation(loc *time.Location) {
	if loc == nil {
		return
	}
	locationMu.Lock()
	location = loc
	locationMu.Unlock()
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}

	locationMu.RLock()
	loc := location
	locationMu.RUnlock()

	formatted := t.In(loc).Format(Layout)
	return []byte(`"` + formatted + `"`), nil
}

func defaultLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return loc
}
