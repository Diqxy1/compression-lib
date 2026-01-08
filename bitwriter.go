package main

import (
	"bufio"
	"io"
)

type BitWriter struct {
	writer *bufio.Writer // bufio é essencial para performance de disco
	cache  uint64        // Acumulador de bits (até 64 bits)
	bits   uint8         // Quantos bits estão ocupados no cache
}

func NewBitWriter(w io.Writer) *BitWriter {
	return &BitWriter{
		writer: bufio.NewWriter(w),
		cache:  0,
		bits:   0,
	}
}

func (bw *BitWriter) WriteBits(val uint64, nbits uint8) error {
	val &= (1 << nbits) - 1

	bw.cache = (bw.cache << nbits) | val
	bw.bits += nbits

	for bw.bits >= 8 {
		byteToWrite := byte(bw.cache >> (bw.bits - 8))
		if err := bw.writer.WriteByte(byteToWrite); err != nil {
			return err
		}
		bw.bits -= 8
		bw.cache &= (1 << bw.bits) - 1
	}
	return nil
}

// Escreve os bits restantes se o último byte não estiver completo
func (bw *BitWriter) Flush() error {
	if bw.bits > 0 {
		byteToWrite := byte(bw.cache << (8 - bw.bits))
		if err := bw.writer.WriteByte(byteToWrite); err != nil {
			return err
		}
		bw.bits = 0
		bw.cache = 0
	}
	return bw.writer.Flush()
}
