package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func Apply2DFilter(data []byte, width int) []byte {
	height := len(data) / width
	filtered := make([]byte, len(data))

	for y := range height {
		for x := range width {
			index := y*width + x
			current := data[index]

			var left, up byte

			if x > 0 {
				left = data[index-1]
			}

			if y > 0 {
				up = data[index-width]
			}

			prediction := byte((int(left) + int(up)) / 2)

			filtered[index] = current - prediction
		}
	}

	return filtered
}

func Apply2DFilterRGB(data []byte, width int) []byte {
	rowSize := width * 3
	filtered := make([]byte, len(data))

	for i := range data {
		var left, up byte

		// Vizinho da esquerda (3 bytes atrás, mesma cor)
		if i%rowSize >= 3 {
			left = data[i-3]
		}

		// Vizinho de cima (uma linha inteira atrás, mesma cor)
		if i >= rowSize {
			up = data[i-rowSize]
		}

		prediction := byte((int(left) + int(up)) / 2)
		filtered[i] = data[i] - prediction
	}
	return filtered
}

func Remove2DFilter(data []byte, width int) []byte {
	height := len(data) / width
	restored := make([]byte, len(data))

	for y := range height {
		for x := range width {
			index := y*width + x
			delta := data[index]

			var left, up byte

			if x > 0 {
				left = restored[index-1]
			}

			if y > 0 {
				up = restored[index-width]
			}

			prediction := byte((int(left) + int(up)) / 2)

			restored[index] = delta + prediction
		}
	}

	return restored
}

func Remove2DFilterRGB(data []byte, width int) []byte {
	rowSize := width * 3
	restored := make([]byte, len(data))

	for i := range data {
		var left, up byte

		if i%rowSize >= 3 {
			left = restored[i-3]
		}

		if i >= rowSize {
			up = restored[i-rowSize]
		}

		prediction := byte((int(left) + int(up)) / 2)
		restored[i] = data[i] + prediction
	}
	return restored
}

func imageToGrayscaleBytes(img image.Image) []byte {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	data := make([]byte, width*height)

	for y := range height {
		for x := range width {
			r, g, b, _ := img.At(x, y).RGBA()
			// Fórmula simples de luminância: 0.299R + 0.587G + 0.114B
			// O RGBA retorna valores de 16 bits, precisa dividir por 256 ou deslocar 8
			lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256.0
			data[y*width+x] = byte(lum)
		}
	}
	return data
}

func imageToRGBBytes(img image.Image) []byte {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	// cria um slice 3x maior para armazenar R, G e B
	data := make([]byte, width*height*3)

	i := 0
	for y := range height {
		for x := range width {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convertendo de 16-bit para 8-bit (0-255)
			data[i] = byte(r >> 8)
			data[i+1] = byte(g >> 8)
			data[i+2] = byte(b >> 8)
			i += 3
		}
	}
	return data
}

func saveBytesAsPNGRGB(data []byte, width int, filename string) error {
	// Calcula a altura baseada no tamanho total e largura (3 bytes por pixel)
	height := len(data) / (width * 3)

	// Cria uma nova imagem RGBA (o canal Alpha será fixado em 255 - opaco)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: data[idx],
				G: data[idx+1],
				B: data[idx+2],
				A: 255, // 255 = Sem transparência
			})
			idx += 3
		}
	}

	// Cria o arquivo no disco
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Salva codificando como PNG
	return png.Encode(f, img)
}

func saveYourSyncFile(data []byte, filename string) error {
	// Este arquivo conterá o Header, a Árvore de Huffman e os Bits LZ77
	return os.WriteFile(filename, data, 0644)
}

func buildImageObject(data []byte, width int) image.Image {
	// Calcula a altura baseada nos bytes (3 por pixel)
	height := len(data) / (width * 3)

	// Cria o objeto na memória
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Preenche o buffer de pixels diretamente
	idx := 0
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = data[idx]     // R
		img.Pix[i+1] = data[idx+1] // G
		img.Pix[i+2] = data[idx+2] // B
		img.Pix[i+3] = 255         // A (Opaco)
		idx += 3
	}

	return img
}
