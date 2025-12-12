package main

import (
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

// TLS version names for display
var tlsVersionNames = map[uint16]string{
	tls.VersionTLS10: "TLS 1.0",
	tls.VersionTLS11: "TLS 1.1",
	tls.VersionTLS12: "TLS 1.2",
	tls.VersionTLS13: "TLS 1.3",
}

// Test TLS versions to check
var tlsVersionsToTest = []uint16{
	tls.VersionTLS10,
	tls.VersionTLS11,
	tls.VersionTLS12,
	tls.VersionTLS13,
}

type TLSResult struct {
	URL       string
	Host      string
	Supported map[uint16]bool
	Error     error
}

func main() {
	// Define flags
	insecure := flag.Bool("insecure", false, "Ignora gli errori di verifica del certificato SSL/TLS")
	timeout := flag.Int("timeout", 5, "Timeout di connessione in secondi")
	csvFile := flag.String("csv", "", "Path al file CSV contenente la lista di URL")
	outFile := flag.String("out", "", "Path al file CSV di output per salvare i risultati")
	flag.Parse()

	// Check if we are in CSV mode
	if *csvFile != "" {
		processCSV(*csvFile, *outFile, *insecure, *timeout)
		return
	}

	// Single URL mode
	args := flag.Args()
	if len(args) < 1 {
		printHelp()
		os.Exit(1)
	}

	targetURL := args[0]
	result := checkDomain(targetURL, *insecure, *timeout)
	printSingleResult(result, *insecure)
}

func printHelp() {
	execName := filepath.Base(os.Args[0])
	fmt.Printf("Uso: %s [opzioni] <URL>\n", execName)
	fmt.Printf("Uso: %s -csv <input.csv> [-out <report.csv>] [opzioni]\n", execName)
	fmt.Println("\nOpzioni:")
	flag.PrintDefaults()
	fmt.Printf("\nEsempio: %s https://www.google.com\n", execName)
	fmt.Printf("Esempio: %s -csv urls.csv -out report.csv\n", execName)
}

func processCSV(filename string, outFilename string, insecure bool, timeout int) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Errore nell'apertura del file CSV: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Errore nella lettura del CSV: %v\n", err)
		os.Exit(1)
	}

	var csvWriter *csv.Writer
	var outFile *os.File

	if outFilename != "" {
		outFile, err = os.Create(outFilename)
		if err != nil {
			fmt.Printf("Errore nella creazione del file CSV output: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		csvWriter = csv.NewWriter(outFile)
		defer csvWriter.Flush()

		// Write header
		header := []string{"URL", "Host", "TLS 1.0", "TLS 1.1", "TLS 1.2", "TLS 1.3", "Error"}
		if err := csvWriter.Write(header); err != nil {
			fmt.Printf("Errore nella scrittura dell'header CSV: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Scansione di %d URL in corso...\n", len(records))
	if outFilename != "" {
		fmt.Printf("Salvataggio risultati in: %s\n", outFilename)
	}
	fmt.Println()

	// Setup tabwriter for clean output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "URL\tTLS 1.0\tTLS 1.1\tTLS 1.2\tTLS 1.3\tStatus")
	fmt.Fprintln(w, "---\t-------\t-------\t-------\t-------\t------")

	for _, record := range records {
		if len(record) == 0 {
			continue
		}

		inputURL := strings.TrimSpace(record[0])
		if inputURL == "" {
			continue
		}

		// Skip header if present (simple heuristic)
		if strings.ToLower(inputURL) == "url" {
			continue
		}

		// Add https if missing
		if !strings.HasPrefix(inputURL, "http") {
			inputURL = "https://" + inputURL
		}

		result := checkDomain(inputURL, insecure, timeout)

		// Format output row
		outRow := fmt.Sprintf("%s\t", result.URL)

		if result.Error != nil {
			outRow += fmt.Sprintf("-\t-\t-\t-\tError: %v", result.Error)
		} else {
			for _, v := range tlsVersionsToTest {
				if result.Supported[v] {
					outRow += "YES\t"
				} else {
					outRow += "NO\t"
				}
			}
			outRow += "OK"
		}
		fmt.Fprintln(w, outRow)

		// 2. Output to CSV file
		if csvWriter != nil {
			csvRow := []string{result.URL, result.Host}

			if result.Error != nil {
				csvRow = append(csvRow, "-", "-", "-", "-", result.Error.Error())
			} else {
				for _, v := range tlsVersionsToTest {
					if result.Supported[v] {
						csvRow = append(csvRow, "true")
					} else {
						csvRow = append(csvRow, "false")
					}
				}
				csvRow = append(csvRow, "") // No error
			}

			if err := csvWriter.Write(csvRow); err != nil {
				fmt.Printf("Errore scrittura riga CSV: %v\n", err)
			}
			csvWriter.Flush()
		}
	}
	w.Flush()

	if outFilename != "" {
		fmt.Printf("\nElaborazione completata. Risultati salvati in %s\n", outFilename)
	}
}

func checkDomain(targetURL string, insecure bool, timeoutSeconds int) TLSResult {
	result := TLSResult{
		URL:       targetURL,
		Supported: make(map[uint16]bool),
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		result.Error = err
		return result
	}

	if parsedURL.Scheme != "https" {
		result.Error = fmt.Errorf("protocollo non supportato: %s", parsedURL.Scheme)
		return result
	}

	host := parsedURL.Host
	if parsedURL.Port() == "" {
		host = host + ":443"
	}
	result.Host = host

	for _, version := range tlsVersionsToTest {
		supported := testTLSVersion(host, version, insecure, timeoutSeconds)
		result.Supported[version] = supported
	}

	return result
}

func printSingleResult(result TLSResult, insecure bool) {
	if result.Error != nil {
		fmt.Printf("Errore: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Verifica delle versioni TLS supportate per: %s\n", result.URL)
	fmt.Printf("Host: %s\n", result.Host)
	if insecure {
		fmt.Println("Modalità: INSECURE (ignora errori certificato)")
	}
	fmt.Println()

	supportedList := []string{}
	unsupportedList := []string{}

	for _, version := range tlsVersionsToTest {
		versionName := tlsVersionNames[version]
		fmt.Printf("Tentativo con %s... ", versionName)

		if result.Supported[version] {
			fmt.Println("✓ Supportata")
			supportedList = append(supportedList, versionName)
		} else {
			fmt.Println("✗ Non supportata")
			unsupportedList = append(unsupportedList, versionName)
		}
	}

	fmt.Println("\n" + "======================")
	fmt.Println("RIEPILOGO")
	fmt.Println("======================")

	if len(supportedList) > 0 {
		fmt.Println("\nVersioni TLS supportate:")
		for _, v := range supportedList {
			fmt.Printf("  ✓ %s\n", v)
		}
	}

	if len(unsupportedList) > 0 {
		fmt.Println("\nVersioni TLS non supportate:")
		for _, v := range unsupportedList {
			fmt.Printf("  ✗ %s\n", v)
		}
	}
}

func testTLSVersion(host string, version uint16, insecure bool, timeoutSeconds int) bool {
	config := &tls.Config{
		MinVersion:         version,
		MaxVersion:         version,
		InsecureSkipVerify: insecure,
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		"tcp",
		host,
		config,
	)

	if err != nil {
		return false
	}

	defer conn.Close()

	err = conn.Handshake()
	return err == nil
}
