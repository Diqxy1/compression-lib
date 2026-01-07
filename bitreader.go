package main

import (
	"io"
)

type BitReader struct {
	reader io.Reader
	// Armazena o byte atual que estamos processando
	byte byte
	// Quantos bits ainda restam para ler no byte atual
	bitsIn uint8
}

func newBitReader(r io.Reader) *BitReader {
	return &BitReader{reader: r}
}

// Lê o próximo bit (retorna 0 ou 1)
func (br *BitReader) ReadBit() (uint8, error) {
	if br.bitsIn == 0 {
		// Se não há mais bits no buffer, lê o próximo byte do arquivo
		buf := make([]byte, 1)
		_, err := br.reader.Read(buf)
		if err != nil {
			return 0, err
		}

		br.byte = buf[0]
		br.bitsIn = 8
	}

	// Extrai o bit mais significativo (da esquerda)
	bit := (br.byte >> (br.bitsIn - 1)) & 1
	br.bitsIn--
	return bit, nil
}
