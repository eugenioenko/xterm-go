package xterm

import (
	"bytes"
	"fmt"
	"testing"
)

func BenchmarkScrollRegion(b *testing.B) {
	const rows, cols = 44, 192
	cases := []struct{ name, setup string }{
		{"fullscreen", "\x1b[?1049h"},
		{"top_region", fmt.Sprintf("\x1b[?1049h\x1b[2;%dr", rows)},
		{"top_small_region", fmt.Sprintf("\x1b[?1049h\x1b[%d;%dr", rows/2, rows)},
		{"bottom_region", fmt.Sprintf("\x1b[?1049h\x1b[1;%dr", rows-1)},
		{"bottom_small_region", fmt.Sprintf("\x1b[?1049h\x1b[1;%dr", rows/2)},
		{"normal_buffer", ""},
	}
	payload := bytes.Repeat([]byte("y\n"), 2048)
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			t := New(WithCols(cols), WithRows(rows), WithScrollback(1000))
			t.Write([]byte(c.setup))
			b.SetBytes(int64(len(payload)))
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				t.Write(payload)
			}
		})
	}
}
