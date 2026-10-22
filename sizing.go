package panes

import "sort"

// Sizes devolve o tamanho atual de cada painel na dimensão principal, na ordem
// do layout.
//
// Em [Horizontal] os valores são larguras; em [Vertical], alturas. Painéis
// ocultos pela degradação têm tamanho zero. Para os painéis visíveis, a soma
// dos tamanhos mais um divisor entre cada par é exatamente a largura (ou
// altura) disponível.
func (l *Layout) Sizes() []int {
	if len(l.panes) == 0 {
		return nil
	}
	tamanhos := make([]int, len(l.sizes))
	copy(tamanhos, l.sizes)
	return tamanhos
}

// Visible devolve quantos painéis cabem na dimensão atual. O restante, se
// houver, é indicado por "+N" no título do último painel visível.
func (l *Layout) Visible() int {
	return l.visibleCount(l.mainSize())
}

// Resize aumenta (delta positivo) ou diminui (delta negativo) o painel focado,
// movendo espaço entre ele e os demais.
//
// O crescimento é limitado pelo espaço que os outros painéis podem ceder sem
// ficar abaixo de [Pane.MinSize]; painéis [Pane.Fixed] não cedem. A diminuição
// nunca deixa o painel focado abaixo do mínimo. Se o delta não couber, a maior
// parte possível é aplicada. Sem painel focado visível, nada acontece.
func (l *Layout) Resize(delta int) {
	if delta == 0 {
		return
	}
	l.fit()

	visiveis := l.visibleCount(l.mainSize())
	if visiveis == 0 {
		return
	}
	foco := l.focus
	if foco < 0 || foco >= visiveis {
		return
	}

	if delta > 0 {
		l.crescerFocado(foco, delta, visiveis)
		return
	}
	l.diminuirFocado(foco, -delta, visiveis)
}

// ResizeTo define o tamanho exato do painel focado, limitado pelo mínimo e pelo
// espaço disponível.
//
// Quando o valor pedido não cabe, o painel recebe o maior tamanho possível; a
// soma total permanece exata. Sem painel focado visível, nada acontece.
func (l *Layout) ResizeTo(size int) {
	l.fit()
	if l.focus < 0 || l.focus >= len(l.panes) {
		return
	}
	if size < l.minimo(l.focus) {
		size = l.minimo(l.focus)
	}
	current := l.sizes[l.focus]
	if size == current {
		return
	}
	l.Resize(size - current)
}

// Reset devolve todos os painéis visíveis ao mesmo tamanho, respeitando
// [Pane.MinSize].
//
// O foco não é alterado. Depois de um Reset, adicionar um painel volta a
// distribuir o espaço igualmente. A partir daí, painéis [Pane.Fixed] mantêm o
// tamanho recebido.
func (l *Layout) Reset() {
	l.resetTamanhos()
	l.dirty = false
}

// resetTamanhos distribui o espaço igualmente entre os painéis visíveis,
// corrigindo eventuais mínimos maiores que a fatia igualitária.
func (l *Layout) resetTamanhos() {
	n := len(l.panes)
	if n == 0 {
		l.sizes = nil
		return
	}
	if len(l.sizes) != n {
		ajustado := make([]int, n)
		copy(ajustado, l.sizes)
		l.sizes = ajustado
	}

	total := l.mainSize()
	visiveis := l.visibleCount(total)
	for i := range l.sizes {
		l.sizes[i] = 0
	}
	if visiveis == 0 {
		return
	}

	disponivel := total - (visiveis - 1)
	base := disponivel / visiveis
	resto := disponivel % visiveis
	for i := 0; i < visiveis; i++ {
		l.sizes[i] = base
		if i < resto {
			l.sizes[i]++
		}
	}
	l.corrigirMinimos(visiveis)
}

// fit recalcula os tamanhos para a dimensão atual, preservando as proporções e
// as restrições de mínimo e de [Pane.Fixed].
func (l *Layout) fit() {
	n := len(l.panes)
	if n == 0 {
		l.sizes = nil
		l.focus = -1
		return
	}
	if len(l.sizes) != n {
		ajustado := make([]int, n)
		copy(ajustado, l.sizes)
		l.sizes = ajustado
	}

	total := l.mainSize()
	visiveis := l.visibleCount(total)
	if visiveis == 0 {
		for i := range l.sizes {
			l.sizes[i] = 0
		}
		return
	}

	disponivel := total - (visiveis - 1)
	if l.consistente(visiveis, disponivel) {
		return
	}

	alvos := make([]int, visiveis)
	minimos := make([]int, visiveis)
	fixos := make([]bool, visiveis)
	for i := 0; i < visiveis; i++ {
		minimos[i] = l.minimo(i)
		fixos[i] = l.panes[i].Fixed
		if l.sizes[i] > 0 {
			alvos[i] = l.sizes[i]
		} else {
			alvos[i] = minimos[i]
		}
	}

	resultado := distribuirEntre(disponivel, alvos, minimos, fixos)
	for i := 0; i < visiveis; i++ {
		l.sizes[i] = resultado[i]
	}
	for i := visiveis; i < n; i++ {
		l.sizes[i] = 0
	}
}

