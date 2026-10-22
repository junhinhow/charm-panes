package panes

import (
	"fmt"
	"strings"
	"testing"
)

// novoLayout cria um layout com o tamanho já definido e adiciona os painéis.
func novoLayout(dir Direction, largura, altura int, paineis ...Pane) *Layout {
	l := New(dir)
	l.SetSize(largura, altura)
	for _, p := range paineis {
		l.Add(p)
	}
	return l
}

// somar devolve a soma dos tamanhos dos painéis visíveis.
func somar(tamanhos []int, visiveis int) int {
	soma := 0
	for i := 0; i < visiveis && i < len(tamanhos); i++ {
		soma += tamanhos[i]
	}
	return soma
}

func TestSomaExataHorizontal(t *testing.T) {
	larguras := []int{80, 81, 40, 7}
	quantidades := []int{2, 3, 5}

	for _, largura := range larguras {
		for _, quantidade := range quantidades {
			t.Run(fmt.Sprintf("%dpainéis_%dcol", quantidade, largura), func(t *testing.T) {
				l := New(Horizontal)
				for i := 0; i < quantidade; i++ {
					l.Add(Pane{ID: fmt.Sprintf("p%d", i)})
				}
				l.SetSize(largura, 10)

				tamanhos := l.Sizes()
				if len(tamanhos) != quantidade {
					t.Fatalf("Sizes devolveu %d tamanhos, esperado %d", len(tamanhos), quantidade)
				}

				visiveis := l.Visible()
				soma := somar(tamanhos, visiveis)
				if visiveis > 0 && soma+visiveis-1 != largura {
					t.Errorf("soma %d + %d divisores = %d, esperado %d", soma, visiveis-1, soma+visiveis-1, largura)
				}
				for i := 0; i < visiveis; i++ {
					if tamanhos[i] < 3 {
						t.Errorf("painel %d tem tamanho %d, mínimo 3", i, tamanhos[i])
					}
				}
				for i := visiveis; i < quantidade; i++ {
					if tamanhos[i] != 0 {
						t.Errorf("painel oculto %d tem tamanho %d, esperado 0", i, tamanhos[i])
					}
				}
			})
		}
	}
}

func TestSomaExataVertical(t *testing.T) {
	alturas := []int{80, 81, 40, 7}

	for _, altura := range alturas {
		for _, quantidade := range []int{2, 3, 5} {
			t.Run(fmt.Sprintf("%dpainéis_%dlin", quantidade, altura), func(t *testing.T) {
				l := New(Vertical)
				for i := 0; i < quantidade; i++ {
					l.Add(Pane{ID: fmt.Sprintf("p%d", i)})
				}
				l.SetSize(40, altura)

				tamanhos := l.Sizes()
				visiveis := l.Visible()
				soma := somar(tamanhos, visiveis)
				if visiveis > 0 && soma+visiveis-1 != altura {
					t.Errorf("soma %d + %d divisores = %d, esperado %d", soma, visiveis-1, soma+visiveis-1, altura)
				}
			})
		}
	}
}

func TestMinSizeRespeitado(t *testing.T) {
	l := novoLayout(Horizontal, 30, 8,
		Pane{ID: "a", MinSize: 10},
		Pane{ID: "b", MinSize: 5},
	)

	tamanhos := l.Sizes()
	if l.Visible() != 2 {
		t.Fatalf("Visible = %d, esperado 2", l.Visible())
	}
	if tamanhos[0] < 10 {
		t.Errorf("painel a = %d, mínimo 10", tamanhos[0])
	}
	if tamanhos[1] < 5 {
		t.Errorf("painel b = %d, mínimo 5", tamanhos[1])
	}
	if soma := tamanhos[0] + tamanhos[1]; soma+1 != 30 {
		t.Errorf("soma = %d, esperado 29", soma)
	}

	// Espaço insuficiente para os dois: o segundo é ocultado.
	l.SetSize(12, 8)
	if visiveis := l.Visible(); visiveis != 1 {
		t.Fatalf("Visible = %d, esperado 1", visiveis)
	}
	tamanhos = l.Sizes()
	if tamanhos[0] < 10 {
		t.Errorf("painel a = %d, mínimo 10", tamanhos[0])
	}
	if tamanhos[1] != 0 {
		t.Errorf("painel b = %d, esperado 0 (oculto)", tamanhos[1])
	}
}

