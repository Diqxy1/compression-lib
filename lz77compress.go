package main

type LZ77Symbol struct {
	Code      int // O código que vai para a árvore de Huffman
	ExtraBits int // Quantidade de bits extras para gravar
	ExtraVal  int // O valor dos bits extras
}

func LZ77Compress(data []byte, isImage bool) []LZ77Symbol {
	const (
		windowSize = 65536
		windowMask = windowSize - 1
		hashSize   = 1 << 15
		hashMask   = hashSize - 1
		maxMatch   = 258
		//minMatch   = 6
		hShift = 6
	)

	minMatch := 6
	if !isImage {
		minMatch = 3
	}

	inputSize := len(data)
	if inputSize < 3 {
		return emitLiterals(data)
	}

	paddedData := make([]byte, inputSize+4)
	copy(paddedData, data)

	var symbols []LZ77Symbol
	symbols = make([]LZ77Symbol, 0, inputSize/2)

	head := make([]int, hashSize)
	for i := range head {
		head[i] = -1
	}
	prev := make([]int, windowSize)

	// Inicializa o hash com os dois primeiros bytes
	h := (uint32(paddedData[0]) << hShift) ^ uint32(paddedData[1])

	for i := 0; i < inputSize; {
		matchLen := 0
		matchDist := 0

		// TRAVA DE SEGURANÇA 1: Só calcula hash se houver o 3º byte disponível
		if i+2 < inputSize {
			h = ((uint32(paddedData[i]) << (hShift * 2)) ^
				(uint32(paddedData[i+1]) << hShift) ^
				uint32(paddedData[i+2])) & hashMask

			pos := head[h]
			prev[i&windowMask] = pos
			head[h] = i

			maxPossible := inputSize - i
			if maxPossible > maxMatch {
				maxPossible = maxMatch
			}

			chainLen := 32768
			for pos != -1 && i-pos <= windowSize && chainLen > 0 {
				// TRAVA DE SEGURANÇA 2: pos+matchLen deve estar dentro dos limites
				if paddedData[pos+matchLen] == paddedData[i+matchLen] && paddedData[pos] == paddedData[i] {
					currLen := 0
					for currLen < maxPossible && paddedData[pos+currLen] == paddedData[i+currLen] {
						currLen++
					}

					if currLen > matchLen {
						matchLen = currLen
						matchDist = i - pos
						if currLen >= maxMatch {
							break
						}
					}
				}
				pos = prev[pos&windowMask]
				chainLen--
			}
		}

		if matchLen >= minMatch {
			currentLen := matchLen
			currentDist := matchDist

			nextDist := 0

			// Espiada no próximo byte (i+1)
			nextLen := 0
			iNext := i + 1
			if iNext+2 < inputSize {
				// Cálculo rápido do hash para a próxima posição
				hNext := ((uint32(paddedData[iNext]) << 10) ^ (uint32(paddedData[iNext+1]) << 5) ^ uint32(paddedData[iNext+2])) & hashMask
				posNext := head[hNext]

				// Busca curta: só queremos saber se existe algo MAIOR que o match atual
				chainNext := 32
				for posNext != -1 && iNext-posNext <= windowSize && chainNext > 0 {
					// Otimização: só verificamos se o byte que superaria o match atual coincide
					if paddedData[posNext+currentLen] == paddedData[iNext+currentLen] {
						cL := 0
						for cL < maxMatch && iNext+cL < inputSize && paddedData[posNext+cL] == paddedData[iNext+cL] {
							cL++
						}
						if cL > nextLen {
							nextLen = cL
							nextDist = iNext - posNext
							if cL >= maxMatch {
								break
							}
						}
					}
					posNext = prev[posNext&windowMask]
					chainNext--
				}
			}

			// DECISÃO: Se o próximo match for estritamente melhor, adiamos o atual
			if nextLen > currentLen || (nextLen == currentLen && nextDist < currentDist/2) {
				symbols = append(symbols, LZ77Symbol{Code: int(paddedData[i])})
				i++
				continue
			}

			// Caso contrário, emite o match atual (que é o melhor)
			c, eb, ev := GetLengthData(currentLen)
			symbols = append(symbols, LZ77Symbol{Code: c, ExtraBits: eb, ExtraVal: ev})
			dc, deb, dev := GetDistanceData(currentDist)
			symbols = append(symbols, LZ77Symbol{Code: dc, ExtraBits: deb, ExtraVal: dev})

			// Atualiza o dicionário para os bytes consumidos
			for j := 1; j < currentLen; j++ {
				currIdx := i + j
				if currIdx+2 < inputSize {
					h = ((h << hShift) ^ uint32(paddedData[currIdx+2])) & hashMask
					prev[currIdx&windowMask] = head[h]
					head[h] = currIdx
				}
			}
			i += currentLen
		} else {
			// Literal (Match menor que minMatch)
			symbols = append(symbols, LZ77Symbol{Code: int(paddedData[i])})
			i++
		}
	}
	// EOF Symbol
	return append(symbols, LZ77Symbol{Code: 256})
}

