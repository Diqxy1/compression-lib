package main

import (
	"encoding/binary" // Desnecesario ?
	"io"
)

// Grava a tabela de frequências no início do ficheiro (OLD)
func WriteHeader(w io.Writer, freqs map[byte]int) error {
	// 1. Escreve quantos caracteres diferentes temos (1 byte)
	numEntries := uint8(len(freqs))

	if err := binary.Write(w, binary.LittleEndian, numEntries); err != nil {
		return err
	}

	// 2. Escreve cada par: [Byte][Frequência]
	for char, freq := range freqs {
		// Grava o byte
		if _, err := w.Write([]byte{char}); err != nil {
			return err
		}

		// Grava a frequência como um uint32 (4 bytes) para suportar ficheiros grandes
		if err := binary.Write(w, binary.LittleEndian, uint32(freq)); err != nil {
			return err
		}
	}

	return nil
}

// Percorre a árvore e grava 0 para nós e 1+byte para folhas
func serializeTree(node *Node, bw *BitWriter) {
	if node.Left == nil && node.Right == nil {
		bw.WriteBit(1) // É folha
		// Escreve os 8 bits do caractere
		for i := 7; i >= 0; i-- {
			bw.WriteBit(uint8((node.Char >> i) & 1))
		}

		return
	}

	bw.WriteBit(0) // Nó interno
	serializeTree(node.Left, bw)
	serializeTree(node.Right, bw)
}

// Inverso para o descompressor conseguir reconstruir a árvore (OLD)
func ReadHeader(r io.Reader) (map[byte]int, error) {
	var numEntries uint8
	if err := binary.Read(r, binary.LittleEndian, &numEntries); err != nil {
		return nil, err
	}

	freqs := make(map[byte]int)
	for i := 0; i < int(numEntries); i++ {
		//fmt.Printf("DEBUG: Reconstruindo tabela com %d entradas\n", numEntries)
		var char byte
		var freq uint32

		// Lê o byte
		b := make([]byte, 1)
		if _, err := r.Read(b); err != nil {
			return nil, err
		}

		char = b[0]

		// Lê a frequência
		if err := binary.Read(r, binary.LittleEndian, &freq); err != nil {
			return nil, err
		}

		freqs[char] = int(freq)
	}

	return freqs, nil
}

// Inverso para o serializer conseguir reconstruir a folha
func deserializeTree(br *BitReader) *Node {
	bit, _ := br.ReadBit()
	if bit == 1 {
		// É folha, lê os próximos 8 bits para formar o caractere
		var char uint8
		for i := 0; i < 8; i++ {
			b, _ := br.ReadBit()
			char = (char << 1) | b
		}

		return &Node{Char: char}
	}

	// É nó interno, cria os filhos
	return &Node{
		Left:  deserializeTree(br),
		Right: deserializeTree(br),
	}
}
