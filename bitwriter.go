package main

import (
	"io"
)

type BitWriter struct {
	// Grava os bytes finais
	writer io.Writer
	// Acumula os bits até formar um byte
	cache byte
	// Conta quantos bits tem no cache
	nBits uint
}

func NewBitWriter(w io.Writer) *BitWriter {
	return &BitWriter{writer: w}
}

// Escreve um único bit (0 ou 1)
func (bw *BitWriter) WriteBit(bit uint8) error {
	// Empurra os bits existentes para a esquerda e adiciona o novo bit na direita
	bw.cache = (bw.cache << 1) | (bit & 1)
	bw.nBits++

	// Se completou 8 bits, descarrega no writer
	if bw.nBits == 8 {
		if _, err := bw.writer.Write([]byte{bw.cache}); err != nil {
			return err
		}
		bw.nBits = 0
		bw.cache = 0
	}
	return nil
}

// Escreve os bits restantes se o último byte não estiver completo
func (bw *BitWriter) Flush() error {
	if bw.nBits > 0 {
		// Alinha os bits à esquerda para completar o byte
		bw.cache <<= (8 - bw.nBits)
		_, err := bw.writer.Write([]byte{bw.cache})
		return err
	}
	return nil
}
