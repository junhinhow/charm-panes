package panes

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Example mostra um editor com uma barra lateral fixa, a área de edição e um
// terminal.
//
// A barra lateral não cede espaço quando outro painel é redimensionado, mas
// continua podendo ser redimensionada explicitamente.
func Example() {
	l := New(Horizontal)
	l.SetSize(60, 8)

	l.Add(Pane{
		ID:      "sidebar",
		Title:   "Arquivos",
		Content: "main.go\nREADME.md",
		MinSize: 12,
		Fixed:   true,
	})
	l.Add(Pane{
		ID:      "editor",
		Title:   "Editor",
		Content: "package main\n\nfunc main() {}",
	})
	l.Add(Pane{
		ID:      "terminal",
		Title:   "Terminal",
		Content: "$ go test ./...",
	})

	l.Focus(1)
	l.Resize(10)

	view := l.View()
	fmt.Println(lipgloss.Width(view), lipgloss.Height(view))
	fmt.Println(l.Sizes())

	// A tecla alt+right aumenta o painel focado em uma coluna.
	l.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight, Mod: tea.ModAlt}))
	fmt.Println(l.Focused(), l.Sizes()[1])

	// Output:
	// 60 8
	// [20 29 9]
	// 1 30
}
