package xterm

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRepeatedSGRDoesNotAllocateExtendedAttrs(t *testing.T) {
	term := New(WithCols(20), WithRows(4))
	for _, seq := range []string{"\x1b[4m", "\x1b[4:3m", "\x1b[58;5;196m", "\x1b[59m"} {
		term.WriteString(seq)
		allocs := testing.AllocsPerRun(100, func() { term.WriteString(seq) })
		if allocs != 0 {
			t.Errorf("repeating %q allocated %v times, want 0", seq, allocs)
		}
	}
}

func TestSGRChangesDoNotMutateStoredCells(t *testing.T) {
	term := New(WithCols(20), WithRows(4))
	term.WriteString("\x1b[4:3ma\x1b[4:1mb\x1b[58;5;196mc\x1b[0md")
	buf := term.Buffer()
	bl := buf.Lines.Get(buf.YBase)
	cd := NewCellData()

	want := []struct {
		style UnderlineStyle
		color bool
	}{{UnderlineStyleCurly, false}, {UnderlineStyleSingle, false}, {UnderlineStyleSingle, true}, {UnderlineStyleNone, false}}
	for x, w := range want {
		bl.LoadCell(x, cd)
		if got := cd.Extended.UnderlineStyle(); got != w.style {
			t.Errorf("cell %d underline style = %v, want %v", x, got, w.style)
		}
		if got := cd.Extended.UnderlineColor() != 0; got != w.color {
			t.Errorf("cell %d has underline color = %v, want %v", x, got, w.color)
		}
	}
}

func BenchmarkSGRUnderlineCells(b *testing.B) {
	var payload bytes.Buffer
	for i := 0; i < 8448; i++ {
		fmt.Fprintf(&payload, "\x1b[38;5;%d;48;5;%d;1;3;4mX", i%156+100, 255-i%156+100)
	}
	data := payload.Bytes()
	term := New(WithCols(192), WithRows(44))
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for b.Loop() {
		term.WriteString("\x1b[H")
		for off := 0; off < len(data); off += 4096 {
			term.Write(data[off:min(off+4096, len(data))])
		}
	}
}
