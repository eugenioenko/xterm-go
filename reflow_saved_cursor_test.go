package xterm

import (
	"fmt"
	"strings"
	"testing"
)

func savedCursorRow(term *Terminal) string {
	term.WriteString("\x1b8")
	buf := term.Buffer()
	return buf.Lines.Get(buf.YBase + buf.Y).TranslateToString(true, 0, -1)
}

func TestReflowSmallerRebasesSavedCursorWhenTrimming(t *testing.T) {
	t.Parallel()
	term := New(WithCols(80), WithRows(24), WithScrollback(100))
	for i := range 200 {
		term.WriteString(fmt.Sprintf("%d:%s\r\n", i, strings.Repeat("x", 66)))
	}
	term.WriteString("\x1b[10;1H\x1b[JANCHOR\x1b7\x1b[24;1H")

	term.Resize(40, 24)

	if got := savedCursorRow(term); got != "ANCHOR" {
		t.Fatalf("restored cursor row = %q, want %q", got, "ANCHOR")
	}
}

func TestReflowSmallerKeepsSavedCursorAboveReflowedLine(t *testing.T) {
	t.Parallel()
	term := New(WithCols(80), WithRows(24), WithScrollback(1000))
	term.WriteString("one\r\ntwo\r\nANCHOR\x1b7\x1b[5;1H" + strings.Repeat("y", 70) + "\x1b[24;1H")

	term.Resize(40, 24)

	if got := savedCursorRow(term); got != "ANCHOR" {
		t.Fatalf("restored cursor row = %q, want %q", got, "ANCHOR")
	}
}
