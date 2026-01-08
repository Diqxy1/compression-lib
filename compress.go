package main

import (
	"encoding/binary" // Desnecesario ?
	"fmt"
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
		bw.WriteBits(1, 1) // Grava o bit '1' indicando que é uma folha

		// Grava os 9 bits do Símbolo de uma só vez
		bw.WriteBits(uint64(node.Symbol), 9)

		return
	}

	bw.WriteBits(0, 1) // Nó interno
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
	// Lê 1 bit para saber se é folha ou nó
	bit, err := br.ReadBits(1)
	if err != nil {
		return nil
	}

	if bit == 1 {
		// Se for folha, lê os 9 bits do símbolo de uma vez
		symbol, _ := br.ReadBits(9)
		return &Node{Symbol: int(symbol)}
	}

	// Se for nó interno (bit 0), reconstrói os filhos
	return &Node{
		Left:  deserializeTree(br),
		Right: deserializeTree(br),
	}
}

func ViktorCompress(data []byte, dataType uint8, width int, output io.Writer) error {
	output.Write([]byte{dataType})

	dataToCompress := data

	if dataType == TYPE_IMG {
		binary.Write(output, binary.LittleEndian, uint32(width))

		fmt.Println("Aplicando filtro 2D...")
		dataToCompress = Apply2DFilterRGB(data, width)
	} else {
		binary.Write(output, binary.LittleEndian, uint32(0))
	}

	return HuffmanCompress(dataToCompress, output)
}

func ViktorDecompress(r io.Reader) ([]byte, error) {
	typeBuf := make([]byte, 1)
	if _, err := r.Read(typeBuf); err != nil {
		return nil, err
	}

	dataType := typeBuf[0]

	var width uint32

	if dataType == TYPE_IMG {
		if err := binary.Read(r, binary.LittleEndian, &width); err != nil {
			return nil, err
		}
	}

	restored, err := HuffmanDecompress(r)
	if err != nil {
		return nil, err
	}

	if dataType == TYPE_IMG {
		fmt.Println("Removendo Filtro 2D...")
		restored = Remove2DFilterRGB(restored, int(width))
	}

	return restored, nil
}

func ViktorDecompressAndGetMetadata(r io.Reader) ([]byte, uint8, int, error) {
	typeBuf := make([]byte, 1)
	if _, err := r.Read(typeBuf); err != nil {
		return nil, 0, 0, err
	}
	dataType := typeBuf[0]

	var width uint32
	if err := binary.Read(r, binary.LittleEndian, &width); err != nil {
		return nil, 0, 0, err
	}

	restored, err := HuffmanDecompress(r)
	if err != nil {
		return nil, 0, 0, err
	}

	if dataType == TYPE_IMG {
		restored = Remove2DFilterRGB(restored, int(width))
	}

	return restored, dataType, int(width), nil
}
