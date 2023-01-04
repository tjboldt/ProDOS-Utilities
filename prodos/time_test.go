// Copyright Terence J. Boldt (c)2021-2023
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides tests for conversion to and from ProDOS time format

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
