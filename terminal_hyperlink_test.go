package xterm_test

import (
	"testing"

	xterm "github.com/eugenioenko/xterm-go"
)

func TestTerminalGetLinkData(t *testing.T) {
	for _, tc := range []struct {
		name       string
		terminator string
		id         string
	}{
		{"ST", "\x1b\\", ""},
		{"BEL", "\a", ""},
		{"ST with ID", "\x1b\\", "example"},
		{"BEL with ID", "\a", "example"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := xterm.New(xterm.WithCols(40), xterm.WithRows(2))
			defer term.Dispose()
			params := ""
			if tc.id != "" {
				params = "id=" + tc.id
			}
			term.WriteString("\x1b]8;" + params + ";https://example.com" + tc.terminator +
				"Click here\x1b]8;;" + tc.terminator + " plain")

			line := term.Buffer().Lines.Get(0)
			cell := xterm.NewCellData()
			line.LoadCell(0, cell)
			linkID := cell.Extended.URLID()
			if linkID == 0 {
				t.Fatal("linked cell has no hyperlink ID")
			}
			want := xterm.OscLinkData{URI: "https://example.com", ID: tc.id}
			for col := 0; col < 10; col++ {
				line.LoadCell(col, cell)
				if got := cell.Extended.URLID(); got != linkID {
					t.Fatalf("cell %d: link ID = %d, want %d", col, got, linkID)
				}
				if got := term.GetLinkData(cell.Extended.URLID()); got == nil || *got != want {
					t.Fatalf("cell %d: link data = %+v, want %+v", col, got, want)
				}
			}
			line.LoadCell(10, cell)
			if got := cell.Extended.URLID(); got != 0 {
				t.Fatalf("plain cell has hyperlink ID %d", got)
			}
			for _, id := range []int{0, -1, linkID + 1} {
				if got := term.GetLinkData(id); got != nil {
					t.Fatalf("unknown ID %d: link data = %+v, want nil", id, got)
				}
			}
			data := term.GetLinkData(linkID)
			data.URI = "changed"
			data.ID = "changed"
			if got := term.GetLinkData(linkID); got == nil || *got != want {
				t.Fatalf("mutating returned data changed stored link: %+v", got)
			}
		})
	}
}

func TestTerminalGetLinkDataScrollback(t *testing.T) {
	term := xterm.New(xterm.WithCols(20), xterm.WithRows(2), xterm.WithScrollback(1))
	defer term.Dispose()
	term.WriteString("\x1b]8;;https://example.com\aLink\x1b]8;;\a")
	cell := xterm.NewCellData()
	term.Buffer().Lines.Get(0).LoadCell(0, cell)
	linkID := cell.Extended.URLID()
	term.WriteString("\r\nsecond\r\nthird")
	if term.Buffer().YBase != 1 {
		t.Fatal("expected the link's line to be in scrollback")
	}
	if got := term.GetLinkData(linkID); got == nil || got.URI != "https://example.com" {
		t.Fatalf("scrollback link data = %+v", got)
	}
	term.WriteString("\r\nfourth")
	if got := term.GetLinkData(linkID); got != nil {
		t.Fatalf("expired link data = %+v, want nil", got)
	}
}
