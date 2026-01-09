# 🛡️ Viktor Compression Engine (YourSync Edition)

O **Viktor** é uma biblioteca de compressão de alto desempenho escrita em Go, especializada em reduzir o tamanho de arquivos de texto e logs utilizando uma combinação de **LZ77** e **Codificação de Huffman**.

Esta versão foi compilada como uma *Shared Library*, permitindo integração direta com **Python**, C++, Rust e outras linguagens através de FFI (Foreign Function Interface).

---

## 📋 Funcionalidades

* **Compressão Ultra-Leve:** Otimizada para padrões repetitivos de logs de servidor.
* **Smart Viewer:** Visualize o conteúdo de arquivos `.ys` diretamente na memória RAM, sem precisar recriar arquivos pesados no disco.
* **Gerenciamento de Memória:** Inclui controle manual de desalocação para evitar Memory Leaks em bots que rodam 24/7.

---

## 🚀 Integração com Python (Bot de Logs)

Para usar o motor Viktor no seu bot, certifique-se de que o arquivo `viktor.so` (Linux) ou `viktor.dll` (Windows) esteja no diretório do seu projeto.

### 1. Configuração da Interface (ctypes)

```python
import ctypes
import os

# Carregar a biblioteca
lib = ctypes.CDLL("./viktor.so")

# Configurar o Viewer (Descompressão para RAM)
lib.ViktorViewData.argtypes = [ctypes.c_char_p, ctypes.c_int]
lib.ViktorViewData.restype = ctypes.c_char_p

# Configurar a Limpeza de Memória (Crucial para Servidores)
lib.ViktorFree.argtypes = [ctypes.c_char_p]
lib.ViktorFree.restype = None
```

### 2. Exemplo de Uso: Visualizador de Logs

```python
def ler_log_comprimido(caminho_ys):
    # 1. Lê os bytes do arquivo comprimido
    with open(caminho_ys, "rb") as f:
        dados_comprimidos = f.read()
    
    # 2. O Viktor processa e retorna um ponteiro para a string original
    ptr_resultado = lib.ViktorViewData(dados_comprimidos, len(dados_comprimidos))
    
    if ptr_resultado:
        # 3. Converte para string Python
        conteudo = ptr_resultado.decode('utf-8')
        
        # 4. LIBERA A MEMÓRIA NO GO (Evita consumo excessivo de RAM no servidor)
        lib.ViktorFree(ptr_resultado)
        
        return conteudo
    return "Erro ao processar o arquivo .ys"
```

## 🛠️ Referência da API (Exports)

| Função | Parâmetros | Retorno | Descrição |
| :--- | :--- | :--- | :--- |
| **ViktorCompressData** | `data, len, type` | `*char, len` | Comprime dados brutos para o formato `.ys`. |
| **ViktorViewData** | `data, len` | `*char` | Descomprime dados `.ys` diretamente para uma string na RAM. |
| **ViktorFree** | `pointer` | `void` | Libera a memória alocada pelo `C.CString` no motor Go. |

## ⚖️ Licença
### Desenvolvido por Diqxy1 - Projeto Viktor. Uso focado em eficiência de armazenamento de logs.