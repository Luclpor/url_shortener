package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Luclpor/url_shortener.git/internal/router/server"
)

func main() {
	addr := flag.String("a", "localhost:8080", "server address or host for certificate SAN")
	certFile := flag.String("tls-cert-file", "cert.pem", "TLS public certificate file path")
	keyFile := flag.String("tls-key-file", "key.pem", "TLS private key file path")
	flag.Parse()

	if err := generateCertificateFiles(*addr, *certFile, *keyFile); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "Certificate written to %s\nPrivate key written to %s\n", *certFile, *keyFile)
}

func generateCertificateFiles(addr, certFile, keyFile string) error {
	certificatePEM, privateKeyPEM, err := server.GenerateSelfSignedCertificatePEM(addr)
	if err != nil {
		return err
	}
	if err := os.WriteFile(certFile, certificatePEM, 0o644); err != nil {
		return err
	}
	return os.WriteFile(keyFile, privateKeyPEM, 0o600)
}
