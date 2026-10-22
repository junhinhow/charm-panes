package panes

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// conferirGeometria garante que todas as linhas têm a largura informada e que
// a quantidade de linhas é exatamente a altura.
func conferirGeometria(t *testing.T, view string, largura, altura int) {
	t.Helper()
	if view == "" {
		return
	}

	linhas := strings.Split(view, "\n")
	if len(linhas) != altura {
		t.Fatalf("altura = %d, esperado %d", len(linhas), altura)
	}
	for i, linha := range linhas {
		if larguraLinha := lipgloss.Width(linha); larguraLinha != largura {
			t.Errorf("linha %d tem largura %d, esperado %d: %q", i, larguraLinha, largura, linha)
		}
	}
}

// colunaDe devolve a coluna de exibição onde texto começa na linha.
func colunaDe(linha, texto string) int {
	idx := strings.Index(linha, texto)
	if idx < 0 {
		return -1
	}
	return ansi.StringWidth(linha[:idx])
}

func TestViewVazio(t *testing.T) {
	l := New(Horizontal)
	l.SetSize(40, 10)
	if view := l.View(); view != "" {
		t.Errorf("View de layout vazio = %q, esperado vazio", view)
	}
}

func TestViewUmPainel(t *testing.T) {
	l := novoLayout(Horizontal, 20, 3,
		Pane{ID: "solo", Title: "Título", Content: "olá"},
	)

	view := l.View()
	conferirGeometria(t, view, 20, 3)
	if !strings.HasPrefix(view, "olá") {
		t.Errorf("View não começa com o conteúdo direto: %q", view)
	}
	if strings.Contains(view, "│") {
		t.Error("View com um painel não deveria ter divisores")
	}
	if strings.Contains(view, "▌") {
		t.Error("View com um painel não deveria ter marcador de foco")
	}
}

func TestViewDoisPainéis(t *testing.T) {
	l := novoLayout(Horizontal, 20, 4,
		Pane{ID: "a", Title: "Lado", Content: "esq"},
		Pane{ID: "b", Title: "Editor", Content: "dir"},
	)

	view := l.View()
	conferirGeometria(t, view, 20, 4)

	tamanhos := l.Sizes()
	linha0 := strings.Split(view, "\n")[0]
	if coluna := colunaDe(linha0, "│"); coluna != tamanhos[0] {
		t.Errorf("divisor na coluna %d, esperado %d", coluna, tamanhos[0])
	}
	if !strings.HasPrefix(linha0, "▌ Lado") {
		t.Errorf("título do painel focado = %q", linha0)
	}
	if !strings.Contains(linha0, " Editor") {
		t.Errorf("título do segundo painel ausente: %q", linha0)
	}
	if strings.Count(linha0, "│") != 1 {
		t.Errorf("quantidade de divisores = %d, esperado 1", strings.Count(linha0, "│"))
	}

	l.FocusNext()
	view = l.View()
	linha0 = strings.Split(view, "\n")[0]
	if strings.Count(linha0, "▌") != 1 {
		t.Errorf("marcador de foco deveria ser único: %q", linha0)
	}
	if !strings.Contains(linha0, "▌ Editor") {
		t.Errorf("foco não se moveu para o segundo painel: %q", linha0)
	}
}

func TestViewVertical(t *testing.T) {
	l := novoLayout(Vertical, 20, 10,
		Pane{ID: "a", Title: "Topo", Content: "linha"},
		Pane{ID: "b", Title: "Base", Content: "fundo"},
	)

	view := l.View()
	conferirGeometria(t, view, 20, 10)

	tamanhos := l.Sizes()
	if tamanhos[0] != 5 || tamanhos[1] != 4 {
		t.Fatalf("tamanhos = %v, esperado [5 4]", tamanhos)
	}

	linhas := strings.Split(view, "\n")
	divisor := strings.Repeat("─", 20)
	if linhas[tamanhos[0]] != divisor {
		t.Errorf("linha divisória = %q, esperado %q", linhas[tamanhos[0]], divisor)
	}
	if strings.Count(view, divisor) != 1 {
		t.Errorf("quantidade de divisores horizontais = %d, esperado 1", strings.Count(view, divisor))
	}
	if !strings.HasPrefix(linhas[0], "▌ Topo") {
		t.Errorf("título do painel focado = %q", linhas[0])
	}
	if !strings.HasPrefix(linhas[tamanhos[0]+1], "  Base") {
		t.Errorf("título do segundo painel = %q", linhas[tamanhos[0]+1])
	}
}

func TestViewTituloMultibyte(t *testing.T) {
	l := novoLayout(Horizontal, 20, 4,
		Pane{ID: "a", Title: "Ação", Content: "conteúdo"},
		Pane{ID: "b", Title: "Café", Content: "texto"},
	)

	view := l.View()
	conferirGeometria(t, view, 20, 4)
	if !strings.Contains(view, "Ação") {
		t.Errorf("título multibyte truncado indevidamente:\n%s", view)
	}
	if !utf8.ValidString(view) {
		t.Error("View produziu UTF-8 inválido")
	}

	// Em um painel estreito o título é truncado por célula, não por byte.
	l.SetSize(9, 4)
	view = l.View()
	conferirGeometria(t, view, 9, 4)
	if !utf8.ValidString(view) {
		t.Error("View estreita produziu UTF-8 inválido")
	}
	if strings.ContainsRune(view, utf8.RuneError) {
		t.Errorf("View estreita contém caractere de substituição:\n%q", view)
	}
}

