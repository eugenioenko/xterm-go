package xterm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnicodeServiceWcwidth(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		Name     string
		Input    rune
		Expected int
	}
	tests := []TestCase{
		// Control characters: width 0
		{"NUL", 0x00, 0},
		{"BEL", 0x07, 0},
		{"BS", 0x08, 0},
		{"TAB", 0x09, 0},
		{"LF", 0x0A, 0},
		{"ESC", 0x1B, 0},
		{"DEL", 0x7F, 0},
		{"C1 control", 0x80, 0},
		{"C1 end", 0x9F, 0},

		// ASCII printable: width 1
		{"space", ' ', 1},
		{"A", 'A', 1},
		{"z", 'z', 1},
		{"tilde", '~', 1},

		// Latin extended: width 1
		{"a-umlaut", '\u00e4', 1},
		{"n-tilde", '\u00f1', 1},

		// CJK wide characters: width 2
		{"CJK ideograph", '\u4e00', 2},
		{"CJK ideograph 2", '\u9fff', 2},
		{"Hangul syllable", '\uac00', 2},
		{"Hangul syllable end", '\ud7a3', 2},
		{"Fullwidth A", '\uff21', 2},
		{"Fullwidth excl", '\uff01', 2},
		{"CJK compat", '\uf900', 2},

		// Korean Jamo: width 2
		{"Jamo start", '\u1100', 2},
		{"Jamo end", '\u115f', 2},

		// Combining marks: width 0
		{"combining acute", '\u0301', 0},
		{"combining tilde", '\u0303', 0},
		{"combining diaeresis", '\u0308', 0},
		{"Thai combining", '\u0e34', 0},
		{"variation selector", '\ufe0f', 0},
		{"zero-width space", '\u200b', 0},
		{"ZWNJ", '\u200c', 0},
		{"ZWJ", '\u200d', 0},

		// Special cases
		{"left angle bracket", '\u2329', 2},
		{"right angle bracket", '\u232a', 2},
		{"0x303f exception", '\u303f', 1},

		// Supplementary planes
		{"CJK ext B", 0x20000, 2},
		{"CJK ext B end", 0x2fffd, 2},
		{"SMP char", 0x10000, 1},
		{"high combining", 0x1D167, 0},
	}
	u := NewUnicodeService()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			type Expectation struct {
				Width int
			}
			got := Expectation{Width: u.Wcwidth(tc.Input)}
			expected := Expectation{Width: tc.Expected}
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnicodeServiceGetStringCellWidth(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		Name     string
		Input    string
		Expected int
	}
	tests := []TestCase{
		{"empty string", "", 0},
		{"ASCII", "hello", 5},
		{"CJK", "你好", 4},                   //nolint:gosmopolitan
		{"mixed ASCII and CJK", "hi你好", 6}, //nolint:gosmopolitan
		{"combining marks", "e\u0301", 1},  // é as e + combining acute
		{"fullwidth", "\uff21\uff22", 4},
		{"control chars ignored", "\x00\x01\x02", 0},
		{"tab and newline", "\t\n", 0},
		{"emoji supplementary", "\U0001F600", 2},
		{"CJK ext B string", "\U00020000", 2},
		{"flag pair", "\U0001F1FA\U0001F1F8", 2}, // two regional indicators, one column each
	}
	u := NewUnicodeService()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			type Expectation struct {
				Width int
			}
			got := Expectation{Width: u.GetStringCellWidth(tc.Input)}
			expected := Expectation{Width: tc.Expected}
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnicodeServiceWcwidthBMPBoundaries(t *testing.T) {
	t.Parallel()

	// Verify boundary conditions of the BMP table
	type TestCase struct {
		Name     string
		Input    rune
		Expected int
	}
	tests := []TestCase{
		{"just below BMP", 0xFFFF, 1},
		{"at BMP boundary", 0x10000, 1},
		{"last wide CJK compat", 0xFAFF, 2},
		{"first after CJK compat", 0xFB00, 1},
		{"fullwidth end", 0xFF60, 2},
		{"halfwidth start", 0xFF61, 1},
	}
	u := NewUnicodeService()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			type Expectation struct {
				Width int
			}
			got := Expectation{Width: u.Wcwidth(tc.Input)}
			expected := Expectation{Width: tc.Expected}
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnicodeServiceWcwidthEmoji(t *testing.T) {
	t.Parallel()

	// Widths follow the Unicode 11 East Asian Wide and emoji presentation
	// ranges (the same set as xterm.js's unicode11 addon).
	type TestCase struct {
		Name     string
		Input    rune
		Expected int
	}
	tests := []TestCase{
		// Emoji presentation characters: width 2
		{"grinning face", 0x1F600, 2},
		{"white heavy check mark", 0x2705, 2},
		{"thumbs up", 0x1F44D, 2},
		{"watch", 0x231A, 2},
		{"star", 0x2B50, 2},
		{"rocket", 0x1F680, 2},

		// Text presentation characters stay narrow without VS16
		{"heavy check mark text", 0x2714, 1},
		{"red heart", 0x2764, 1},
		{"scissors", 0x2702, 1},

		// Regional indicators are one column each; a pair totals two.
		{"regional indicator A", 0x1F1E6, 1},
		{"regional indicator U", 0x1F1FA, 1},

		// Zero-width sequences and combining marks stay width 0.
		{"VS15", 0xFE0E, 0},
		{"VS16", 0xFE0F, 0},
		{"ZWJ", 0x200D, 0},
		{"combining acute", 0x0301, 0},

		// CJK stays wide.
		{"CJK", 0x4E00, 2},
		{"Hangul jamo extended", 0xA960, 2},
	}
	u := NewUnicodeService()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if got := u.Wcwidth(tc.Input); got != tc.Expected {
				t.Errorf("Wcwidth(%U) = %d, want %d", tc.Input, got, tc.Expected)
			}
		})
	}
}

func TestUnicodeServiceHighCombining(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		Name     string
		Input    rune
		Expected int
	}
	tests := []TestCase{
		{"musical combining", 0x1D167, 0},
		{"musical combining end", 0x1D169, 0},
		{"variation selector supplement start", 0xE0100, 0},
		{"variation selector supplement end", 0xE01EF, 0},
		{"tag space", 0xE0020, 0},
		{"language tag", 0xE0001, 0},
		{"outside high combining", 0x1D166, 1},
	}
	u := NewUnicodeService()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			type Expectation struct {
				Width int
			}
			got := Expectation{Width: u.Wcwidth(tc.Input)}
			expected := Expectation{Width: tc.Expected}
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
