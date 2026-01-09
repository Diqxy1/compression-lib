# 🛡️ Viktor Compression Engine (YourSync Edition)

O **Viktor** é uma biblioteca de compressão de alto desempenho escrita em Go, especializada em reduzir o tamanho de arquivos de texto e logs utilizando uma combinação de **LZ77** e **Codificação de Huffman**.

Esta versão foi compilada como uma *Shared Library*, permitindo integração direta com **Python**, C++, Rust e outras linguagens através de FFI (Foreign Function Interface).

## 📊 Resultados e Performance

Baseado em testes reais com logs de sistema, o motor Viktor entrega uma eficiência notável:

* **Taxa de Compressão:** ~68.9% de economia de espaço (103 KB → 31 KB).
* **Velocidade de Acesso:** Descompressão em apenas **4.4ms**.
* **Acesso Instantâneo:** Visualização direta em RAM sem reconstrução física no disco.

---

## 📋 Funcionalidades

* **Compressão Ultra-Leve:** Algoritmo híbrido otimizado para padrões repetitivos em logs de servidor.
* **Smart Viewer:** Visualize o conteúdo de arquivos `.ys` diretamente na memória, economizando ciclos de I/O do SSD/HD.
* **Gerenciamento de Memória:** Inclui controle manual de desalocação (`ViktorFree`) para garantir estabilidade absoluta em bots e serviços que rodam 24/7.

---

## 🚀 Integração com Python (Bot de Logs)

Para usar o motor Viktor no seu bot, certifique-se de que o arquivo `viktor.so` (Linux) ou `viktor.dll` (Windows) esteja no diretório do seu projeto.

### 1. Configuração da Interface (ctypes)

```python
import ctypes
import os

# Carregar a biblioteca compilada em Go
lib = ctypes.CDLL("./viktor.so")

# Configurar a Compressão
# Retorna uma struct contendo o ponteiro da string e o tamanho
class GoSlice(ctypes.Structure):
    _fields_ = [("data", ctypes.c_char_p), ("len", ctypes.c_longlong), ("cap", ctypes.c_longlong)]

class CompressResult(ctypes.Structure):
    _fields_ = [("r0", ctypes.c_char_p), ("r1", ctypes.c_int)]

lib.ViktorCompressData.argtypes = [ctypes.c_char_p, ctypes.c_int, ctypes.c_uint8]
lib.ViktorCompressData.restype = CompressResult

# Configurar o Viewer (Descompressão direta para RAM)
lib.ViktorViewData.argtypes = [ctypes.c_char_p, ctypes.c_int]
lib.ViktorViewData.restype = ctypes.c_char_p

# Configurar a Limpeza de Memória
lib.ViktorFree.argtypes = [ctypes.c_char_p]
lib.ViktorFree.restype = None
```

### 2. Exemplo: Comprimindo um Arquivo de Log

```python
def comprimir_log(conteudo_texto):
    # Converte o texto para bytes
    dados_bytes = conteudo_texto.encode('utf-8')
    
    # Chama a compressão (Tipo 0 = Texto)
    res = lib.ViktorCompressData(dados_bytes, len(dados_bytes), 0)
    
    if res.r0:
        # Pega os bytes comprimidos do ponteiro retornado pelo Go
        dados_comprimidos = ctypes.string_at(res.r0, res.r1)
        
        # Salva o arquivo .ys no disco
        with open("log_comprimido.ys", "wb") as f:
            f.write(dados_comprimidos)
        
        # LIBERA A MEMÓRIA ALOCADA NO GO
        lib.ViktorFree(res.r0)
        print(f"Sucesso! Arquivo gerado com {len(dados_comprimidos)} bytes.")
```

### 3. Exemplo: Visualizando o Log (Viewer)

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
| **ViktorCompressData** | `char*, int, uint8` | `char*, int` | Comprime dados brutos para o formato `.ys`. |
| **ViktorViewData** | `char*, int` | `char*` | Descomprime dados `.ys` diretamente para uma string na RAM. |
| **ViktorFree** | `void*` | `void` | Libera a memória alocada pelo `C.CString` no motor Go. |

## ⚖️ Licença
### Desenvolvido por Diqxy1 - Projeto Viktor. Uso focado em eficiência de armazenamento de logs.