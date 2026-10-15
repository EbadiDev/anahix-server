package numtofa

import (
	"fmt"
	"strings"
)

var (
	ones = []string{
		"", "یک", "دو", "سه", "چهار", "پنج", "شش", "هفت", "هشت", "نه",
	}
	teens = []string{
		"ده", "یازده", "دوازده", "سیزده", "چهارده", "پانزده", "شانزده", "هفده", "هجده", "نوزده",
	}
	tens = []string{
		"", "", "بیست", "سی", "چهل", "پنجاه", "شصت", "هفتاد", "هشتاد", "نود",
	}
	hundreds = []string{
		"", "یکصد", "دویست", "سیصد", "چهارصد", "پانصد", "ششصد", "هفتصد", "هشتصد", "نهصد",
	}
	thousands = []string{
		"", "هزار", "میلیون", "میلیارد", "تریلیون",
	}
)

func convertThreeDigits(n int) string {
	if n == 0 {
		return ""
	}

	var parts []string

	h := n / 100
	rem := n % 100

	if h > 0 {
		parts = append(parts, hundreds[h])
	}

	if rem >= 10 && rem <= 19 {
		parts = append(parts, teens[rem-10])
	} else {
		t := rem / 10
		o := rem % 10
		if t > 0 {
			parts = append(parts, tens[t])
		}
		if o > 0 {
			parts = append(parts, ones[o])
		}
	}

	return strings.Join(parts, " و ")
}

// NumberToPersianWords converts an integer to Persian text words
func NumberToPersianWords(n int64) string {
	if n == 0 {
		return "صفر"
	}
	if n < 0 {
		return "منفی " + NumberToPersianWords(-n)
	}

	var chunks []int
	temp := n
	for temp > 0 {
		chunks = append(chunks, int(temp%1000))
		temp /= 1000
	}

	var parts []string
	for i := len(chunks) - 1; i >= 0; i-- {
		val := chunks[i]
		if val == 0 {
			continue
		}
		text := convertThreeDigits(val)
		if text != "" {
			if thousands[i] != "" {
				text = fmt.Sprintf("%s %s", text, thousands[i])
			}
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " و ")
}

// AmountToPersianWords converts an Iranian Toman amount to Persian words with "تومان" suffix
func AmountToPersianWords(amountToman int64) string {
	if amountToman == 0 {
		return "صفر تومان"
	}
	words := NumberToPersianWords(amountToman)
	return fmt.Sprintf("%s تومان", words)
}
