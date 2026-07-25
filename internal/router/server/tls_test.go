package server

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelfSignedCertificatePEM(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSignedCertificatePEM("localhost:8080")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCertificatePEM() error = %v", err)
	}
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		t.Fatalf("X509KeyPair() error = %v", err)
	}
}

func TestTLSCertificateLoadsConfiguredFiles(t *testing.T) {
	certFile, keyFile := writeTestCertificateFiles(t)
	server := &Server{
		tlsCertFile: certFile,
		tlsKeyFile:  keyFile,
	}

	certificate, err := server.tlsCertificate()
	if err != nil {
		t.Fatalf("tlsCertificate() error = %v", err)
	}
	if len(certificate.Certificate) == 0 {
		t.Fatal("tlsCertificate() should load at least one certificate")
	}
}

func TestTLSCertificateGeneratesEphemeralCertificate(t *testing.T) {
	server := &Server{httpServer: &http.Server{Addr: "localhost:8080"}}

	certificate, err := server.tlsCertificate()
	if err != nil {
		t.Fatalf("tlsCertificate() error = %v", err)
	}
	if len(certificate.Certificate) == 0 {
		t.Fatal("tlsCertificate() should generate at least one certificate")
	}
}

func TestTLSCertificateRejectsIncompleteFilePair(t *testing.T) {
	server := &Server{tlsCertFile: "cert.pem"}

	if _, err := server.tlsCertificate(); err == nil {
		t.Fatal("tlsCertificate() should reject incomplete cert/key file pair")
	}
}

func writeTestCertificateFiles(t *testing.T) (string, string) {
	t.Helper()

	certPEM, keyPEM, err := GenerateSelfSignedCertificatePEM("localhost:8080")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCertificatePEM() error = %v", err)
	}
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")
	if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", certFile, err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", keyFile, err)
	}
	return certFile, keyFile
}
