// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/sbtlocalization/dcanvas"
)

// model is the Bubble Tea state for the view TUI. It wraps the pure tree
// (tree.go) with an expand state, a cursor over the currently visible lines, a
// jump-history stack, and a viewport that scrolls the tree pane. All display
// derives from these fields; the tree logic itself knows nothing of the UI.
type model struct {
	forest    []*item        // the walkable dialogue, built once
	collapsed map[*item]bool // branches the user has closed (absent = open)
	lines     []line         // flattened visible rows under the current state
	cursor    int            // index into lines
	jumpStack []*item        // origins of jumps, for Backspace

	ready            bool
	width, height    int
	side             bool // true → tree and detail side by side, else stacked
	treeW, treeH     int
	detailW, detailH int

	tree viewport.Model
}

// sideBySideMin is the terminal width at or above which the tree and detail
// panes sit side by side; below it they stack vertically. A starting value,
// tuned on real data.
const sideBySideMin = 90

// --- styling ----------------------------------------------------------------

var (
	lineStyle     = lipgloss.NewStyle()                                   // NPC line / narrator
	replyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("150")) // player reply
	jumpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("24")).
			Foreground(lipgloss.Color("231")).
			Bold(true)

	sepStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ruleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	detailStyle = lipgloss.NewStyle()
	detailLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Bold(true)

	barStyle = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("252"))
	barKey   = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("220")).Bold(true)
)

// newModel builds the tree from the canvas and starts fully expanded (a reading
// tool shows the whole flow; the user collapses long branches as needed).
func newModel(c *dcanvas.Canvas) model {
	m := model{
		forest:    build(c),
		collapsed: make(map[*item]bool),
		tree:      viewport.New(0, 0),
	}
	m.rebuild()
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m.resize()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey applies one keypress. Arrow keys drive the cursor and
// expand/collapse; Enter is context-sensitive; Backspace unwinds jumps. Every
// non-quit key ends by re-rendering the tree and keeping the cursor on screen.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+c":
		return m, tea.Quit
	}
	if len(m.lines) == 0 {
		return m, nil
	}

	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.lines)-1 {
			m.cursor++
		}
	case "home", "ctrl+home":
		m.cursor = 0
	case "end", "ctrl+end":
		m.cursor = len(m.lines) - 1
	case "right":
		l := m.lines[m.cursor]
		if l.expandable && !l.expanded {
			delete(m.collapsed, l.item) // expand
			m.rebuild()
		} else if l.expandable && l.expanded {
			if m.cursor < len(m.lines)-1 { // move into first child
				m.cursor++
			}
		}
	case "left":
		l := m.lines[m.cursor]
		if l.expandable && l.expanded {
			m.collapsed[l.item] = true // collapse
			m.rebuild()
		} else if l.item.parent != nil { // move to parent
			if idx := m.lineIndex(l.item.parent); idx >= 0 {
				m.cursor = idx
			}
		}
	case "enter":
		l := m.lines[m.cursor]
		switch {
		case l.item.jump:
			m.jump(l.item)
		case l.expandable:
			if l.expanded {
				m.collapsed[l.item] = true
			} else {
				delete(m.collapsed, l.item)
			}
			m.rebuild()
		}
	case "backspace":
		if n := len(m.jumpStack); n > 0 {
			origin := m.jumpStack[n-1]
			m.jumpStack = m.jumpStack[:n-1]
			if idx := m.lineIndex(origin); idx >= 0 {
				m.cursor = idx
			}
		}
	}

	m.renderTree()
	m.ensureVisible()
	return m, nil
}

// jump moves to the canonical occurrence of a jump line's target, auto-expanding
// the target's ancestors so it is never hidden, and records the origin so
// Backspace can return. A jump with no resolved target is a no-op.
func (m *model) jump(j *item) {
	if j.target == nil {
		return
	}
	for a := j.target.parent; a != nil; a = a.parent {
		delete(m.collapsed, a)
	}
	m.jumpStack = append(m.jumpStack, j)
	m.rebuild()
	if idx := m.lineIndex(j.target); idx >= 0 {
		m.cursor = idx
	}
}

