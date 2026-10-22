package panes

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// tecla converte uma tecla do Bubble Tea em uma mensagem de pressionamento.
func tecla(k tea.Key) tea.KeyPressMsg {
	return tea.KeyPressMsg(k)
}

func TestUpdateNavegacao(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)
	l.Focus(0)

	if cmd := l.Update(tecla(tea.Key{Code: tea.KeyTab})); cmd != nil {
		t.Error("Update devolveu um comando inesperado")
	}
	if foco := l.Focused(); foco != 1 {
		t.Errorf("foco após tab = %d, esperado 1", foco)
	}

	l.Update(tecla(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	if foco := l.Focused(); foco != 0 {
		t.Errorf("foco após shift+tab = %d, esperado 0", foco)
	}

	l.Update(tecla(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	if foco := l.Focused(); foco != 2 {
		t.Errorf("foco após shift+tab circular = %d, esperado 2", foco)
	}
}

func TestUpdateResizeHorizontal(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)
	l.Focus(0)

	casos := []struct {
		nome  string
		tecla tea.Key
		delta int
	}{
		{"alt+right", tea.Key{Code: tea.KeyRight, Mod: tea.ModAlt}, 1},
		{"alt+left", tea.Key{Code: tea.KeyLeft, Mod: tea.ModAlt}, -1},
		{"+", tea.Key{Text: "+", Code: '+'}, 1},
		{"-", tea.Key{Text: "-", Code: '-'}, -1},
		{"=", tea.Key{Text: "=", Code: '='}, 1},
		{"_", tea.Key{Text: "_", Code: '_'}, -1},
	}

	for _, caso := range casos {
		antes := l.Sizes()
		l.Update(tecla(caso.tecla))
		depois := l.Sizes()
		if delta := depois[0] - antes[0]; delta != caso.delta {
			t.Errorf("%s: delta = %d, esperado %d", caso.nome, delta, caso.delta)
		}
		if soma := depois[0] + depois[1] + depois[2]; soma+2 != 60 {
			t.Errorf("%s: soma = %d, esperado 58", caso.nome, soma)
		}
	}
}

func TestUpdateResizeVertical(t *testing.T) {
	l := novoLayout(Vertical, 40, 60,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)
	l.Focus(0)
	antes := l.Sizes()

	// Setas horizontais não redimensionam um layout vertical.
	l.Update(tecla(tea.Key{Code: tea.KeyRight, Mod: tea.ModAlt}))
	if depois := l.Sizes(); depois[0] != antes[0] {
		t.Errorf("alt+right alterou layout vertical: %d -> %d", antes[0], depois[0])
	}

	l.Update(tecla(tea.Key{Code: tea.KeyDown, Mod: tea.ModAlt}))
	if depois := l.Sizes(); depois[0] != antes[0]+1 {
		t.Errorf("alt+down = %d, esperado %d", depois[0], antes[0]+1)
	}

	l.Update(tecla(tea.Key{Code: tea.KeyUp, Mod: tea.ModAlt}))
	if depois := l.Sizes(); depois[0] != antes[0] {
		t.Errorf("alt+up = %d, esperado %d", depois[0], antes[0])
	}
}

func TestUpdateWindowSize(t *testing.T) {
	l := New(Horizontal)
	l.Add(Pane{ID: "a", Content: "um"})
	l.Add(Pane{ID: "b", Content: "dois"})

	l.Update(tea.WindowSizeMsg{Width: 33, Height: 7})

	if largura, altura := l.Size(); largura != 33 || altura != 7 {
		t.Errorf("Size = (%d, %d), esperado (33, 7)", largura, altura)
	}
	view := l.View()
	linhas := strings.Split(view, "\n")
	if len(linhas) != 7 {
		t.Errorf("altura da View = %d, esperado 7", len(linhas))
	}
	for i, linha := range linhas {
		if largura := lipgloss.Width(linha); largura != 33 {
			t.Errorf("linha %d tem largura %d, esperado 33", i, largura)
		}
	}
}

func TestUpdateTeclaDesconhecida(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)
	l.Focus(1)
	antes := l.Sizes()

	l.Update(tecla(tea.Key{Text: "x", Code: 'x'}))

	if foco := l.Focused(); foco != 1 {
		t.Errorf("foco = %d, esperado 1", foco)
	}
	if depois := l.Sizes(); depois[1] != antes[1] {
		t.Errorf("tamanho mudou com tecla desconhecida: %v -> %v", antes, depois)
	}
}

func TestInitEViewModel(t *testing.T) {
	l := novoLayout(Horizontal, 40, 6,
		Pane{ID: "a", Title: "A", Content: "um"},
		Pane{ID: "b", Title: "B", Content: "dois"},
	)

	if cmd := l.Init(); cmd != nil {
		t.Error("Init deveria devolver nil")
	}

	viewModel := l.ViewModel()
	if viewModel.Content != l.View() {
		t.Error("ViewModel não corresponde a View")
	}
}
