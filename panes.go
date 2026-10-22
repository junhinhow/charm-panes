// Package panes implementa painéis divisíveis e redimensionáveis para
// interfaces de terminal.
//
// Um [Layout] organiza painéis lado a lado ([Horizontal]) ou empilhados
// ([Vertical]). Entre painéis vizinhos existe sempre um divisor de uma coluna
// (│) ou de uma linha (─); não há moldura ao redor dos painéis.
//
// # Distribuição exata
//
// O espaço útil é distribuído por inteiro, usando o método do maior resto: a
// soma dos tamanhos dos painéis visíveis mais os divisores é exatamente igual à
// largura (em [Horizontal]) ou à altura (em [Vertical]) informada em
// [Layout.SetSize]. Nenhuma coluna ou linha é perdida.
//
// # Redimensionamento
//
// [Layout.Resize] e [Layout.ResizeTo] alteram o painel focado. Os demais
// painéis cedem espaço, mas nunca ficam abaixo de [Pane.MinSize]. Painéis com
// [Pane.Fixed] marcado não cedem espaço automaticamente: apenas mudam de
// tamanho quando são o alvo explícito de um redimensionamento.
//
// # Degradação
//
// Quando não há espaço para todos os painéis respeitarem o tamanho mínimo, o
// layout mostra apenas os primeiros painéis que cabem e indica a quantidade
// restante com "+N" no título do último painel visível. Nesse caso,
// [Layout.Sizes] devolve zero para os painéis ocultos; [Layout.Visible] informa
// quantos estão visíveis.
package panes

import (
	"strconv"

	tea "charm.land/bubbletea/v2"
)

// Direction define o eixo em que os painéis são organizados.
type Direction int

const (
	// Horizontal organiza os painéis lado a lado, separados por uma coluna
	// divisória. A dimensão principal é a largura.
	Horizontal Direction = iota
	// Vertical organiza os painéis empilhados, separados por uma linha
	// divisória. A dimensão principal é a altura.
	Vertical
)

// tamanhoMinimoPadrao é o mínimo aplicado a painéis sem [Pane.MinSize]
// explícito.
const tamanhoMinimoPadrao = 3

// Pane descreve um painel do layout.
type Pane struct {
	// ID identifica o painel de forma única dentro do layout. Adicionar um
	// painel com um ID já existente substitui o anterior.
	ID string
	// Title é o texto exibido na linha de título do painel. O título também
	// indica o foco e a quantidade de painéis ocultos na degradação.
	Title string
	// Content é o conteúdo textual do painel. Ele é quebrado na largura
	// disponível e truncado quando não couber na altura.
	Content string
	// MinSize é o tamanho mínimo, em colunas ([Horizontal]) ou linhas
	// ([Vertical]). Valores menores ou iguais a zero usam o padrão 3.
	MinSize int
	// Fixed impede que o painel ceda espaço em redimensionamentos de outros
	// painéis. O painel ainda pode ser redimensionado explicitamente.
	Fixed bool
}

// Layout é um conjunto ordenado de painéis com foco e tamanhos
// redimensionáveis.
//
// O valor zero é utilizável e se comporta como um layout horizontal vazio, mas
// começa sem área (0x0). Prefira [New], que define a direção e parte de 80x24.
type Layout struct {
	dir    Direction
	panes  []Pane
	sizes  []int
	focus  int
	width  int
	height int
	// dirty indica que o usuário já ajustou tamanhos manualmente. Enquanto
	// estiver falso, adicionar um painel redistribui o espaço igualmente.
	dirty bool
}

// New cria um layout vazio na direção informada.
//
// Direções desconhecidas são tratadas como [Horizontal]. O layout começa com
// 80x24 até que [Layout.SetSize] seja chamado.
func New(dir Direction) *Layout {
	if dir != Vertical {
		dir = Horizontal
	}
	return &Layout{
		dir:    dir,
		focus:  -1,
		width:  80,
		height: 24,
	}
}