func emitLiterals(data []byte) []LZ77Symbol {
	symbols := make([]LZ77Symbol, 0, len(data)+1)
	for _, b := range data {
		symbols = append(symbols, LZ77Symbol{Code: int(b)})
	}
	return append(symbols, LZ77Symbol{Code: 256})
}

// Retorna o código Huffman base e os bits extras necessários
func GetLengthData(length int) (code int, extraBits int, extraVal int) {
	switch {
	case length >= 3 && length <= 10:
		return 257 + (length - 3), 0, 0

	case length >= 11 && length <= 18:
		code = 265 + ((length - 11) / 2)
		extraBits = 1
		extraVal = (length - 11) % 2
		return

	case length >= 19 && length <= 34:
		code = 269 + ((length - 19) / 4)
		extraBits = 2
		extraVal = (length - 19) % 4
		return

	case length >= 35 && length <= 66:
		code = 273 + ((length - 35) / 8)
		extraBits = 3
		extraVal = (length - 35) % 8
		return

	case length >= 67 && length <= 130:
		code = 277 + ((length - 67) / 16)
		extraBits = 4
		extraVal = (length - 67) % 16
		return

	case length >= 131 && length <= 257:
		code = 281 + ((length - 131) / 32)
		extraBits = 5
		extraVal = (length - 131) % 32
		return

	case length == 258:
		// O 258 é especial, tem código próprio e 0 bits extras
		return 285, 0, 0

	default:
		return -1, 0, 0 // Erro: comprimento inválido
	}
}

