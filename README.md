# charm-panes

Resizable, keyboard-driven panes for terminal interfaces built with Bubble Tea and Lip Gloss.

`charm-panes` is the interactive counterpart of a static split layout: panels
share the available width or height, and the focused panel can grow or shrink
while the others rebalance without ever crossing their minimum size.

## Install

```bash
go get github.com/junhinhow/charm-panes
```

## Usage

```go
package main

import (
	tea "charm.land/bubbletea/v2"
	panes "github.com/junhinhow/charm-panes"
)

type model struct {
	layout *panes.Layout
}

func (m model) Init() tea.Cmd { return m.layout.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.layout.Update(msg)
}

func (m model) View() tea.View { return m.layout.ViewModel() }

func main() {
	l := panes.New(panes.Horizontal)
	l.SetSize(80, 24)

	// A fixed sidebar keeps its width when the editor is resized.
	l.Add(panes.Pane{
		ID:      "sidebar",
		Title:   "Files",
		MinSize: 20,
		Fixed:   true,
		Content: "main.go\nREADME.md\ngo.mod",
	})
	l.Add(panes.Pane{
		ID:      "editor",
		Title:   "main.go",
		MinSize: 30,
		Content: "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}",
	})

	if _, err := tea.NewProgram(model{layout: l}).Run(); err != nil {
		panic(err)
	}
}
```

The layout itself does not start a program. Call `SetSize` once (or let the
first `tea.WindowSizeMsg` do it), then render with `View` or `ViewModel`.

| Key | Action |
| --- | --- |
| `tab` / `shift+tab` | next / previous pane |
| `alt+left` / `alt+right` | shrink / grow the focused pane (horizontal) |
| `alt+up` / `alt+down` | shrink / grow the focused pane (vertical) |
| `+` / `=` | grow the focused pane by one cell |
| `-` / `_` | shrink the focused pane by one cell |

## API

| Symbol | Description |
| --- | --- |
| `Direction`, `Horizontal`, `Vertical` | Layout axis. |
| `Pane` | ID, title, content, minimum size and fixed flag. |
| `New(dir)` | Creates an empty layout. |
| `(*Layout) Add(p)` | Appends a pane; an existing ID is replaced in place. |
| `(*Layout) Remove(id)`, `IDrop(id)` | Removes a pane. |
| `(*Layout) Pane(id)` | Returns a copy of a pane. |
| `(*Layout) SetContent(id, s)`, `SetTitle(id, s)` | Updates a pane in place. |
| `(*Layout) IDs()` | IDs in layout order. |
| `(*Layout) SetSize(w, h)`, `Size()` | Available area. |
| `(*Layout) Focus(i)`, `Focused()` | Moves and reads the focus (clamped). |
| `(*Layout) FocusNext()`, `FocusPrev()` | Circular focus navigation. |
| `(*Layout) Resize(delta)`, `ResizeTo(size)` | Resizes the focused pane. |
| `(*Layout) Reset()` | Gives every pane an equal share. |
| `(*Layout) Sizes()`, `Visible()` | Current pane sizes and visible count. |
| `(*Layout) View()` | Renders dividers, titles and focus marker. |
| `(*Layout) Update(msg)`, `Init()`, `ViewModel()` | Bubble Tea integration. |

## Design notes

- **Exact geometry.** Space is split with the largest-remainder method. For the
  visible panes, `sum(Sizes()) + dividers` is always exactly the width
  (`Horizontal`) or height (`Vertical`). `Sizes()` returns one entry per pane in
  layout order; panes hidden by degradation have size `0`.
- **Main axis only.** `MinSize` and resizing apply to width in `Horizontal`
  layouts and to height in `Vertical` layouts. The cross axis always follows
  `SetSize`.
- **Degradation.** When the minimum sizes do not fit, the first panes that fit
  remain visible and the last visible title shows `+N`. The layout never
  silently drops below `MinSize` (default `3`); if not even one pane fits,
  `View` returns `""`.
- **Focus and titles.** Multi-pane layouts reserve the first line of each pane
  for its title. The focused pane is marked with `▌`; there is no box around
  panes, only the `│` / `─` dividers. A single-pane layout renders its content
  directly, without a title row.
- **Fixed panes.** A `Fixed` pane does not donate space when another pane is
  resized. It can still be resized explicitly, and it shrinks proportionally
  when the container is too small for every minimum.
- **Initial distribution.** A fresh layout divides space equally, ignoring
  `Fixed`. After the first manual resize, new panes start at the average size
  and the existing proportions are preserved.
- **No styling imposed.** The library emits no colors or ANSI styles of its own,
  so it composes with the application theme. ANSI sequences inside pane content
  are preserved and measured with `lipgloss.Width`.
- **Not thread-safe.** Like a Bubble Tea model, a `Layout` is meant to be used
  from a single event loop.

Limitations: panes are flat (no nested layouts), resizing is keyboard-only (no
mouse dragging), and the title row costs one line of content in multi-pane
layouts. Layouts start at `80x24` until `SetSize` is called.

## License

MIT
