package xterm

import (
	"bytes"
	"testing"
)

func BenchmarkWritePlainText(b *testing.B) {
	const rows, cols = 44, 192
	line := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 4)[:cols-12]
	var payload []byte
	for range rows * 4 {
		payload = append(payload, line...)
		payload = append(payload, '\r', '\n')
	}
	t := New(WithCols(cols), WithRows(rows), WithScrollback(1000))
	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()
	for b.Loop() {
		t.Write(payload)
	}
}
