package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func GenerateRSA() (string, string, error) {
	filename := "key"
	bitSize := 4096

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return "", "", err
	}

	// Extract public component.
	pub := key.Public()

	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)
	keyFile, err := os.Create(filename + ".rsa")
	pubFile, err := os.Create(filename + ".rsa.pub")

	// Write private key to file.
	if err := os.WriteFile(filename+".rsa", keyPEM, 0700); err != nil {
		return "", "", err
	}

	// Write public key to file.
	if err := os.WriteFile(filename+".rsa.pub", pubPEM, 0755); err != nil {
		return "", "", err
	}
	return keyFile.Name(), pubFile.Name(), nil
}
