package main

import (
	"container/heap"
	"encoding/binary"
	"fmt"
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

		minSymbol := min(right.Symbol, left.Symbol)

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

func HuffmanCompress(data []byte, output io.Writer, isImage bool) error {
	lz77Symbols := LZ77Compress(data, isImage)
	fmt.Printf("[Compress] Símbolos LZ77 gerados: %d\n", len(lz77Symbols))

	symbolFrequencies := make(map[int]int)
	for _, symbol := range lz77Symbols {
		symbolFrequencies[symbol.Code]++
	}

	root := BuildTree(symbolFrequencies)
	codes := make(map[int]string)
	GenerateCodes(root, "", codes)

	binary.Write(output, binary.LittleEndian, uint32(len(data)))

	bw := NewBitWriter(output)
	serializeTree(root, bw)
	//bw.Flush()

	for i, symbol := range lz77Symbols {
		code, ok := codes[symbol.Code]
		if !ok {
			return fmt.Errorf("erro: símbolo %d não possui código huffman", symbol.Code)
		}

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

		// Log periódico para não travar o terminal
		if i%10000 == 0 {
			fmt.Printf("[Compress] Progresso: %d/%d símbolos processados\n", i, len(lz77Symbols))
		}
	}

	fmt.Println("[Compress] Gravação concluída. Fazendo Flush...")
	return bw.Flush()
}

func HuffmanDecompress(r io.Reader) ([]byte, error) {
	var totalChars uint32
	if err := binary.Read(r, binary.LittleEndian, &totalChars); err != nil {
		return nil, err
	}
	fmt.Printf("[Decompress] Iniciando. Tamanho esperado: %d bytes\n", totalChars)

	br := newBitReader(r)
	root := deserializeTree(br)
	//br.ByteAlign()
	if root == nil {
		return nil, fmt.Errorf("falha ao reconstruir árvore")
	}

	result := make([]byte, 0, totalChars)

	for uint32(len(result)) < totalChars {
		symbol := decodeNextSymbol(root, br)

		if symbol < 256 {
			result = append(result, byte(symbol))
		} else if symbol == 256 {
			fmt.Println("[Decompress] Símbolo EOF (256) encontrado.")
			break
		} else if symbol >= 257 && symbol <= 285 {
			baseLen, eBitsL := GetLengthBase(symbol)
			extraL, _ := br.ReadBits(uint8(eBitsL))
			finalLen := baseLen + int(extraL)

			distSymbol := decodeNextSymbol(root, br)

			if distSymbol < 300 || distSymbol > 331 {
				return nil, fmt.Errorf("Erro de Sincronia: Lido símbolo %d onde deveria ser uma Distância (300-331) na pos %d", distSymbol, len(result))
			}

			baseDist, eBitsD := GetDistanceBase(distSymbol)
			extraD, _ := br.ReadBits(uint8(eBitsD))
			finalDist := baseDist + int(extraD)

			// LOG DE DIAGNÓSTICO ANTES DO PANIC
			if finalDist > len(result) {
				fmt.Printf("\n--- ERRO DE SINCRONIA DETECTADO ---\n")
				fmt.Printf("Posição no buffer: %d\n", len(result))
				fmt.Printf("Símbolo Length: %d (Len real: %d)\n", symbol, finalLen)
				fmt.Printf("Símbolo Dist: %d (Dist real: %d)\n", distSymbol, finalDist)
				fmt.Printf("Problema: Distância aponta para fora do buffer inicial!\n")
				return nil, fmt.Errorf("distância inválida: %d na pos %d", finalDist, len(result))
			}

			for k := 0; k < finalLen; k++ {
				readIdx := len(result) - finalDist
				if readIdx < 0 {
					// Proteção final redundante
					return nil, fmt.Errorf("falha crítica de indexação: %d", readIdx)
				}
				result = append(result, result[readIdx])
			}
		} else {
			return nil, fmt.Errorf("símbolo desconhecido detectado: %d", symbol)
		}

		// Log de progresso a cada 20%
		if len(result) > 0 && len(result)%(int(totalChars)/5) == 0 {
			fmt.Printf("[Decompress] %d%% concluído (%d/%d)\n", (len(result)*100)/int(totalChars), len(result), totalChars)
		}
	}

	fmt.Printf("[Decompress] Sucesso! Total: %d bytes\n", len(result))
	return result, nil
}

func decodeNextSymbol(root *Node, br *BitReader) int {
	curr := root
	for curr.Left != nil || curr.Right != nil {
		bit, err := br.ReadBits(1)
		if err != nil {
			return 256 // Se o bit acabar, força o fim
		}

		if bit == 0 {
			if curr.Left == nil {
				break
			} // Proteção contra árvore malformada
			curr = curr.Left
		} else {
			if curr.Right == nil {
				break
			} // Proteção contra árvore malformada
			curr = curr.Right
		}
	}
	return curr.Symbol
}
