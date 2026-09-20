package xterm

import "testing"

func TestLoadCellDoesNotAllocateForPlainCells(t *testing.T) {
	term := New(WithCols(120), WithRows(40))
	term.WriteString("hello world")
	buf := term.Buffer()
	bl := buf.Lines.Get(buf.YBase)
	cd := NewCellData()
	allocs := testing.AllocsPerRun(100, func() {
		for x := 0; x < 120; x++ {
			bl.LoadCell(x, cd)
		}
	})
	if allocs != 0 {
		t.Fatalf("LoadCell allocated %v times per 120-cell line, want 0", allocs)
	}
}

func TestLoadCellKeepsExtendedAttrs(t *testing.T) {
	term := New(WithCols(20), WithRows(4))
	term.WriteString("a\x1b[4:3mb\x1b[0mc")
	buf := term.Buffer()
	bl := buf.Lines.Get(buf.YBase)
	cd := NewCellData()

	bl.LoadCell(0, cd)
	if cd.IsUnderline() != 0 || cd.Extended == nil || !cd.Extended.IsEmpty() {
		t.Fatalf("plain cell: underline=%d extended=%v", cd.IsUnderline(), cd.Extended)
	}
	bl.LoadCell(1, cd)
	if cd.IsUnderline() == 0 || cd.Extended == nil || cd.Extended.UnderlineStyle() != UnderlineStyleCurly {
		t.Fatalf("curly underline cell lost its extended attrs: %+v", cd.Extended)
	}
	bl.LoadCell(2, cd)
	if cd.IsUnderline() != 0 {
		t.Fatalf("cell after reset still underlined")
	}
}

func BenchmarkLoadCell(b *testing.B) {
	term := New(WithCols(120), WithRows(40))
	term.WriteString("hello world")
	buf := term.Buffer()
	bl := buf.Lines.Get(buf.YBase)
	cd := NewCellData()
	b.ReportAllocs()
	for b.Loop() {
		for x := 0; x < 120; x++ {
			bl.LoadCell(x, cd)
		}
	}
}
