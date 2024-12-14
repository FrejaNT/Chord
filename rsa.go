package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func (n *Node) LoadTLS() {
	key, pub, err := GenerateRSA()
	if err != nil {
		log.Fatal(err.Error())
	}
	cert, err := tls.LoadX509KeyPair(pub, key)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	n.tls = tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}

func GenerateRSA() (string, string, error) {
	filename := "key"
	bitSize := 4096

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return "", "", err
	}

	// Create a self-signed certificate template.
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"burpie"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year validity
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create the certificate.
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode the certificate to PEM format.
	certPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		},
	)

	// Write the private key to a file.
	if err := os.WriteFile(filename+".rsa", keyPEM, 0700); err != nil {
		return "", "", err
	}

	// Write the certificate to a file.
	if err := os.WriteFile(filename+".rsa.pub", certPEM, 0755); err != nil {
		return "", "", err
	}

	return filename + ".rsa", filename + ".rsa.pub", nil
}