// consistente informa se os tamanhos atuais já respeitam a geometria informada.
func (l *Layout) consistente(visiveis, disponivel int) bool {
	soma := 0
	for i := 0; i < visiveis; i++ {
		if l.sizes[i] < l.minimo(i) {
			return false
		}
		soma += l.sizes[i]
	}
	for i := visiveis; i < len(l.sizes); i++ {
		if l.sizes[i] != 0 {
			return false
		}
	}
	return soma == disponivel
}

// visibleCount devolve o maior número de painéis iniciais que cabem na
// dimensão total, contando um divisor entre cada par.
//
// A lista é truncada a partir do fim: os primeiros painéis têm prioridade.
func (l *Layout) visibleCount(total int) int {
	if total <= 0 {
		return 0
	}
	usado := 0
	visiveis := 0
	for i := range l.panes {
		custo := l.minimo(i)
		if i > 0 {
			custo++ // divisor entre o painel anterior e este
		}
		if usado+custo > total {
			break
		}
		usado += custo
		visiveis++
	}
	return visiveis
}

// corrigirMinimos eleva painéis abaixo do mínimo e compensa retirando espaço
// de quem tem folga. A soma total é preservada.
func (l *Layout) corrigirMinimos(visiveis int) {
	deficit := 0
	for i := 0; i < visiveis; i++ {
		if m := l.minimo(i); l.sizes[i] < m {
			deficit += m - l.sizes[i]
			l.sizes[i] = m
		}
	}
	if deficit == 0 {
		return
	}
	comFolga := make([]int, 0, visiveis)
	for i := 0; i < visiveis; i++ {
		if l.sizes[i] > l.minimo(i) {
			comFolga = append(comFolga, i)
		}
	}
	l.reduzir(comFolga, deficit)
}

// crescerFocado aumenta o painel focado retirando espaço dos doadores.
func (l *Layout) crescerFocado(foco, delta, visiveis int) {
	doadores := make([]int, 0, visiveis)
	folga := 0
	for i := 0; i < visiveis; i++ {
		if i == foco || l.panes[i].Fixed {
			continue
		}
		if disponivel := l.sizes[i] - l.minimo(i); disponivel > 0 {
			doadores = append(doadores, i)
			folga += disponivel
		}
	}
	if folga == 0 {
		return
	}

	mover := min(delta, folga)
	l.reduzir(doadores, mover)
	l.sizes[foco] += mover
	l.dirty = true
}

// diminuirFocado reduz o painel focado e distribui o espaço liberado.
func (l *Layout) diminuirFocado(foco, delta, visiveis int) {
	receptores := make([]int, 0, visiveis)
	for i := 0; i < visiveis; i++ {
		if i == foco || l.panes[i].Fixed {
			continue
		}
		receptores = append(receptores, i)
	}
	if len(receptores) == 0 {
		// Todos os outros são fixos: para manter a soma exata, eles recebem
		// o espaço, mesmo sem ceder em outros redimensionamentos.
		for i := 0; i < visiveis; i++ {
			if i != foco {
				receptores = append(receptores, i)
			}
		}
	}
	if len(receptores) == 0 {
		// Painel único ou único visível: não há para quem transferir espaço.
		return
	}

	reduzir := min(delta, l.sizes[foco]-l.minimo(foco))
	if reduzir <= 0 {
		return
	}
	l.sizes[foco] -= reduzir
	l.aumentar(receptores, reduzir)
	l.dirty = true
}

// reduzir retira quantidade dos painéis indicados, proporcionalmente à folga
// de cada um acima do mínimo.
func (l *Layout) reduzir(indices []int, quantidade int) {
	if quantidade <= 0 || len(indices) == 0 {
		return
	}

	folgaTotal := 0
	for _, i := range indices {
		folgaTotal += l.sizes[i] - l.minimo(i)
	}
	if folgaTotal <= 0 {
		return
	}
	quantidade = min(quantidade, folgaTotal)

	usado := 0
	restos := make([]int64, len(indices))
	for j, i := range indices {
		folga := int64(l.sizes[i] - l.minimo(i))
		numero := int64(quantidade) * folga
		parte := int(numero / int64(folgaTotal))
		l.sizes[i] -= parte
		usado += parte
		restos[j] = numero % int64(folgaTotal)
	}

	falta := quantidade - usado
	for _, j := range indicesComMaiorResto(restos) {
		if falta == 0 {
			break
		}
		l.sizes[indices[j]]--
		falta--
	}
}

