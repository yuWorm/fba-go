package datetime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/datetime"
)

func TestDateTimeMarshalsWithConfiguredLocation(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	datetime.SetLocation(loc)
	t.Cleanup(func() {
		datetime.SetLocation(time.UTC)
	})

	value := datetime.DateTime(time.Date(2026, 5, 30, 7, 8, 9, 0, time.UTC))
	got, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	const want = `"2026-05-30 15:08:09"`
	if string(got) != want {
		t.Fatalf("DateTime JSON = %s, want %s", got, want)
	}
}

func TestDateTimeZeroMarshalsNull(t *testing.T) {
	var value datetime.DateTime
	got, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	const want = `null`
	if string(got) != want {
		t.Fatalf("zero DateTime JSON = %s, want %s", got, want)
	}
}
