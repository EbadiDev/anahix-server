package numtofa

import (
	"testing"
)

func TestNumberToPersianWords(t *testing.T) {
	tests := []struct {
		num      int64
		expected string
	}{
		{0, "صفر تومان"},
		{1000, "یک هزار تومان"},
		{50000, "پنجاه هزار تومان"},
		{150000, "یکصد و پنجاه هزار تومان"},
		{2500000, "دو میلیون و پانصد هزار تومان"},
		{12345678, "دوازده میلیون و سیصد و چهل و پنج هزار و ششصد و هفتاد و هشت تومان"},
	}

	for _, tt := range tests {
		actual := AmountToPersianWords(tt.num)
		if actual != tt.expected {
			t.Errorf("For %d, expected '%s', got '%s'", tt.num, tt.expected, actual)
		}
	}
}
