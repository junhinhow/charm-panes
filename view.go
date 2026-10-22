package panes

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// marcadores de foco exibidos no início da linha de título.
const (
	marcadorFocado    = "▌ "
	marcadorSemFoco   = "  "
	divisorVertical   = "│"
	divisorHorizontal = "─"
)

// View renderiza o layout com divisores entre os painéis e um marcador de foco
// no título.
//
// Cada linha tem exatamente a largura configurada e o resultado tem no máximo a
// altura configurada. Linhas mais curtas são completadas com espaços. Um
// layout vazio devolve ""; com um único painel, o conteúdo é renderizado
// diretamente, sem título nem divisores.
func (l *Layout) View() string {
	if len(l.panes) == 0 || l.width <= 0 || l.height <= 0 {
		return ""
	}

	visiveis := l.visibleCount(l.mainSize())
	if visiveis == 0 {
		return ""
	}

	// Um único painel não tem divisores nem título: o conteúdo ocupa tudo.
	if len(l.panes) == 1 {
		return renderConteudo(l.panes[0].Content, l.width, l.height)
	}

	if l.dir == Vertical {
		return l.renderVertical(visiveis)
	}
	return l.renderHorizontal(visiveis)
}

// renderHorizontal compõe os painéis lado a lado com divisores verticais.
func (l *Layout) renderHorizontal(visiveis int) string {
	blocos := make([][]string, visiveis)
	for i := 0; i < visiveis; i++ {
		ocultos := 0
		if i == visiveis-1 {
			ocultos = l.numeroOcultos(visiveis)
		}
		bloco := l.renderPane(i, l.sizes[i], l.height, ocultos)
		blocos[i] = strings.Split(bloco, "\n")
	}

	var b strings.Builder
	for linha := 0; linha < l.height; linha++ {
		for i := 0; i < visiveis; i++ {
			if i > 0 {
				b.WriteString(divisorVertical)
			}
			if linha < len(blocos[i]) {
				b.WriteString(blocos[i][linha])
			}
		}
		if linha < l.height-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// renderVertical compõe os painéis empilhados com divisores horizontais.
func (l *Layout) renderVertical(visiveis int) string {
	var b strings.Builder
	for i := 0; i < visiveis; i++ {
		if i > 0 {
			b.WriteString(strings.Repeat(divisorHorizontal, l.width))
			b.WriteByte('\n')
		}
		ocultos := 0
		if i == visiveis-1 {
			ocultos = l.numeroOcultos(visiveis)
		}
		b.WriteString(l.renderPane(i, l.width, l.sizes[i], ocultos))
		if i < visiveis-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// renderPane renderiza um painel com largura e altura exatas.
//
// A primeira linha é o título com o marcador de foco; o restante é conteúdo.
// Com apenas uma linha de altura, o conteúdo tem prioridade e o título é
// omitido.
func (l *Layout) renderPane(i, largura, altura, ocultos int) string {
	if largura <= 0 || altura <= 0 {
		return ""
	}
	if altura == 1 {
		return renderConteudo(l.panes[i].Content, largura, 1)
	}

	titulo := linhaTitulo(l.panes[i].Title, i == l.focus, largura, ocultos)
	corpo := renderConteudo(l.panes[i].Content, largura, altura-1)
	return titulo + "\n" + corpo
}

// linhaTitulo monta a linha de título com largura exata.
//
// O marcador "▌" indica o painel focado. Quando há painéis ocultos, o
// indicador "+N" é reservado no fim da linha para não ser truncado pelo título.
func linhaTitulo(titulo string, focado bool, largura, ocultos int) string {
	if largura <= 0 {
		return ""
	}

	marcador := marcadorSemFoco
	if focado {
		marcador = marcadorFocado
	}
	sufixo := ""
	if ocultos > 0 {
		sufixo = rotuloOcultos(ocultos)
		if titulo != "" {
			sufixo = " " + sufixo
		}
	}

	larguraMarcador := ansi.StringWidth(marcador)
	if largura <= larguraMarcador {
		return preencher(ansi.Truncate(marcador, largura, ""), largura)
	}
	// O indicador de painéis ocultos tem prioridade sobre o marcador de foco:
	// em painéis muito estreitos, "+N" sozinho ainda comunica a degradação.
	if ansi.StringWidth(sufixo) > 0 && larguraMarcador+ansi.StringWidth(sufixo) > largura {
		return preencher(sufixo, largura)
	}

	disponivel := largura - larguraMarcador - ansi.StringWidth(sufixo)
	if disponivel < 0 {
		disponivel = 0
	}
	textoTitulo := ansi.Truncate(titulo, disponivel, "")
	return preencher(marcador+textoTitulo+sufixo, largura)
}

// renderConteudo renderiza o conteúdo em um bloco de largura e altura exatas.
//
// O texto é quebrado na largura disponível e truncado quando excede a altura.
// Todas as linhas são completadas com espaços.
func renderConteudo(conteudo string, largura, altura int) string {
	if largura <= 0 || altura <= 0 {
		return ""
	}

	estilo := lipgloss.NewStyle().Width(largura).Height(altura)
	bloco := estilo.Render(conteudo)

	linhas := strings.Split(bloco, "\n")
	if len(linhas) > altura {
		linhas = linhas[:altura]
	}
	for len(linhas) < altura {
		linhas = append(linhas, "")
	}
	for i := range linhas {
		linhas[i] = preencher(linhas[i], largura)
	}
	return strings.Join(linhas, "\n")
}

// preencher ajusta a linha para exatamente largura colunas de exibição,
// truncando o excesso e completando com espaços.
func preencher(linha string, largura int) string {
	if largura <= 0 {
		return ""
	}
	if ansi.StringWidth(linha) > largura {
		linha = ansi.Truncate(linha, largura, "")
	}
	if atual := ansi.StringWidth(linha); atual < largura {
		linha += strings.Repeat(" ", largura-atual)
	}
	return linha
}
