package main

import (
	"bufio"
	"io"
)

type BitReader struct {
	reader *bufio.Reader
	cache  uint64 // Acumulador de bits
	bits   uint8  // Quantos bits úteis ainda restam no cache
}

func newBitReader(r io.Reader) *BitReader {
	return &BitReader{
		reader: bufio.NewReader(r),
	}
}

// Lê o próximo bit (retorna 0 ou 1)
func (br *BitReader) ReadBits(nbits uint8) (uint64, error) {
	for br.bits < nbits {
		nextByte, err := br.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		br.cache = (br.cache << 8) | uint64(nextByte)
		br.bits += 8
	}

	shift := br.bits - nbits
	val := (br.cache >> shift) & ((1 << nbits) - 1)

	br.bits -= nbits
	br.cache &= (1 << br.bits) - 1

	return val, nil
}
