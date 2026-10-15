package jalali

import (
	"fmt"
	"time"
)

// GregorianToJalali converts a Gregorian date to Jalali (Solar Hijri / Shamsi)
func GregorianToJalali(gy, gm, gd int) (jy, jm, jd int) {
	gDM := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	var gy2 int
	if gm > 2 {
		gy2 = gy + 1
	} else {
		gy2 = gy
	}
	days := 355666 + (365 * gy) + ((gy2 + 3) / 4) - ((gy2 + 99) / 100) + ((gy2 + 399) / 400) + gd + gDM[gm-1]
	jy = -1595 + (33 * (days / 12053))
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	if days < 186 {
		jm = 1 + (days / 31)
		jd = 1 + (days % 31)
	} else {
		jm = 7 + ((days - 186) / 30)
		jd = 1 + ((days - 186) % 30)
	}
	return jy, jm, jd
}

// ToJalali converts time.Time into Jalali year, month, day
func ToJalali(t time.Time) (jy, jm, jd int) {
	// Offset to Tehran timezone if needed, or use local
	loc, err := time.LoadLocation("Asia/Tehran")
	if err == nil {
		t = t.In(loc)
	}
	return GregorianToJalali(t.Year(), int(t.Month()), t.Day())
}

// FormatDateJalali returns date in YYYY/MM/DD format (e.g. 1405/07/24)
func FormatDateJalali(t time.Time) string {
	jy, jm, jd := ToJalali(t)
	return fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd)
}

// FormatDateTimeJalali returns date and time in YYYY/MM/DD HH:MM format
func FormatDateTimeJalali(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err == nil {
		t = t.In(loc)
	}
	jy, jm, jd := GregorianToJalali(t.Year(), int(t.Month()), t.Day())
	return fmt.Sprintf("%04d/%02d/%02d %02d:%02d", jy, jm, jd, t.Hour(), t.Minute())
}
