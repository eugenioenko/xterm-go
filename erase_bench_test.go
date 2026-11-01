package xterm

import (
	"strings"
	"testing"
)

func BenchmarkErase(b *testing.B) {
	cases := []struct{ name, seq string }{
		{"display", "\x1b[2J"},
		{"line", "\x1b[5H\x1b[2K"},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			t := New(WithCols(192), WithRows(44))
			t.WriteString("\x1b[?1049h")
			payload := []byte(strings.Repeat(c.seq, 16))
			b.SetBytes(int64(len(payload)))
			b.ReportAllocs()
			for b.Loop() {
				t.Write(payload)
			}
		})
	}
}
