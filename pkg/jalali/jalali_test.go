package jalali

import (
	"testing"
	"time"
)

func TestGregorianToJalali(t *testing.T) {
	// 2026-03-21 should be around 1405-01-01
	jy, jm, jd := GregorianToJalali(2026, 3, 21)
	if jy != 1405 || jm != 1 || jd != 1 {
		t.Errorf("Expected 1405/01/01, got %04d/%02d/%02d", jy, jm, jd)
	}

	date := time.Date(2026, time.October, 15, 14, 30, 0, 0, time.UTC)
	formatted := FormatDateJalali(date)
	if len(formatted) != 10 {
		t.Errorf("Expected length 10 formatted date, got %s", formatted)
	}
}
