package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}

	privateKeyFile, err := os.Create("private_key.pem")
	if err != nil {
		log.Fatalf("Failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	err = pem.Encode(privateKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err != nil {
		log.Fatalf("Failed to write private key to file: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyFile, err := os.Create("public_key.pem")
	if err != nil {
		log.Fatalf("Failed to create public key file: %v", err)
	}
	defer publicKeyFile.Close()

	err = pem.Encode(publicKeyFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if err != nil {
		log.Fatalf("Failed to write public key to file: %v", err)
	}

	fmt.Println("RSA key pair generated successfully.")
	fmt.Println("Private key saved to: private_key.pem")
	fmt.Println("Public key saved to: public_key.pem")
}