func TestMinSizePadrao(t *testing.T) {
	l := novoLayout(Horizontal, 80, 10, Pane{ID: "a", MinSize: -5})
	if tamanhos := l.Sizes(); tamanhos[0] != 80 {
		t.Errorf("tamanho = %d, esperado 80", tamanhos[0])
	}

	p, _ := l.Pane("a")
	if p.MinSize != 3 {
		t.Errorf("MinSize normalizado = %d, esperado 3", p.MinSize)
	}
}

func TestResizeParaCimaEParaBaixo(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)

	iniciais := l.Sizes()
	if iniciais[0] != 20 || iniciais[1] != 19 || iniciais[2] != 19 {
		t.Fatalf("tamanhos iniciais = %v, esperado [20 19 19]", iniciais)
	}

	l.Focus(0)
	l.Resize(10)
	tamanhos := l.Sizes()
	if tamanhos[0] != 30 {
		t.Errorf("painel focado = %d, esperado 30", tamanhos[0])
	}
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 60 {
		t.Errorf("soma = %d, esperado 58", soma)
	}

	// Reduzir além do mínimo aplica apenas o possível.
	l.Resize(-100)
	tamanhos = l.Sizes()
	if tamanhos[0] != 3 {
		t.Errorf("painel focado = %d, esperado 3", tamanhos[0])
	}
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 60 {
		t.Errorf("soma = %d, esperado 58", soma)
	}

	// Crescer além do possível não estoura o total.
	l.Focus(1)
	l.Resize(1000)
	tamanhos = l.Sizes()
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 60 {
		t.Errorf("soma após crescimento máximo = %d, esperado 58", soma)
	}
	if tamanhos[0] < 3 || tamanhos[2] < 3 {
		t.Errorf("painéis doadores abaixo do mínimo: %v", tamanhos)
	}
}

func TestResizeTo(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)
	l.Focus(0)

	atual := l.Sizes()[0]
	l.ResizeTo(atual)
	if tamanhos := l.Sizes(); tamanhos[0] != atual {
		t.Errorf("ResizeTo(tamanho atual) = %d, esperado %d", tamanhos[0], atual)
	}

	l.ResizeTo(30)
	if tamanhos := l.Sizes(); tamanhos[0] != 30 {
		t.Errorf("ResizeTo(30) = %d, esperado 30", tamanhos[0])
	}

	l.ResizeTo(0)
	if tamanhos := l.Sizes(); tamanhos[0] != 3 {
		t.Errorf("ResizeTo(0) = %d, esperado 3", tamanhos[0])
	}

	l.ResizeTo(9999)
	tamanhos := l.Sizes()
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 60 {
		t.Errorf("soma = %d, esperado 58", soma)
	}
	if tamanhos[1] < 3 || tamanhos[2] < 3 {
		t.Errorf("doadores abaixo do mínimo: %v", tamanhos)
	}
}

func TestFixedNaoCede(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "fixo", Fixed: true},
		Pane{ID: "b"},
		Pane{ID: "c"},
	)

	iniciais := l.Sizes()
	if iniciais[0] != 20 {
		t.Fatalf("painel fixo = %d, esperado 20", iniciais[0])
	}

	l.Focus(1)
	l.Resize(5)
	tamanhos := l.Sizes()
	if tamanhos[0] != iniciais[0] {
		t.Errorf("painel fixo cedeu espaço: %d -> %d", iniciais[0], tamanhos[0])
	}
	if tamanhos[1] != 24 {
		t.Errorf("painel b = %d, esperado 24", tamanhos[1])
	}

	// Ao crescer o container, o fixo mantém o tamanho e os flexíveis absorvem.
	l.SetSize(80, 10)
	tamanhos = l.Sizes()
	if tamanhos[0] != iniciais[0] {
		t.Errorf("painel fixo mudou no SetSize: %d -> %d", iniciais[0], tamanhos[0])
	}
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 80 {
		t.Errorf("soma = %d, esperado 78", soma)
	}

	// O fixo ainda pode ser alvo explícito de redimensionamento.
	l.Focus(0)
	l.ResizeTo(30)
	if tamanhos := l.Sizes(); tamanhos[0] != 30 {
		t.Errorf("painel fixo após ResizeTo = %d, esperado 30", tamanhos[0])
	}

	l.ResizeTo(10)
	tamanhos = l.Sizes()
	if tamanhos[0] != 10 {
		t.Errorf("painel fixo após encolher = %d, esperado 10", tamanhos[0])
	}
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 80 {
		t.Errorf("soma após encolher o fixo = %d, esperado 78", soma)
	}
}

