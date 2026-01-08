package main

import (
	"container/heap"
	"encoding/binary"
	"io"
)

// Arvore
type Node struct {
	Symbol int
	Freq   int
	Left   *Node
	Right  *Node
}

// PriorityQueue implementa heap.Interface e guarda os Nodes
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

// Menor frequência sai primeiro
func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Freq == pq[j].Freq {
		return pq[i].Symbol < pq[j].Symbol
	}
	return pq[i].Freq < pq[j].Freq
}

func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func BuildTree(frequencies map[int]int) *Node {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	// 1. Cria um nó para cada caractere e coloca na fila
	for symbol, freq := range frequencies {
		heap.Push(&pq, &Node{Symbol: symbol, Freq: freq})
	}

	// 2. Enquanto houver mais de um nó, une os dois menores
	for pq.Len() > 1 {
		left := heap.Pop(&pq).(*Node)
		right := heap.Pop(&pq).(*Node)

		if left.Freq == right.Freq && left.Symbol > right.Symbol {
			left, right = right, left
		}

		minSymbol := left.Symbol
		if right.Symbol < minSymbol {
			minSymbol = right.Symbol
		}

		// Cria um nó pai com a soma das frequências
		parent := &Node{
			Symbol: minSymbol,
			Freq:   left.Freq + right.Freq,
			Left:   left,
			Right:  right,
		}
		heap.Push(&pq, parent)
	}

	if pq.Len() == 0 {
		return nil
	}

	// O último nó restante é a raiz da árvore
	return heap.Pop(&pq).(*Node)
}

// Percorre a árvore recursivamente
func GenerateCodes(node *Node, code string, table map[int]string) {
	if node == nil {
		return
	}

	if node.Left == nil && node.Right == nil {
		table[node.Symbol] = code
	}

	GenerateCodes(node.Left, code+"0", table)
	GenerateCodes(node.Right, code+"1", table)
}

func HuffmanCompress(data []byte, output io.Writer) error {
	// 1. Converte dados brutos em uma sequência de símbolos (literais/lengths/distances)
	lz77Symbols := LZ77Compress(data)

	// 2. Contar frequências dos SÍMBOLOS LZ77
	symbolFrequencies := make(map[int]int)
	for _, symbol := range lz77Symbols {
		symbolFrequencies[symbol.Code]++
	}

	// 3. Criar a Árvore de Huffman a partir dos SÍMBOLOS
	root := BuildTree(symbolFrequencies)
	codes := make(map[int]string)
	GenerateCodes(root, "", codes)

	// 4. Header Escrever o tamanho original ANTES do BitWriter (para não sujar os bits)
	binary.Write(output, binary.LittleEndian, uint32(len(data)))

	// 5. Iniciar o BitWriter
	bw := NewBitWriter(output)

	// 6. Header serializar a arvore lz77 e huffman
	serializeTree(root, bw)

	// 7. Dados grava o corpo do arquivo SÍMBOLOS LZ77 codificados em Huffman
	for _, symbol := range lz77Symbols {
		code := codes[symbol.Code]
		for _, bitChar := range code {
			if bitChar == '1' {
				bw.WriteBits(1, 1)
			} else {
				bw.WriteBits(0, 1)
			}
		}
		if symbol.ExtraBits > 0 {
			bw.WriteBits(uint64(symbol.ExtraVal), uint8(symbol.ExtraBits))
		}
	}

	return bw.Flush()
}

func HuffmanDecompress(r io.Reader) ([]byte, error) {
	// 1. Ler o tamanho total de caracteres (4 bytes)
	var totalChars uint32
	if err := binary.Read(r, binary.LittleEndian, &totalChars); err != nil {
		return nil, err
	}

	// 2. Iniciar o BitReader
	br := newBitReader(r)

	// 3. Reconstruir a arvore a partir dos bits do cabeçalho
	root := deserializeTree(br)

	// 4. pre aloca o tamnho final para perfomance
	result := make([]byte, 0, totalChars)

	// 5. Decodificar os dados
	for uint32(len(result)) < totalChars {
		symbol := decodeNextSymbol(root, br)

		if symbol < 256 {
			result = append(result, byte(symbol))
		} else if symbol == 256 {
			break
		} else if symbol >= 257 && symbol <= 285 {
			baseLen, eBitsL := GetLengthBase(symbol)
			extraL, _ := br.ReadBits(uint8(eBitsL))
			finalLen := baseLen + int(extraL)

			distSymbol := decodeNextSymbol(root, br)
			baseDist, eBitsD := GetDistanceBase(distSymbol)
			extraD, _ := br.ReadBits(uint8(eBitsD))
			finalDist := baseDist + int(extraD)

			startIndex := len(result) - finalDist
			for k := range finalLen {
				result = append(result, result[startIndex+k])
			}
		}
	}

	return result, nil
}

func decodeNextSymbol(root *Node, br *BitReader) int {
	curr := root
	for curr.Left != nil || curr.Right != nil {
		bit, _ := br.ReadBits(1)
		if bit == 0 {
			curr = curr.Left
		} else {
			curr = curr.Right
		}
	}
	return curr.Symbol
}