// rebuild recomputes the visible lines under the current expand state and keeps
// the cursor in range.
func (m *model) rebuild() {
	m.lines = visibleLines(m.forest, m.collapsed)
	if m.cursor >= len(m.lines) {
		m.cursor = len(m.lines) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m model) lineIndex(it *item) int {
	for i, l := range m.lines {
		if l.item == it {
			return i
		}
	}
	return -1
}

// resize recomputes the pane geometry from the terminal size and re-renders the
// tree into its viewport.
func (m *model) resize() {
	contentH := max(m.height-1, 1) // one row for the status bar
	if m.width >= sideBySideMin {
		m.side = true
		m.treeW = m.width * 6 / 10
		m.detailW = m.width - m.treeW - 1 // one column separator
		m.treeH, m.detailH = contentH, contentH
	} else {
		m.side = false
		m.treeW, m.detailW = m.width, m.width
		m.treeH = max(contentH*6/10, 1)
		m.detailH = max(contentH-m.treeH-1, 1) // one row rule
	}
	m.detailW = max(m.detailW, 1)
	m.tree.Width = m.treeW
	m.tree.Height = m.treeH
	m.renderTree()
	m.ensureVisible()
}

// ensureVisible scrolls the tree viewport just enough to keep the cursor line
// within the visible window.
func (m *model) ensureVisible() {
	switch {
	case m.cursor < m.tree.YOffset:
		m.tree.SetYOffset(m.cursor)
	case m.cursor >= m.tree.YOffset+m.tree.Height:
		m.tree.SetYOffset(m.cursor - m.tree.Height + 1)
	}
}

// renderTree draws every visible line to the tree width and loads the result
// into the viewport, highlighting the cursor line.
func (m *model) renderTree() {
	rows := make([]string, len(m.lines))
	for i, l := range m.lines {
		rows[i] = m.renderLine(l, m.treeW, i == m.cursor)
	}
	m.tree.SetContent(strings.Join(rows, "\n"))
}

// renderLine formats one tree row: indentation, an expand/collapse indicator for
// branches, a marker distinguishing an NPC line (with the speaker's name) from a
// player reply, and the node's truncated text. A jump line uses the ↩ marker and
// its edge label as caption.
func (m model) renderLine(l line, width int, selected bool) string {
	n := l.item.node

	var body string
	switch {
	case l.item.jump:
		caption := ""
		if l.item.edge != nil {
			caption = flatten(l.item.edge.Label)
		}
		if caption != "" {
			body = "↩ " + caption + " → " + flatten(n.Text)
		} else {
			body = "↩ → " + flatten(n.Text)
		}
	case n.Kind == dcanvas.KindReply:
		body = "▷ " + flatten(n.Text)
	default: // line / narrator / plain text node
		if n.Character != nil && n.Character.Name != "" {
			body = "● " + n.Character.Name + ": " + flatten(n.Text)
		} else {
			body = "● " + flatten(n.Text)
		}
	}

	var indicator string
	switch {
	case l.item.jump:
		indicator = ""
	case l.expandable && l.expanded:
		indicator = "▾ "
	case l.expandable && !l.expanded:
		indicator = "▸ "
	default:
		indicator = "  " // childless leaf
	}

	plain := strings.Repeat("  ", l.item.depth) + indicator + body
	plain = runewidth.Truncate(plain, width, "…")

	if selected {
		return selectedStyle.Width(width).Render(plain)
	}
	switch {
	case l.item.jump:
		return jumpStyle.Render(plain)
	case n.Kind == dcanvas.KindReply:
		return replyStyle.Render(plain)
	default:
		return lineStyle.Render(plain)
	}
}

// renderDetail draws the detail pane for the selected node: its full
// (untruncated) text, speaker, dialogue x- fields, and the label of the edge
// that led to it.
func (m model) renderDetail(width, height int) string {
	if len(m.lines) == 0 {
		return detailStyle.Width(width).Height(height).MaxHeight(height).Render("")
	}
	it := m.lines[m.cursor].item
	n := it.node

	var b strings.Builder
	b.WriteString(detailLabel.Render("Текст") + "\n" + n.Text + "\n")

	if n.Character != nil && n.Character.Name != "" {
		b.WriteString("\n" + detailLabel.Render("Мовець") + ": " + n.Character.Name + "\n")
		if n.Character.Portrait != "" {
			b.WriteString(detailLabel.Render("Портрет") + ": " + n.Character.Portrait + "\n")
		}
		if n.Character.Gender != "" {
			b.WriteString(detailLabel.Render("Стать") + ": " + n.Character.Gender + "\n")
		}
	}

	field := func(k, v string) {
		if v != "" {
			b.WriteString(detailLabel.Render(k) + ": " + v + "\n")
		}
	}
	b.WriteString("\n")
	field("d-kind", n.Kind)
	field("d-role", n.Role)
	field("d-condition", n.Condition)
	field("d-action", n.Action)
	field("d-sound", n.Sound)

	if it.edge != nil && it.edge.Label != "" {
		b.WriteString("\n" + detailLabel.Render("Мітка переходу") + ": " + it.edge.Label + "\n")
	}

	return detailStyle.Width(width).Height(height).MaxHeight(height).Render(b.String())
}

// statusBar draws the Turbo-Vision-style bottom strip: a solid background with
// each hotkey highlighted and its Ukrainian label beside it. It is
// context-sensitive, showing only the action that applies to the selected line,
// and never Enter next to the arrow that does the same thing.
func (m model) statusBar(width int) string {
	seg := func(key, label string) string { return barKey.Render(key) + barStyle.Render(" "+label) }

	segs := []string{seg("↑↓", "Рух")}
	if len(m.lines) > 0 {
		l := m.lines[m.cursor]
		switch {
		case l.item.jump:
			segs = append(segs, seg("⏎", "Стрибнути"))
		case l.expandable && !l.expanded:
			segs = append(segs, seg("→", "Розгорнути"))
		case l.expandable && l.expanded:
			segs = append(segs, seg("←", "Згорнути"))
		}
	}
	segs = append(segs, seg("Home", "Початок"), seg("End", "Кінець"))
	if len(m.jumpStack) > 0 {
		segs = append(segs, seg("⌫", "Назад"))
	}
	segs = append(segs, seg("Esc", "Вихід"))

	content := strings.Join(segs, barStyle.Render("   "))
	if pad := width - lipgloss.Width(content); pad > 0 {
		content += barStyle.Render(strings.Repeat(" ", pad))
	}
	return content
}

func (m model) View() string {
	if !m.ready || m.width == 0 || m.height == 0 {
		return "Завантаження…"
	}
	var content string
	if m.side {
		sep := sepStyle.Render(strings.TrimSuffix(strings.Repeat("│\n", m.treeH), "\n"))
		content = lipgloss.JoinHorizontal(lipgloss.Top, m.tree.View(), sep, m.renderDetail(m.detailW, m.detailH))
	} else {
		rule := ruleStyle.Render(strings.Repeat("─", m.width))
		content = lipgloss.JoinVertical(lipgloss.Left, m.tree.View(), rule, m.renderDetail(m.detailW, m.detailH))
	}
	return lipgloss.JoinVertical(lipgloss.Left, content, m.statusBar(m.width))
}

// flatten collapses a node's text to a single line for the tree, so multi-line
// utterances stay on one row (the detail pane shows the full text).
func flatten(s string) string {
	return strings.TrimSpace(strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(s))
}
