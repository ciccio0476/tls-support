# TLS Support Checker

Un tool da riga di comando scritto in Go per verificare quali versioni di TLS (Transport Layer Security) sono supportate da un server HTTPS.

## 🚀 Caratteristiche

- ✅ Verifica supporto TLS 1.0, 1.1, 1.2 e 1.3
- ✅ Output chiaro e formattato con simboli visivi (✓/✗)
- ✅ Modalità insecure per testare server con certificati auto-firmati
- ✅ Timeout configurabile per connessioni
- ✅ Riepilogo finale delle versioni supportate

## 📋 Prerequisiti

- Go 1.16 o superiore (solo per compilazione)

## 🔧 Installazione

### Opzione 1: Scarica l'eseguibile pre-compilato

Se è disponibile un eseguibile pre-compilato, puoi scaricarlo direttamente e utilizzarlo.

### Opzione 2: Compila dal sorgente

```bash
# Clona il repository
git clone <url-repository>
cd tls-support

# Compila il programma
go build -o tls-support.exe main.go
```

### Opzione 3: Esegui direttamente con Go

```bash
go run main.go <URL>
```

## 📖 Utilizzo

### Sintassi base

```bash
tls-support.exe [opzioni] <URL>
```

### Opzioni disponibili

- `--insecure`: Ignora gli errori di verifica del certificato SSL/TLS (utile per certificati auto-firmati o scaduti)

### Esempi

#### Verifica versioni TLS supportate da Google

```bash
tls-support.exe https://www.google.com
```

**Output:**
```
Verifica delle versioni TLS supportate per: https://www.google.com
Host: www.google.com:443

Tentativo con TLS 1.0... ✓ Supportata
Tentativo con TLS 1.1... ✓ Supportata
Tentativo con TLS 1.2... ✓ Supportata
Tentativo con TLS 1.3... ✓ Supportata

======================
RIEPILOGO
======================

Versioni TLS supportate:
  ✓ TLS 1.0
  ✓ TLS 1.1
  ✓ TLS 1.2
  ✓ TLS 1.3
```

#### Verifica server con certificato auto-firmato

```bash
tls-support.exe --insecure https://self-signed.badssl.com
```

**Output:**
```
Verifica delle versioni TLS supportate per: https://self-signed.badssl.com
Host: self-signed.badssl.com:443
Modalità: INSECURE (ignora errori certificato)

Tentativo con TLS 1.0... ✓ Supportata
Tentativo con TLS 1.1... ✓ Supportata
Tentativo con TLS 1.2... ✓ Supportata
Tentativo con TLS 1.3... ✗ Non supportata

======================
RIEPILOGO
======================

Versioni TLS supportate:
  ✓ TLS 1.0
  ✓ TLS 1.1
  ✓ TLS 1.2

Versioni TLS non supportate:
  ✗ TLS 1.3
```

#### Mostra l'help

```bash
tls-support.exe --help
```


## 🔍 Come funziona

Il programma:

1. Accetta un URL HTTPS come input
2. Estrae l'host e la porta dall'URL
3. Per ogni versione TLS (1.0, 1.1, 1.2, 1.3):
   - Crea una configurazione TLS specifica per quella versione
   - Tenta una connessione al server
   - Esegue l'handshake TLS
   - Registra se la connessione ha avuto successo
4. Mostra un riepilogo delle versioni supportate e non supportate

## ⚙️ Configurazione

### Timeout connessione

Il timeout di connessione predefinito è di 5 secondi. Per modificarlo, modifica la riga nel codice sorgente:

```go
Timeout: 5 * time.Second,
```

### Versioni TLS da testare

Le versioni TLS testate sono definite in:

```go
var tlsVersionsToTest = []uint16{
    tls.VersionTLS10,
    tls.VersionTLS11,
    tls.VersionTLS12,
    tls.VersionTLS13,
}
```

## 🛡️ Note sulla sicurezza

> [!WARNING]
> L'opzione `--insecure` disabilita la verifica dei certificati SSL/TLS. Usala solo per test e debugging, mai in produzione o per dati sensibili.

- **TLS 1.0 e 1.1** sono considerati obsoleti e insicuri. Molti server moderni li hanno disabilitati.
- **TLS 1.2** è considerato sicuro ma si raccomanda di passare a TLS 1.3 quando possibile.
- **TLS 1.3** è la versione più recente e sicura del protocollo.

## 🤝 Contributi

I contributi sono benvenuti! Sentiti libero di aprire issue o pull request.

## 🔗 Link utili

- [TLS su Wikipedia](https://it.wikipedia.org/wiki/Transport_Layer_Security)
- [Go crypto/tls package](https://pkg.go.dev/crypto/tls)
- [BadSSL - Test SSL/TLS](https://badssl.com/)

## 📊 Casi d'uso

Questo tool è utile per:

- **Audit di sicurezza**: Verificare quali versioni TLS sono abilitate sui tuoi server
- **Debugging**: Identificare problemi di compatibilità TLS
- **Compliance**: Assicurarsi che versioni obsolete di TLS siano disabilitate
- **Testing**: Verificare la configurazione TLS dopo modifiche al server
