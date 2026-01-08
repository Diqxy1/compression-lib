package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

const (
	TYPE_TEXT = 0
	TYPE_IMG  = 1
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Your Sync CLI - Uso:")
		fmt.Println("  run . compress <arquivo.png>  - Comprime uma imagem para .ys")
		fmt.Println("  run . view <arquivo.ys>      - Abre o visualizador web")
		return
	}

	command := os.Args[1]

	switch command {
	case "compress":
		if len(os.Args) < 3 {
			fmt.Println("Erro: informe o caminho da imagem.")
			return
		}
		execCompress(os.Args[2])

	case "view":
		if len(os.Args) < 3 {
			fmt.Println("Erro: informe o arquivo comprimido (pode ser .ys ou .txt")
			return
		}
		startYourSyncServer(os.Args[2])

	default:
		fmt.Println("Comando desconhecido.")
	}
}

func execCompress(inputPath string) {
	fmt.Printf("--- Your Sync: Comprimindo %s ---\n", inputPath)

	ext := strings.ToLower(inputPath)

	var rawData []byte
	var width int
	var dataType uint8

	// 1. Identificação de IMAGEM
	if strings.HasSuffix(ext, ".png") || strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		fmt.Printf("--- YourSync: Modo IMAGEM [%s] ---\n", inputPath)
		file, _ := os.Open(inputPath)
		defer file.Close()
		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Println("Erro ao decodificar imagem:", err)
			return
		}
		width = img.Bounds().Dx()
		rawData = imageToRGBBytes(img)
		dataType = TYPE_IMG

		// 2. Identificação de TEXTO (TXT ou CSV)
	} else if strings.HasSuffix(ext, ".txt") || strings.HasSuffix(ext, ".csv") {
		fmt.Printf("--- YourSync: Modo TEXTO [%s] ---\n", inputPath)
		var err error
		rawData, err = os.ReadFile(inputPath)
		if err != nil {
			fmt.Println("Erro ao ler arquivo:", err)
			return
		}
		dataType = TYPE_TEXT
		width = 0

		// 3. Bloqueio de outros formatos
	} else {
		fmt.Printf("Erro: O formato '%s' não é suportado.\n", ext)
		fmt.Println("Formatos aceitos: .png, .jpg, .txt, .csv")
		return
	}

	// 3. Execução da Compressão com Barra de Progresso
	var compressedBuffer bytes.Buffer

	// Inicia a compressão
	err := ViktorCompress(rawData, dataType, width, &compressedBuffer)
	if err != nil {
		fmt.Println("Erro na compressão:", err)
		return
	}

	// 4. Salva o arquivo .ys
	var outputName string

	ext = strings.ToLower(filepath.Ext(inputPath))

	if dataType == TYPE_IMG {
		outputName = "resultado.ys"
	} else {
		baseName := strings.TrimSuffix(inputPath, ext)
		outputName = baseName + "_comprimido.txt"
	}

	err = os.WriteFile(outputName, compressedBuffer.Bytes(), 0644)
	if err != nil {
		fmt.Println("Erro crítico ao salvar arquivo:", err)
		return
	}

	fmt.Printf("Sucesso! Economia: %.2f%%\n", 100.0-(float64(compressedBuffer.Len())/float64(len(rawData))*100.0))
}
