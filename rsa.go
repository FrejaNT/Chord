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

// sets up tls certificate for a node
func (n *Node) loadTLS() {
	key, pub, err := generateRSA()
	if err != nil {
		log.Fatal(err.Error())
	}
	cert, err := tls.LoadX509KeyPair(pub, key)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	n.tls = tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}

func generateRSA() (string, string, error) {
	filename := "key"
	bitSize := 4096

	// Generate key
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return "", "", err
	}

	// Create certificate (we use )
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"burpie"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}

	// encode
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	certPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		},
	)

	// write
	if err := os.WriteFile(filename+".rsa", keyPEM, 0700); err != nil {
		return "", "", err
	}

	if err := os.WriteFile(filename+".rsa.pub", certPEM, 0755); err != nil {
		return "", "", err
	}

	return filename + ".rsa", filename + ".rsa.pub", nil
}