func TestFixedTodosMantemSoma(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8,
		Pane{ID: "a", Fixed: true},
		Pane{ID: "b", Fixed: true},
	)
	l.Focus(0)

	l.Resize(-5)
	tamanhos := l.Sizes()
	if tamanhos[0] != 15 {
		t.Errorf("painel a = %d, esperado 15", tamanhos[0])
	}
	if soma := tamanhos[0] + tamanhos[1]; soma+1 != 40 {
		t.Errorf("soma = %d, esperado 39", soma)
	}

	// Crescer não é possível: não há doador flexível.
	l.Resize(5)
	if tamanhos := l.Sizes(); tamanhos[0] != 15 {
		t.Errorf("crescimento sem doador alterou o fixo: %d", tamanhos[0])
	}
}

func TestFocoCircularEClampado(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)

	l.Focus(999)
	if foco := l.Focused(); foco != 2 {
		t.Errorf("Focus(999) = %d, esperado 2", foco)
	}
	l.Focus(-7)
	if foco := l.Focused(); foco != 0 {
		t.Errorf("Focus(-7) = %d, esperado 0", foco)
	}

	l.FocusNext()
	l.FocusNext()
	l.FocusNext()
	if foco := l.Focused(); foco != 0 {
		t.Errorf("FocusNext circular = %d, esperado 0", foco)
	}
	l.FocusPrev()
	if foco := l.Focused(); foco != 2 {
		t.Errorf("FocusPrev circular = %d, esperado 2", foco)
	}

	vazio := New(Horizontal)
	if foco := vazio.Focused(); foco != -1 {
		t.Errorf("foco em layout vazio = %d, esperado -1", foco)
	}
	vazio.FocusNext()
	vazio.FocusPrev()
	vazio.Focus(10)
	if foco := vazio.Focused(); foco != -1 {
		t.Errorf("foco em layout vazio após navegação = %d, esperado -1", foco)
	}
}

func TestAddRemoveEPane(t *testing.T) {
	l := New(Horizontal)
	l.SetSize(40, 8)
	l.Add(Pane{ID: "a", Title: "A", Content: "um"})
	l.Add(Pane{ID: "b", Title: "B", Content: "dois"})

	if ids := l.IDs(); strings.Join(ids, ",") != "a,b" {
		t.Fatalf("IDs = %v, esperado [a b]", ids)
	}

	l.SetContent("a", "novo")
	l.SetTitle("a", "Novo")
	p, ok := l.Pane("a")
	if !ok || p.Content != "novo" || p.Title != "Novo" {
		t.Errorf("Pane(a) = %+v, %v", p, ok)
	}

	if _, ok := l.Pane("inexistente"); ok {
		t.Error("Pane(inexistente) devolveu true")
	}
	l.SetContent("inexistente", "x")
	l.SetTitle("inexistente", "x")

	if !l.Remove("a") {
		t.Error("Remove(a) devolveu false")
	}
	if l.Remove("a") {
		t.Error("Remove(a) repetido devolveu true")
	}
	if ids := l.IDs(); strings.Join(ids, ",") != "b" {
		t.Errorf("IDs após remoção = %v, esperado [b]", ids)
	}

	l.Add(Pane{ID: "c"})
	l.IDrop("c")
	if ids := l.IDs(); strings.Join(ids, ",") != "b" {
		t.Errorf("IDs após IDrop = %v, esperado [b]", ids)
	}
}