// aumentar distribui quantidade entre os painéis indicados,
// proporcionalmente ao tamanho atual.
func (l *Layout) aumentar(indices []int, quantidade int) {
	if quantidade <= 0 || len(indices) == 0 {
		return
	}

	total := 0
	for _, i := range indices {
		total += l.sizes[i]
	}
	if total <= 0 {
		base := quantidade / len(indices)
		resto := quantidade % len(indices)
		for j, i := range indices {
			l.sizes[i] += base
			if j < resto {
				l.sizes[i]++
			}
		}
		return
	}

	usado := 0
	restos := make([]int64, len(indices))
	for j, i := range indices {
		numero := int64(quantidade) * int64(l.sizes[i])
		parte := int(numero / int64(total))
		l.sizes[i] += parte
		usado += parte
		restos[j] = numero % int64(total)
	}

	falta := quantidade - usado
	for _, j := range indicesComMaiorResto(restos) {
		if falta == 0 {
			break
		}
		l.sizes[indices[j]]++
		falta--
	}
}

// distribuirEntre reparte o total entre painéis fixos e flexíveis.
//
// Painéis fixos tentam manter o tamanho alvo; se não couber, encolhem até o
// mínimo. Os flexíveis dividem todo o espaço restante.
func distribuirEntre(total int, alvos, minimos []int, fixos []bool) []int {
	n := len(minimos)
	resultado := make([]int, n)

	var indicesFixos, indicesFlex []int
	somaMinFixos, somaMinFlex := 0, 0
	somaAlvoFixos := 0
	for i := 0; i < n; i++ {
		if fixos[i] {
			indicesFixos = append(indicesFixos, i)
			somaMinFixos += minimos[i]
			somaAlvoFixos += max(alvos[i], minimos[i])
		} else {
			indicesFlex = append(indicesFlex, i)
			somaMinFlex += minimos[i]
		}
	}

	orcamentoFixos := somaAlvoFixos
	if maximo := total - somaMinFlex; orcamentoFixos > maximo {
		orcamentoFixos = maximo
	}
	orcamentoFixos = max(orcamentoFixos, somaMinFixos)

	alvosFixos := make([]int, len(indicesFixos))
	minimosFixos := make([]int, len(indicesFixos))
	for j, i := range indicesFixos {
		alvosFixos[j] = alvos[i]
		minimosFixos[j] = minimos[i]
	}
	fatiasFixas := dividirComMinimos(orcamentoFixos, alvosFixos, minimosFixos)
	for j, i := range indicesFixos {
		resultado[i] = fatiasFixas[j]
	}

	alvosFlex := make([]int, len(indicesFlex))
	minimosFlex := make([]int, len(indicesFlex))
	for j, i := range indicesFlex {
		alvosFlex[j] = alvos[i]
		minimosFlex[j] = minimos[i]
	}
	fatiasFlex := dividirComMinimos(total-orcamentoFixos, alvosFlex, minimosFlex)
	for j, i := range indicesFlex {
		resultado[i] = fatiasFlex[j]
	}

	return resultado
}

// dividirComMinimos reparte o orçamento garantindo o mínimo de cada fatia.
//
// A sobra acima dos mínimos é distribuída proporcionalmente à diferença entre
// o alvo e o mínimo, pelo método do maior resto. Empates são resolvidos pelo
// índice menor, o que mantém o resultado determinístico.
func dividirComMinimos(orcamento int, alvos, minimos []int) []int {
	n := len(minimos)
	fatias := make([]int, n)
	if n == 0 {
		return fatias
	}

	somaMin := 0
	for i := 0; i < n; i++ {
		fatias[i] = minimos[i]
		somaMin += minimos[i]
	}
	sobra := orcamento - somaMin
	if sobra <= 0 {
		return fatias
	}

	pesos := make([]int, n)
	somaPesos := 0
	for i := 0; i < n; i++ {
		peso := alvos[i] - minimos[i]
		if peso < 0 {
			peso = 0
		}
		pesos[i] = peso
		somaPesos += peso
	}

	if somaPesos == 0 {
		base := sobra / n
		resto := sobra % n
		for i := 0; i < n; i++ {
			fatias[i] += base
			if i < resto {
				fatias[i]++
			}
		}
		return fatias
	}

	usado := 0
	restos := make([]int64, n)
	for i := 0; i < n; i++ {
		numero := int64(sobra) * int64(pesos[i])
		parte := int(numero / int64(somaPesos))
		fatias[i] += parte
		usado += parte
		restos[i] = numero % int64(somaPesos)
	}

	falta := sobra - usado
	for _, i := range indicesComMaiorResto(restos) {
		if falta == 0 {
			break
		}
		fatias[i]++
		falta--
	}
	return fatias
}

// indicesComMaiorResto devolve os índices ordenados pelo resto decrescente.
// Empates preservam a ordem original, garantindo determinismo.
func indicesComMaiorResto(restos []int64) []int {
	ordem := make([]int, len(restos))
	for i := range ordem {
		ordem[i] = i
	}
	sort.SliceStable(ordem, func(a, b int) bool {
		return restos[ordem[a]] > restos[ordem[b]]
	})
	return ordem
}