// Add adiciona um painel ao fim do layout e devolve o próprio layout para
// encadeamento.
//
// Se já existir um painel com o mesmo ID, ele é substituído no lugar, sem
// alterar a ordem nem o foco. O [Pane.MinSize] é normalizado: valores menores
// ou iguais a zero viram 3.
func (l *Layout) Add(p Pane) *Layout {
	p.MinSize = normalizarMinimo(p.MinSize)

	if i := l.indice(p.ID); i >= 0 {
		l.panes[i] = p
		l.fit()
		return l
	}

	alvo := l.tamanhoMedio()
	l.panes = append(l.panes, p)
	l.sizes = append(l.sizes, alvo)
	if l.focus < 0 {
		l.focus = 0
	}

	if l.dirty {
		l.fit()
		return l
	}
	// Layout novo: todos os painéis começam com o mesmo tamanho,
	// independentemente de Fixed ou MinSize.
	l.resetTamanhos()
	return l
}

// Remove remove o painel com o ID informado e informa se ele existia.
//
// O foco é mantido no mesmo índice quando possível; se o painel removido era o
// último, o foco passa para o novo último painel.
func (l *Layout) Remove(id string) bool {
	i := l.indice(id)
	if i < 0 {
		return false
	}

	l.panes = append(l.panes[:i], l.panes[i+1:]...)
	l.sizes = append(l.sizes[:i], l.sizes[i+1:]...)

	switch {
	case len(l.panes) == 0:
		l.focus = -1
	case l.focus > i:
		l.focus--
	case l.focus == i && l.focus >= len(l.panes):
		l.focus = len(l.panes) - 1
	}

	l.fit()
	return true
}

// IDrop remove o painel com o ID informado, sem devolver indicação de sucesso.
// É um atalho para [Layout.Remove] quando o resultado booleano não interessa.
func (l *Layout) IDrop(id string) {
	l.Remove(id)
}

// Pane devolve uma cópia do painel com o ID informado.
func (l *Layout) Pane(id string) (Pane, bool) {
	if i := l.indice(id); i >= 0 {
		return l.panes[i], true
	}
	return Pane{}, false
}

// SetContent substitui o conteúdo do painel com o ID informado. IDs
// inexistentes são ignorados.
func (l *Layout) SetContent(id, content string) {
	if i := l.indice(id); i >= 0 {
		l.panes[i].Content = content
	}
}

// SetTitle substitui o título do painel com o ID informado. IDs inexistentes
// são ignorados.
func (l *Layout) SetTitle(id, title string) {
	if i := l.indice(id); i >= 0 {
		l.panes[i].Title = title
	}
}

// IDs devolve os identificadores dos painéis na ordem do layout.
func (l *Layout) IDs() []string {
	if len(l.panes) == 0 {
		return nil
	}
	ids := make([]string, len(l.panes))
	for i, p := range l.panes {
		ids[i] = p.ID
	}
	return ids
}

// SetSize define a largura e a altura disponíveis para o layout.
//
// Valores negativos são tratados como zero. O layout preserva as proporções
// atuais dos painéis; se o espaço ficar menor que o mínimo necessário, os
// painéis excedentes são ocultados.
func (l *Layout) SetSize(width, height int) {
	l.width = max(width, 0)
	l.height = max(height, 0)
	l.fit()
}

// Size devolve a largura e a altura disponíveis para o layout.
func (l *Layout) Size() (width, height int) {
	return l.width, l.height
}

// Focus move o foco para o índice informado, limitado ao intervalo válido.
//
// Índices negativos viram 0 e índices maiores que o último painel viram o
// último. Em um layout vazio o foco fica em -1.
func (l *Layout) Focus(index int) {
	if len(l.panes) == 0 {
		l.focus = -1
		return
	}
	l.focus = min(max(index, 0), len(l.panes)-1)
}

// Focused devolve o índice do painel focado, ou -1 se o layout estiver vazio.
func (l *Layout) Focused() int {
	return l.focus
}