func TestAddDuplicadoSubstitui(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8,
		Pane{ID: "a", Content: "antigo"},
		Pane{ID: "b", Content: "b"},
	)
	l.Focus(1)

	l.Add(Pane{ID: "a", Title: "novo", Content: "substituto"})

	if ids := l.IDs(); strings.Join(ids, ",") != "a,b" {
		t.Errorf("IDs = %v, esperado [a b]", ids)
	}
	p, _ := l.Pane("a")
	if p.Content != "substituto" || p.Title != "novo" {
		t.Errorf("painel substituído = %+v", p)
	}
	if foco := l.Focused(); foco != 1 {
		t.Errorf("foco = %d, esperado 1", foco)
	}
	if tamanhos := l.Sizes(); len(tamanhos) != 2 {
		t.Errorf("Sizes = %v, esperado 2 tamanhos", tamanhos)
	}
}

func TestRemoveAjustaFoco(t *testing.T) {
	l := novoLayout(Horizontal, 60, 10,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"},
	)

	l.Focus(2)
	l.Remove("c")
	if foco := l.Focused(); foco != 1 {
		t.Errorf("foco após remover o último = %d, esperado 1", foco)
	}

	l.Focus(1)
	l.Remove("a")
	if foco := l.Focused(); foco != 0 {
		t.Errorf("foco após remover anterior = %d, esperado 0", foco)
	}
}

func TestReset(t *testing.T) {
	l := novoLayout(Horizontal, 80, 10,
		Pane{ID: "a", MinSize: 5},
		Pane{ID: "b"},
		Pane{ID: "c"},
	)
	l.Focus(0)
	l.ResizeTo(60)
	l.Reset()

	tamanhos := l.Sizes()
	if tamanhos[0] < 5 {
		t.Errorf("tamanho após Reset = %d, mínimo 5", tamanhos[0])
	}
	if tamanhos[1] != tamanhos[2] {
		t.Errorf("painéis flexíveis desiguais após Reset: %v", tamanhos)
	}
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2]; soma+2 != 80 {
		t.Errorf("soma = %d, esperado 78", soma)
	}

	// Depois do Reset, adicionar volta a dividir igualmente.
	l.Add(Pane{ID: "d"})
	tamanhos = l.Sizes()
	if len(tamanhos) != 4 {
		t.Fatalf("Sizes = %v, esperado 4 tamanhos", tamanhos)
	}
	if soma := tamanhos[0] + tamanhos[1] + tamanhos[2] + tamanhos[3]; soma+3 != 80 {
		t.Errorf("soma = %d, esperado 77", soma)
	}
}

func TestResetComMinimosGrandes(t *testing.T) {
	l := novoLayout(Horizontal, 20, 8,
		Pane{ID: "a", MinSize: 15},
		Pane{ID: "b"},
	)
	l.Reset()

	tamanhos := l.Sizes()
	if tamanhos[0] < 15 {
		t.Errorf("painel a = %d, mínimo 15", tamanhos[0])
	}
	if tamanhos[1] < 3 {
		t.Errorf("painel b = %d, mínimo 3", tamanhos[1])
	}
	if soma := tamanhos[0] + tamanhos[1]; soma+1 != 20 {
		t.Errorf("soma = %d, esperado 19", soma)
	}
}

func TestResetSemEspaco(t *testing.T) {
	l := novoLayout(Horizontal, 20, 8, Pane{ID: "a"}, Pane{ID: "b"})
	l.SetSize(2, 8)
	l.Reset()

	for i, tamanho := range l.Sizes() {
		if tamanho != 0 {
			t.Errorf("painel %d = %d, esperado 0", i, tamanho)
		}
	}
	if l.View() != "" {
		t.Errorf("View = %q, esperado vazio", l.View())
	}
}

