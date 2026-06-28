package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/doze-dev/doze/internal/control"
)

func key(s string) tea.KeyMsg {
	if s == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func threeInstances() model {
	fi := textinput.New()
	fi.Prompt = "/"
	return model{
		width: 110, height: 30,
		follow: true,
		filter: fi,
		hist:   map[string]*history{},
		logVP:  viewport.New(40, 8),
		resp: control.Response{
			Listen: "127.0.0.1:6432",
			Instances: []control.InstanceView{
				{Name: "app", Engine: "postgres", State: "active", Conns: 1},
				{Name: "cache", Engine: "valkey", State: "idle"},
				{Name: "media", Engine: "s3", State: "reaped", LastError: "boom"},
			},
		},
	}
}

func send(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func TestCursorNavigationClamps(t *testing.T) {
	m := threeInstances()
	m = send(m, key("j"))
	m = send(m, key("j"))
	m = send(m, key("j")) // clamp at last
	if m.cursor != 2 {
		t.Fatalf("cursor = %d, want 2 (clamped)", m.cursor)
	}
	m = send(m, key("k"))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
}

func TestSelectionIsNameSorted(t *testing.T) {
	m := threeInstances()
	// cursor 0 should be the alphabetically-first instance regardless of input order.
	if v, _ := m.selected(); v.Name != "app" {
		t.Fatalf("first selection = %q, want app", v.Name)
	}
	m.cursor = 2
	if v, _ := m.selected(); v.Name != "media" {
		t.Fatalf("last selection = %q, want media", v.Name)
	}
}

func TestFilterToggleAndClear(t *testing.T) {
	m := threeInstances()
	m = send(m, key("/"))
	if !m.filtering {
		t.Fatal("'/' should enter filter mode")
	}
	m = send(m, key("z")) // matches nothing
	if len(m.visible()) != 0 {
		t.Fatalf("filter 'z' should hide all, got %d", len(m.visible()))
	}
	m = send(m, key("esc"))
	if m.filtering {
		t.Fatal("esc should leave filter mode")
	}
	if len(m.visible()) != 3 {
		t.Fatalf("esc should clear the filter, got %d visible", len(m.visible()))
	}
}

// sqsActs mirrors the sqs engine's published actions.
func sqsActs() []control.ActionView {
	return []control.ActionView{
		{ID: "peek", Label: "Peek", Kind: "queue"},
		{ID: "send", Label: "Send", Kind: "queue", InputHint: "message body"},
		{ID: "purge", Label: "Purge", Kind: "queue", Destructive: true},
		{ID: "redrive", Label: "Redrive", Kind: "queue"},
	}
}

func TestViewerActionAndKeys(t *testing.T) {
	m := model{adminActs: sqsActs()}
	if v := m.viewerActionID(); v != "peek" {
		t.Fatalf("viewer = %q, want peek (first non-destructive, no-input)", v)
	}
	km := m.adminActionKeys()
	for id, want := range map[string]string{"send": "s", "purge": "p", "redrive": "r"} {
		if km[id] != want {
			t.Fatalf("key for %q = %q, want %q", id, km[id], want)
		}
	}
	if _, ok := km["peek"]; ok {
		t.Fatal("the viewer action should not get a letter key (it's enter)")
	}
	if a, ok := m.actionForKey("p"); !ok || a.ID != "purge" {
		t.Fatalf("key p → %v/%v, want purge", a.ID, ok)
	}
}

func TestInvokeActionStaging(t *testing.T) {
	m := model{
		cmd:         textinput.New(),
		adminMode:   true,
		adminActs:   sqsActs(),
		adminRes:    []control.ResourceView{{Kind: "queue", Name: "emails"}},
		adminCursor: 0,
	}
	get := func(id string) control.ActionView {
		for _, a := range m.adminActs {
			if a.ID == id {
				return a
			}
		}
		t.Fatalf("no action %q", id)
		return control.ActionView{}
	}
	// Destructive → confirm prompt, not an immediate run.
	nm, _ := m.invokeAction(get("purge"))
	if mm := nm.(model); mm.adminConfirm != "purge" {
		t.Fatalf("purge should stage a confirm, got %q", mm.adminConfirm)
	}
	// Input action → input prompt.
	nm, _ = m.invokeAction(get("send"))
	if mm := nm.(model); mm.adminInput != "send" {
		t.Fatalf("send should open an input prompt, got %q", mm.adminInput)
	}
}

func TestCharSelectionText(t *testing.T) {
	m := model{copyCharMode: true, copyLines: []string{"hello world", "second line"}}
	// single line: [0,0)→[0,5) = "hello"
	m.copyAnchor, m.copyAnchorColCh = 0, 0
	m.copyCursor, m.copyColCh = 0, 5
	if got := m.selectedCharText(); got != "hello" {
		t.Fatalf("single-line char select = %q, want hello", got)
	}
	// reversed drag (cursor before anchor) still orders correctly
	m.copyAnchor, m.copyAnchorColCh = 0, 5
	m.copyCursor, m.copyColCh = 0, 0
	if got := m.selectedCharText(); got != "hello" {
		t.Fatalf("reversed char select = %q, want hello", got)
	}
	// multi-line: from (0,6) to (1,6) = "world\nsecond"
	m.copyAnchor, m.copyAnchorColCh = 0, 6
	m.copyCursor, m.copyColCh = 1, 6
	if got := m.selectedCharText(); got != "world\nsecond" {
		t.Fatalf("multi-line char select = %q, want %q", got, "world\nsecond")
	}
}

func TestWorkspaceRenders(t *testing.T) {
	m := threeInstances()
	m.cursor = 2 // media (s3)
	m.adminMode = true
	m.adminName = "media"
	m.adminActs = []control.ActionView{
		{ID: "browse", Label: "Browse", Kind: "bucket"},
		{ID: "empty", Label: "Empty", Kind: "bucket", Destructive: true},
	}
	m.adminRes = []control.ResourceView{{Kind: "bucket", Name: "uploads", Status: "3 objects"}}
	m.adminVP = viewport.New(40, 8)
	out := m.View()
	for _, want := range []string{"media", "uploads", "browse", "empty", "esc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("workspace view missing %q:\n%s", want, out)
		}
	}
}

func TestViewRendersInstancesAndKeys(t *testing.T) {
	m := threeInstances()
	out := m.View()
	for _, want := range []string{"app", "cache", "media", "boot", "reap", "follow", "doze"} {
		if !strings.Contains(out, want) {
			t.Fatalf("view missing %q:\n%s", want, out)
		}
	}
	// The full key set (incl. restart) and the mouse/state legend live in `?` help.
	m.showHelp = true
	help := m.View()
	for _, want := range []string{"restart", "Mouse", "asleep", "cycle theme"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help overlay missing %q:\n%s", want, help)
		}
	}
	m.showHelp = false
	// Selecting the failed instance surfaces its error detail.
	m.cursor = 2
	if !strings.Contains(m.View(), "boom") {
		t.Fatalf("view should surface the selected instance's error:\n%s", m.View())
	}
}