// FocusNext move o foco para o próximo painel, circulando do último para o
// primeiro.
func (l *Layout) FocusNext() {
	if len(l.panes) == 0 {
		return
	}
	if l.focus < 0 {
		l.focus = 0
		return
	}
	l.focus = (l.focus + 1) % len(l.panes)
}

// FocusPrev move o foco para o painel anterior, circulando do primeiro para o
// último.
func (l *Layout) FocusPrev() {
	if len(l.panes) == 0 {
		return
	}
	if l.focus < 0 {
		l.focus = 0
		return
	}
	l.focus = (l.focus - 1 + len(l.panes)) % len(l.panes)
}

// Init integra o layout ao ciclo do Bubble Tea. Não dispara comandos iniciais.
func (l *Layout) Init() tea.Cmd {
	return nil
}

// Update trata as teclas de navegação e de redimensionamento, além de
// mensagens de tamanho de terminal.
//
// Teclas tratadas:
//
//   - tab e shift+tab: próximo e anterior;
//   - alt+seta: redimensiona o focado no eixo principal;
//   - + e -: redimensiona o focado em uma unidade.
//
// [tea.WindowSizeMsg] chama [Layout.SetSize]. Nenhuma tecla produz comando.
func (l *Layout) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		l.trataTecla(msg.String())
	}
	return nil
}

// ViewModel devolve a visualização do layout como um [tea.View], pronta para
// ser usada em um modelo do Bubble Tea.
func (l Layout) ViewModel() tea.View {
	return tea.NewView(l.View())
}

// trataTecla aplica uma tecla representada pela string do Bubble Tea, como
// "tab", "shift+tab", "alt+left" ou "+". Teclas desconhecidas são ignoradas.
func (l *Layout) trataTecla(tecla string) {
	switch tecla {
	case "tab":
		l.FocusNext()
	case "shift+tab":
		l.FocusPrev()
	case "alt+left":
		if l.dir == Horizontal {
			l.Resize(-1)
		}
	case "alt+right":
		if l.dir == Horizontal {
			l.Resize(1)
		}
	case "alt+up":
		if l.dir == Vertical {
			l.Resize(-1)
		}
	case "alt+down":
		if l.dir == Vertical {
			l.Resize(1)
		}
	case "+", "=":
		l.Resize(1)
	case "-", "_":
		l.Resize(-1)
	}
}

// indice devolve a posição do painel com o ID informado, ou -1.
func (l *Layout) indice(id string) int {
	for i, p := range l.panes {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// normalizarMinimo aplica o mínimo padrão a valores não positivos.
func normalizarMinimo(m int) int {
	if m <= 0 {
		return tamanhoMinimoPadrao
	}
	return m
}

// minimo devolve o tamanho mínimo efetivo do painel no índice informado.
func (l *Layout) minimo(i int) int {
	return normalizarMinimo(l.panes[i].MinSize)
}

// mainSize devolve a dimensão principal do layout: largura em [Horizontal],
// altura em [Vertical].
func (l *Layout) mainSize() int {
	if l.dir == Vertical {
		return l.height
	}
	return l.width
}

// tamanhoMedio devolve a média dos tamanhos visíveis, usada como alvo de um
// painel recém-adicionado. Sem painéis visíveis, devolve o mínimo padrão.
func (l *Layout) tamanhoMedio() int {
	soma, n := 0, 0
	for _, s := range l.sizes {
		if s > 0 {
			soma += s
			n++
		}
	}
	if n == 0 {
		return tamanhoMinimoPadrao
	}
	return max(soma/n, tamanhoMinimoPadrao)
}

// numeroOcultos devolve quantos painéis existem além dos visíveis.
func (l *Layout) numeroOcultos(visiveis int) int {
	return max(len(l.panes)-visiveis, 0)
}

// rotuloOcultos devolve o indicador "+N" usado no título do último painel
// visível quando há painéis ocultos.
func rotuloOcultos(n int) string {
	if n <= 0 {
		return ""
	}
	return "+" + strconv.Itoa(n)
}
