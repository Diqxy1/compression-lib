package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func main() {
	input := []byte("PIED PIPER COMPRESSION TEST - GO GO GO!")

	// Simulando um arquivo usando um buffer em memória
	var compressedData bytes.Buffer

	// 1. Comprimindo
	err := HuffmanCompress(input, &compressedData)
	if err != nil {
		fmt.Println("Erro na compressão: ", err)
	}

	fmt.Printf("Original: %d bytes\n", len(input))
	fmt.Printf("Comprimindo (com cabeçalho): %d bytes\n", compressedData.Len())

	// 2. Descomprimindo
	//reader := bytes.NewReader(compressedData.Bytes())
	restored, err := HuffmanDecompress(&compressedData)
	if err != nil {
		fmt.Println("Erro na descompressão: ", err)
		return
	}

	fmt.Printf("Resultado: %s\n", string(restored))
}

func HuffmanCompress(data []byte, output io.Writer) error {
	// 1. Contar frequências e criar arvore
	freqs := make(map[byte]int)
	for _, b := range data {
		freqs[b]++
	}
	root := BuildTree(freqs)

	// 2. Gerar codigos
	codes := make(map[byte]string)
	GenerateCodes(root, "", codes)

	// 3. Iniciar o BitWriter
	bw := NewBitWriter(output)

	// 4. Header salvar o tamanho total do arquivo (4 bytes)
	binary.Write(output, binary.LittleEndian, uint32(len(data)))

	// 5. Header serializar a arvore bit a bit
	serializeTree(root, bw)

	// 6. Dados grava o corpo do arquivo
	for _, b := range data {
		code := codes[b]
		for _, bitChar := range code {
			if bitChar == '1' {
				bw.WriteBit(1)
			} else {
				bw.WriteBit(0)
			}
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

	// 4. Decodificar os dados
	var result []byte
	for i := 0; i < int(totalChars); i++ {
		currentNode := root
		for currentNode.Left != nil || currentNode.Right != nil {
			bit, err := br.ReadBit()
			if err != nil {
				return nil, err
			}

			if bit == 0 {
				currentNode = currentNode.Left
			} else {
				currentNode = currentNode.Right
			}
		}
		result = append(result, currentNode.Char)
	}

	return result, nil
}
