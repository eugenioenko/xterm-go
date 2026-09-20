package xterm

import (
	"reflect"
	"testing"
)

func TestScrollRegionReusesLinesWithoutLeakingContent(t *testing.T) {
	const rows, cols = 6, 10
	cases := []struct {
		name   string
		region string
		want   []string
	}{
		{"top at row 0", "\x1b[1;4r", []string{"row1", "row2", "row3", "", "row4", "row5"}},
		{"top below row 0", "\x1b[2;5r", []string{"row0", "row2", "row3", "row4", "", "row5"}},
		{"full screen", "\x1b[1;6r", []string{"row1", "row2", "row3", "row4", "row5", ""}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			term := New(WithCols(cols), WithRows(rows), WithScrollback(0))
			term.WriteString("\x1b[?1049h\x1b[41;31m")
			for y := 0; y < rows; y++ {
				term.WriteString("\x1b[" + string(rune('1'+y)) + ";1Hrow" + string(rune('0'+y)))
			}
			term.WriteString("\x1b[0m" + c.region)
			bottom := map[string]string{"\x1b[1;4r": "4", "\x1b[2;5r": "5", "\x1b[1;6r": "6"}[c.region]
			term.WriteString("\x1b[" + bottom + ";1H\n")

			buf := term.Buffer()
			var got []string
			for y := 0; y < rows; y++ {
				got = append(got, buf.TranslateBufferLineToString(buf.YBase+y, true, 0, cols))
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("rows = %q, want %q", got, c.want)
			}
		})
	}
}
