package prodos

import (
	"testing"
	"time"
)

func TestDateTimeToAndFromProDOS(t *testing.T) {
	now := time.Now().Round(time.Minute)

	got := DateTimeFromProDOS(DateTimeToProDOS(now))
	if got != now {
		t.Errorf("DateTimeFromProDOS(DateTimeToProDOS(now)) = %s; want %s", got.String(), now.String())
	}
}
