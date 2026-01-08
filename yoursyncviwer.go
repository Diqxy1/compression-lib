package main

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"
)

func YoursyncToMemory(inputPath string) (image.Image, error) {
	// 1. Abre o arquivo ys
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 2. Descompressão
	restored, dataType, width, err := ViktorDecompressAndGetMetadata(file)
	if err != nil {
		return nil, err
	}

	if dataType != TYPE_IMG {
		return nil, fmt.Errorf("o arquivo não contém dados de imagem")
	}

	// 3. Monta o objeto de imagem na RAM
	height := len(restored) / (width * 3)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Preenche os pixels (R, G, B, A)
	idx := 0
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = restored[idx]     // R
		img.Pix[i+1] = restored[idx+1] // G
		img.Pix[i+2] = restored[idx+2] // B
		img.Pix[i+3] = 255             // A
		idx += 3
	}

	return img, nil
}

func showProgress(current, total int) {
	percent := float64(current) / float64(total) * 100
	bars := int(percent / 5)
	strBars := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
	fmt.Printf("\r[%s] %.1f%%", strBars, percent)
	if current == total {
		fmt.Println()
	}
}

func DecompressToInterface(inputPath string) ([]byte, uint8, int, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	restored, dataType, width, err := ViktorDecompressAndGetMetadata(file)
	return restored, dataType, width, err
}

func startYourSyncServer(ppPath string) {
	http.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open(ppPath)
		defer file.Close()
		restored, dataType, width, _ := ViktorDecompressAndGetMetadata(file)

		if dataType == TYPE_IMG {
			img := buildImageObject(restored, width)
			w.Header().Set("Content-Type", "image/png")
			png.Encode(w, img)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		restoredData, dataType, _, err := DecompressToInterface(ppPath)
		duration := time.Since(start)

		if err != nil {
			fmt.Fprintf(w, "Erro na descompressão: %v", err)
			return
		}

		fileInfo, _ := os.Stat(ppPath)
		sizeKB := fileInfo.Size() / 1024

		var contentHTML string
		if dataType == TYPE_IMG {
			contentHTML = `<img src="/raw" />`
		} else {
			contentHTML = fmt.Sprintf(`
                <div class="text-container">
                    <pre>%s</pre>
                </div>`, string(restoredData))
		}

		fmt.Fprintf(w, `
            <html>
                <head>
                    <title>Your Sync Dashboard</title>
                    <style>
                        body { font-family: 'Segoe UI', Tahoma, sans-serif; background: #121212; color: #e0e0e0; text-align: center; padding: 20px; }
                        .stats { background: #1e1e1e; padding: 25px; border-radius: 15px; border: 1px solid #333; display: inline-block; box-shadow: 0 4px 15px rgba(0,0,0,0.5); }
                        .highlight { color: #00ff88; font-weight: bold; }
                        img { margin-top: 30px; border: 2px solid #00ff88; border-radius: 8px; max-width: 95%%; box-shadow: 0 0 20px rgba(0,255,136,0.2); }
                        .text-container { margin-top: 30px; background: #000; padding: 20px; text-align: left; display: inline-block; border-radius: 8px; border: 1px solid #444; max-width: 90%%; overflow-x: auto; }
                        pre { color: #00ff88; font-family: 'Consolas', monospace; margin: 0; }
                    </style>
                </head>
                <body>
                    <h1>Your Sync <span class="highlight">Visualizer</span></h1>
                    <div class="stats">
                        <p>📦 Arquivo: <span class="highlight">%s</span></p>
                        <p>📏 Tamanho: <span class="highlight">%d KB</span></p>
                        <p>⚡ Descompressão: <span class="highlight">%v</span></p>
                        <p>🏷️ Tipo: <span class="highlight">%s</span></p>
                    </div>
                    <br>
                    %s
                </body>
            </html>
        `, ppPath, sizeKB, duration, getTypeName(dataType), contentHTML)
	})

	fmt.Println("🚀 Dashboard Pied Piper em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func getTypeName(t uint8) string {
	if t == TYPE_IMG {
		return "IMAGEM"
	}
	return "TEXTO"
}
