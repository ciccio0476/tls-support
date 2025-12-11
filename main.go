package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
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

func main() {
	// Define flags
	insecure := flag.Bool("insecure", false, "Ignora gli errori di verifica del certificato SSL/TLS")
	flag.Parse()

	// Get the URL from remaining arguments
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Uso: go run main.go [opzioni] <URL>")
		fmt.Println("\nOpzioni:")
		flag.PrintDefaults()
		fmt.Println("\nEsempio: go run main.go https://www.google.com")
		fmt.Println("Esempio: go run main.go --insecure https://self-signed.example.com")
		os.Exit(1)
	}

	targetURL := args[0]

	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		fmt.Printf("Errore durante il parsing dell'URL: %v\n", err)
		os.Exit(1)
	}

	if parsedURL.Scheme != "https" {
		fmt.Println("Errore: L'URL deve utilizzare HTTPS")
		os.Exit(1)
	}

	// Get host and port
	host := parsedURL.Host
	if parsedURL.Port() == "" {
		host = host + ":443"
	}

	fmt.Printf("Verifica delle versioni TLS supportate per: %s\n", targetURL)
	fmt.Printf("Host: %s\n", host)
	if *insecure {
		fmt.Println("Modalità: INSECURE (ignora errori certificato)")
	}
	fmt.Println()

	supported := []string{}
	unsupported := []string{}

	// Test each TLS version
	for _, version := range tlsVersionsToTest {
		versionName := tlsVersionNames[version]
		fmt.Printf("Tentativo con %s... ", versionName)

		if testTLSVersion(host, version, *insecure) {
			fmt.Println("✓ Supportata")
			supported = append(supported, versionName)
		} else {
			fmt.Println("✗ Non supportata")
			unsupported = append(unsupported, versionName)
		}
	}

	// Summary
	fmt.Println("\n" + "======================")
	fmt.Println("RIEPILOGO")
	fmt.Println("======================")

	if len(supported) > 0 {
		fmt.Println("\nVersioni TLS supportate:")
		for _, v := range supported {
			fmt.Printf("  ✓ %s\n", v)
		}
	}

	if len(unsupported) > 0 {
		fmt.Println("\nVersioni TLS non supportate:")
		for _, v := range unsupported {
			fmt.Printf("  ✗ %s\n", v)
		}
	}
}

func testTLSVersion(host string, version uint16, insecure bool) bool {
	config := &tls.Config{
		MinVersion:         version,
		MaxVersion:         version,
		InsecureSkipVerify: insecure,
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: 5 * time.Second,
		},
		"tcp",
		host,
		config,
	)

	if err != nil {
		return false
	}

	defer conn.Close()

	// Perform handshake
	err = conn.Handshake()
	if err != nil {
		return false
	}

	return true
}
