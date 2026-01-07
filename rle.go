package main

import (
	"bytes"
	"errors"
)

// Compress recebe dados brutos e retorna dados comprimidos em RLE
func Compress(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	var buffer bytes.Buffer
	// Olha para frente para ver se o próximo byte é igual ao atual
	// limitar count a 255 porque precisa que caiba em 1 único Byte (uint8)
	for i := 0; i < len(data); {
		count := 1
		for i+count < len(data) && data[i] == data[i+count] && count < 255 {
			count++
		}

		// Escreve primeiro a quantidade (Contagem)
		buffer.WriteByte(byte(count))
		// Escreve qual é o caractere (O Byte)
		buffer.WriteByte(data[i])

		// Pula o índice para o próximo bloco de caracteres novos
		i += count
	}

	return buffer.Bytes()
}

// Decompress faz o inverso: lê [count][char] e expande
func Decompress(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	// Precisa ler de 2 em 2 bytes (par: contagem + valor)
	for i := 0; i < len(data); i += 2 {
		// Proteção para não ler fora do array se o arquivo estiver corrompido
		if i+1 >= len(data) {
			return nil, errors.New("Arquivo comprimido ou incompleto")
		}

		// Quantas vezes repetir
		count := int(data[i])
		// Qual valor repetir
		val := data[i+1]

		// Escreve o valor 'count' vezes
		for j := 0; j < count; j++ {
			buffer.WriteByte(val)
		}
	}

	return buffer.Bytes(), nil
}
