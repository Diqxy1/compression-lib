package main

import (
	"container/heap"
)

// Arvore
type Node struct {
	Char  byte
	Freq  int
	Left  *Node
	Right *Node
}

// PriorityQueue implementa heap.Interface e guarda os Nodes
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

// Menor frequência sai primeiro
func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Freq == pq[j].Freq {
		return pq[i].Char < pq[j].Char
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

func BuildTree(frequencies map[byte]int) *Node {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	// 1. Cria um nó para cada caractere e coloca na fila
	for char, freq := range frequencies {
		heap.Push(&pq, &Node{Char: char, Freq: freq})
	}

	// 2. Enquanto houver mais de um nó, une os dois menores
	for pq.Len() > 1 {
		left := heap.Pop(&pq).(*Node)
		right := heap.Pop(&pq).(*Node)

		if left.Freq == right.Freq && left.Char > right.Char {
			left, right = right, left
		}

		minChar := left.Char
		if right.Char < minChar {
			minChar = right.Char
		}

		// Cria um nó pai com a soma das frequências
		parent := &Node{
			Char:  minChar,
			Freq:  left.Freq + right.Freq,
			Left:  left,
			Right: right,
		}
		heap.Push(&pq, parent)
	}

	// O último nó restante é a raiz da árvore
	return heap.Pop(&pq).(*Node)
}

// Percorre a árvore recursivamente
func GenerateCodes(node *Node, code string, table map[byte]string) {
	if node == nil {
		return
	}

	if node.Left == nil && node.Right == nil {
		table[node.Char] = code
	}

	GenerateCodes(node.Left, code+"0", table)
	GenerateCodes(node.Right, code+"1", table)
}