func GetDistanceData(distance int) (code int, extraBits int, extraVal int) {
	const OFFSET = 300

	switch {
	case distance >= 1 && distance <= 4:
		return OFFSET + (distance - 1), 0, 0

	case distance >= 5 && distance <= 8:
		code = 4 + ((distance - 5) / 2)
		return OFFSET + code, 1, (distance - 5) % 2

	case distance >= 9 && distance <= 16:
		code = 6 + ((distance - 9) / 4)
		return OFFSET + code, 2, (distance - 9) % 4

	case distance >= 17 && distance <= 32:
		code = 8 + ((distance - 17) / 8)
		return OFFSET + code, 3, (distance - 17) % 8

	case distance >= 33 && distance <= 64:
		code = 10 + ((distance - 33) / 16)
		return OFFSET + code, 4, (distance - 33) % 16

	case distance >= 65 && distance <= 128:
		code = 12 + ((distance - 65) / 32)
		return OFFSET + code, 5, (distance - 65) % 32

	case distance >= 129 && distance <= 256:
		code = 14 + ((distance - 129) / 64)
		return OFFSET + code, 6, (distance - 129) % 64

	case distance >= 257 && distance <= 512:
		code = 16 + ((distance - 257) / 128)
		return OFFSET + code, 7, (distance - 257) % 128

	case distance >= 513 && distance <= 1024:
		code = 18 + ((distance - 513) / 256)
		return OFFSET + code, 8, (distance - 513) % 256

	case distance >= 1025 && distance <= 2048:
		code = 20 + ((distance - 1025) / 512)
		return OFFSET + code, 9, (distance - 1025) % 512

	case distance >= 2049 && distance <= 4096:
		code = 22 + ((distance - 2049) / 1024)
		return OFFSET + code, 10, (distance - 2049) % 1024

	case distance >= 4097 && distance <= 8192:
		code = 24 + ((distance - 4097) / 2048)
		return OFFSET + code, 11, (distance - 4097) % 2048

	case distance >= 8193 && distance <= 16384:
		code = 26 + ((distance - 8193) / 4096)
		return OFFSET + code, 12, (distance - 8193) % 4096

	case distance >= 16385 && distance <= 32768:
		code = 28 + ((distance - 16385) / 8192)
		return OFFSET + code, 13, (distance - 16385) % 8192

	case distance >= 32769 && distance <= 49152:
		return OFFSET + 30, 14, distance - 32769

	case distance >= 49153 && distance <= 65536:
		return OFFSET + 31, 14, distance - 49153

	default:
		// Se cair aqui, precisa implementar os casos intermediários (bits 6 a 12)
		// A lógica é sempre: divide por 2^bits e pega o resto
		return -1, 0, 0
	}
}

func GetLengthBase(code int) (base int, extraBits int) {
	// Subtraímos 257 porque os códigos de comprimento começam em 257
	c := code - 257
	switch {
	case c >= 0 && c <= 7: // 257 - 264
		return c + 3, 0

	case c >= 8 && c <= 11: // 265 - 268
		return 11 + (c-8)*2, 1

	case c >= 12 && c <= 15: // 269 - 272
		return 19 + (c-12)*4, 2

	case c >= 16 && c <= 19: // 273 - 276
		return 35 + (c-16)*8, 3

	case c >= 20 && c <= 23: // 277 - 280
		return 67 + (c-20)*16, 4

	case c >= 24 && c <= 27: // 281 - 284
		return 131 + (c-24)*32, 5

	case c == 28: // 285
		return 258, 0
	}

	return 0, 0
}

func GetDistanceBase(code int) (base int, extraBits int) {
	const OFFSET = 300
	c := code - OFFSET

	switch {
	case c >= 0 && c <= 3:
		return c + 1, 0

	case c >= 4 && c <= 5: // 304 - 305
		return 5 + (c-4)*2, 1

	case c >= 6 && c <= 7: // 306 - 307
		return 9 + (c-6)*4, 2

	case c >= 8 && c <= 9: // 308 - 309
		return 17 + (c-8)*8, 3

	case c >= 10 && c <= 11: // 310 - 311
		return 33 + (c-10)*16, 4

	case c >= 12 && c <= 13: // 312 - 313
		return 65 + (c-12)*32, 5

	case c >= 14 && c <= 15:
		return 129 + (c-14)*64, 6

	case c >= 16 && c <= 17:
		return 257 + (c-16)*128, 7

	case c >= 18 && c <= 19:
		return 513 + (c-18)*256, 8

	case c >= 20 && c <= 21:
		return 1025 + (c-20)*512, 9

	case c >= 22 && c <= 23:
		return 2049 + (c-22)*1024, 10

	case c >= 24 && c <= 25:
		return 4097 + (c-24)*2048, 11

	case c >= 26 && c <= 27:
		return 8193 + (c-26)*4096, 12

	case c >= 28 && c <= 29:
		return 16385 + (c-28)*8192, 13

	case c == 30:
		return 32769, 14

	case c == 31:
		return 49153, 14
	}

	return 0, 0
}