func TestViewConteudoMultibyte(t *testing.T) {
	l := novoLayout(Horizontal, 11, 4,
		Pane{ID: "a", Content: "日本語日本語日本語"},
		Pane{ID: "b", Content: "ação"},
	)

	view := l.View()
	conferirGeometria(t, view, 11, 4)
	if !utf8.ValidString(view) {
		t.Error("View produziu UTF-8 inválido")
	}
}

func TestViewConteudoLongo(t *testing.T) {
	conteudo := strings.Repeat("palavra ", 100) + "\n" + strings.Repeat("x", 300)
	l := novoLayout(Horizontal, 13, 5,
		Pane{ID: "a", Content: conteudo},
		Pane{ID: "b", Content: conteudo},
	)

	view := l.View()
	conferirGeometria(t, view, 13, 5)
}

func TestViewLarguraZero(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8, Pane{ID: "a", Content: "x"}, Pane{ID: "b", Content: "y"})

	l.SetSize(0, 8)
	if view := l.View(); view != "" {
		t.Errorf("View com largura 0 = %q, esperado vazio", view)
	}

	l.SetSize(40, 0)
	if view := l.View(); view != "" {
		t.Errorf("View com altura 0 = %q, esperado vazio", view)
	}

	l.SetSize(0, 0)
	if view := l.View(); view != "" {
		t.Errorf("View com tamanho 0x0 = %q, esperado vazio", view)
	}
}

func TestViewTamanhosVariados(t *testing.T) {
	for largura := 0; largura <= 12; largura++ {
		for altura := 0; altura <= 8; altura++ {
			for quantidade := 0; quantidade <= 5; quantidade++ {
				l := New(Horizontal)
				for i := 0; i < quantidade; i++ {
					l.Add(Pane{ID: fmt.Sprintf("p%d", i), Title: "Ação"})
				}
				l.SetSize(largura, altura)

				view := l.View()
				if view == "" {
					continue
				}
				t.Run(fmt.Sprintf("%dx%d_%dpainéis", largura, altura, quantidade), func(t *testing.T) {
					conferirGeometria(t, view, largura, altura)
				})
			}
		}
	}
}

func TestViewNaoExcedeVertical(t *testing.T) {
	for largura := 1; largura <= 10; largura++ {
		for altura := 0; altura <= 12; altura++ {
			l := New(Vertical)
			for i := 0; i < 4; i++ {
				l.Add(Pane{ID: fmt.Sprintf("p%d", i), Title: "Ação", Content: "conteúdo"})
			}
			l.SetSize(largura, altura)
			if view := l.View(); view != "" {
				conferirGeometria(t, view, largura, altura)
			}
		}
	}
}

func TestViewAbaixoDoMinimo(t *testing.T) {
	l := novoLayout(Horizontal, 40, 8, Pane{ID: "solo", Content: "x"})

	// A largura 2 não comporta o mínimo de 3 colunas.
	l.SetSize(2, 8)
	if view := l.View(); view != "" {
		t.Errorf("View abaixo do mínimo = %q, esperado vazio", view)
	}

	l.SetSize(3, 8)
	if view := l.View(); view == "" {
		t.Error("View com exatamente o mínimo não deveria ser vazia")
	}
}

func TestViewDeterministico(t *testing.T) {
	l := novoLayout(Horizontal, 42, 7,
		Pane{ID: "a", Title: "Ação", Content: "primeira linha\nsegunda linha"},
		Pane{ID: "b", Title: "Editor", Content: "código"},
		Pane{ID: "c", Title: "Saída", Content: "日本語"},
	)

	primeira := l.View()
	for i := 0; i < 100; i++ {
		if view := l.View(); view != primeira {
			t.Fatalf("renderização %d diferente da primeira", i+1)
		}
	}
}

func TestViewIndicadorFocoUnico(t *testing.T) {
	l := novoLayout(Horizontal, 40, 6,
		Pane{ID: "a", Title: "um"},
		Pane{ID: "b", Title: "dois"},
		Pane{ID: "c", Title: "três"},
	)

	for i := 0; i < 3; i++ {
		l.Focus(i)
		view := l.View()
		if quantidade := strings.Count(view, "▌"); quantidade != 1 {
			t.Errorf("foco %d: marcadores = %d, esperado 1", i, quantidade)
		}
	}
}

func TestViewDivisoresExatos(t *testing.T) {
	l := novoLayout(Horizontal, 31, 6,
		Pane{ID: "a"}, Pane{ID: "b"}, Pane{ID: "c"}, Pane{ID: "d"},
	)

	tamanhos := l.Sizes()
	view := l.View()
	for numero, linha := range strings.Split(view, "\n") {
		partes := strings.Split(linha, "│")
		if len(partes) != 4 {
			t.Fatalf("linha %d tem %d células, esperado 4", numero, len(partes))
		}
		for i, parte := range partes {
			if largura := lipgloss.Width(parte); largura != tamanhos[i] {
				t.Errorf("linha %d, célula %d tem largura %d, esperado %d", numero, i, largura, tamanhos[i])
			}
		}
	}
}
