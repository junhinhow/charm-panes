# charm-panes

Painéis redimensionáveis e controlados por teclado para interfaces de terminal
feitas com Bubble Tea e Lip Gloss.

`charm-panes` é a contraparte interativa de um layout dividido estático: os
painéis compartilham a largura ou a altura disponível, e o painel focado pode
crescer ou diminuir enquanto os demais se reequilibram sem cruzar o tamanho
mínimo.

## Instalação

```bash
go get github.com/junhinhow/charm-panes
```

## Uso

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

	// Uma barra lateral fixa mantém a largura quando o editor é redimensionado.
	l.Add(panes.Pane{
		ID:      "sidebar",
		Title:   "Arquivos",
		MinSize: 20,
		Fixed:   true,
		Content: "main.go\nREADME.md\ngo.mod",
	})
	l.Add(panes.Pane{
		ID:      "editor",
		Title:   "main.go",
		MinSize: 30,
		Content: "package main\n\nfunc main() {\n\tprintln(\"olá\")\n}",
	})

	if _, err := tea.NewProgram(model{layout: l}).Run(); err != nil {
		panic(err)
	}
}
```

O layout não inicia um programa sozinho. Chame `SetSize` uma vez (ou deixe a
primeira `tea.WindowSizeMsg` fazer isso) e renderize com `View` ou `ViewModel`.

| Tecla | Ação |
| --- | --- |
| `tab` / `shift+tab` | próximo / anterior |
| `alt+esquerda` / `alt+direita` | diminui / aumenta o focado (horizontal) |
| `alt+cima` / `alt+baixo` | diminui / aumenta o focado (vertical) |
| `+` / `=` | aumenta o focado em uma célula |
| `-` / `_` | diminui o focado em uma célula |

## API

| Símbolo | Descrição |
| --- | --- |
| `Direction`, `Horizontal`, `Vertical` | Eixo do layout. |
| `Pane` | ID, título, conteúdo, tamanho mínimo e flag de fixo. |
| `New(dir)` | Cria um layout vazio. |
| `(*Layout) Add(p)` | Adiciona um painel; um ID existente é substituído no lugar. |
| `(*Layout) Remove(id)`, `IDrop(id)` | Remove um painel. |
| `(*Layout) Pane(id)` | Devolve uma cópia do painel. |
| `(*Layout) SetContent(id, s)`, `SetTitle(id, s)` | Atualiza um painel no lugar. |
| `(*Layout) IDs()` | IDs na ordem do layout. |
| `(*Layout) SetSize(w, h)`, `Size()` | Área disponível. |
| `(*Layout) Focus(i)`, `Focused()` | Move e lê o foco (limitado). |
| `(*Layout) FocusNext()`, `FocusPrev()` | Navegação circular de foco. |
| `(*Layout) Resize(delta)`, `ResizeTo(size)` | Redimensiona o painel focado. |
| `(*Layout) Reset()` | Dá a todos os painéis uma fatia igual. |
| `(*Layout) Sizes()`, `Visible()` | Tamanhos atuais e quantidade visível. |
| `(*Layout) View()` | Renderiza divisores, títulos e marcador de foco. |
| `(*Layout) Update(msg)`, `Init()`, `ViewModel()` | Integração com Bubble Tea. |

## Notas de projeto

- **Geometria exata.** O espaço é dividido pelo método do maior resto. Para os
  painéis visíveis, `soma(Sizes()) + divisores` é sempre exatamente a largura
  (`Horizontal`) ou a altura (`Vertical`). `Sizes()` devolve uma entrada por
  painel na ordem do layout; painéis ocultos pela degradação têm tamanho `0`.
- **Somente o eixo principal.** `MinSize` e o redimensionamento valem para a
  largura em layouts `Horizontal` e para a altura em layouts `Vertical`. O eixo
  transversal sempre acompanha `SetSize`.
- **Degradação.** Quando os tamanhos mínimos não cabem, os primeiros painéis
  que cabem continuam visíveis e o título do último visível mostra `+N`. O
  layout nunca fica abaixo de `MinSize` (padrão `3`) silenciosamente; se nem um
  painel couber, `View` devolve `""`.
- **Foco e títulos.** Layouts com mais de um painel reservam a primeira linha de
  cada painel para o título. O focado é marcado com `▌`; não há moldura ao redor
  dos painéis, apenas os divisores `│` / `─`. Um layout de painel único
  renderiza o conteúdo diretamente, sem linha de título.
- **Painéis fixos.** Um painel `Fixed` não cede espaço quando outro é
  redimensionado. Ele ainda pode ser redimensionado explicitamente e encolhe
  proporcionalmente quando o container é pequeno demais para todos os mínimos.
- **Distribuição inicial.** Um layout novo divide o espaço igualmente, ignorando
  `Fixed`. Depois do primeiro redimensionamento manual, painéis novos começam
  com o tamanho médio e as proporções existentes são preservadas.
- **Sem estilo imposto.** A biblioteca não emite cores nem estilos ANSI
  próprios, então compõe com o tema da aplicação. Sequências ANSI no conteúdo
  dos painéis são preservadas e medidas com `lipgloss.Width`.
- **Não é thread-safe.** Como um modelo Bubble Tea, um `Layout` deve ser usado
  em um único laço de eventos.

Limitações: os painéis são planos (sem layouts aninhados), o redimensionamento é
somente por teclado (sem arrastar com o mouse) e a linha de título custa uma
linha de conteúdo em layouts com mais de um painel. O layout começa em `80x24`
até que `SetSize` seja chamado.

## Licença

MIT