func TestDegradacao(t *testing.T) {
	l := New(Horizontal)
	for i := 0; i < 5; i++ {
		l.Add(Pane{ID: fmt.Sprintf("p%d", i)})
	}
	l.SetSize(7, 10)

	if visiveis := l.Visible(); visiveis != 2 {
		t.Fatalf("Visible = %d, esperado 2", visiveis)
	}
	tamanhos := l.Sizes()
	if len(tamanhos) != 5 {
		t.Fatalf("Sizes = %v, esperado 5 entradas", tamanhos)
	}
	if tamanhos[0] != 3 || tamanhos[1] != 3 {
		t.Errorf("tamanhos visíveis = %v, esperado [3 3 ...]", tamanhos)
	}
	for i := 2; i < 5; i++ {
		if tamanhos[i] != 0 {
			t.Errorf("painel oculto %d = %d, esperado 0", i, tamanhos[i])
		}
	}
	if !strings.Contains(l.View(), "+3") {
		t.Errorf("View não indica os 3 painéis ocultos:\n%s", l.View())
	}

	// Ao crescer, os ocultos reaparecem.
	l.SetSize(80, 10)
	if visiveis := l.Visible(); visiveis != 5 {
		t.Errorf("Visible após crescer = %d, esperado 5", visiveis)
	}
	if soma := somar(l.Sizes(), 5); soma+4 != 80 {
		t.Errorf("soma após crescer = %d, esperado 76", soma)
	}
}

func TestSetSizeNegativo(t *testing.T) {
	l := layoutSeguro()
	l.SetSize(-10, -20)
	if largura, altura := l.Size(); largura != 0 || altura != 0 {
		t.Errorf("Size = (%d, %d), esperado (0, 0)", largura, altura)
	}
	if l.View() != "" {
		t.Errorf("View com tamanho zero = %q, esperado vazio", l.View())
	}
}

// layoutSeguro devolve um layout com painéis para testes de valores
// inválidos.
func layoutSeguro() *Layout {
	return novoLayout(Horizontal, 40, 8, Pane{ID: "a"}, Pane{ID: "b"})
}

func TestIDSaoCopia(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8, Pane{ID: "a"}, Pane{ID: "b"})
	ids := l.IDs()
	ids[0] = "alterado"
	if l.IDs()[0] != "a" {
		t.Error("IDs devolveu o slice interno")
	}
}

func TestSizesSaoCopia(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8, Pane{ID: "a"}, Pane{ID: "b"})
	tamanhos := l.Sizes()
	tamanhos[0] = 999
	if l.Sizes()[0] == 999 {
		t.Error("Sizes devolveu o slice interno")
	}
}

func TestResizeSemFocoVisivel(t *testing.T) {
	l := New(Horizontal)
	for i := 0; i < 5; i++ {
		l.Add(Pane{ID: fmt.Sprintf("p%d", i)})
	}
	l.SetSize(7, 10)
	l.Focus(4)

	antes := l.Sizes()
	l.Resize(10)
	depois := l.Sizes()
	if fmt.Sprint(antes) != fmt.Sprint(depois) {
		t.Errorf("Resize em painel oculto alterou tamanhos: %v -> %v", antes, depois)
	}
}

func TestResizePainelUnico(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8, Pane{ID: "solo", Content: "x"})

	l.Resize(-5)
	l.Resize(5)
	l.ResizeTo(10)

	if tamanhos := l.Sizes(); len(tamanhos) != 1 || tamanhos[0] != 40 {
		t.Errorf("tamanhos = %v, esperado [40]", tamanhos)
	}
	if l.View() == "" {
		t.Error("View de painel único não deveria ser vazia")
	}
}

func TestResizeUnicoVisivel(t *testing.T) {
	l := New(Horizontal)
	for i := 0; i < 5; i++ {
		l.Add(Pane{ID: fmt.Sprintf("p%d", i)})
	}
	l.SetSize(7, 10)
	l.Focus(0)

	antes := l.Sizes()
	l.Resize(-1)
	depois := l.Sizes()

	if fmt.Sprint(antes) != fmt.Sprint(depois) {
		t.Errorf("Resize no único painel visível alterou tamanhos: %v -> %v", antes, depois)
	}
	if soma := somar(depois, l.Visible()); soma+l.Visible()-1 != 7 {
		t.Errorf("soma após Resize = %d, esperado 7", soma+l.Visible()-1)
	}
}
