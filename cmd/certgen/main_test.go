package main

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCertificateFiles(t *testing.T) {
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	if err := generateCertificateFiles("localhost:8080", certFile, keyFile); err != nil {
		t.Fatalf("generateCertificateFiles() error = %v", err)
	}

	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", certFile, err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", keyFile, err)
	}
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		t.Fatalf("X509KeyPair() error = %v", err)
	}
}
