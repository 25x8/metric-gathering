package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	publicKey := &privateKey.PublicKey

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Small data",
			data: []byte("test message"),
		},
		{
			name: "Medium data",
			data: []byte("This is a longer test message to encrypt and decrypt using RSA keys"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.data) == 0 {
				t.Skip("RSA encryption doesn't support empty data")
			}

			encrypted, err := EncryptWithPublicKey(tc.data, publicKey)
			if err != nil {
				t.Fatalf("Failed to encrypt data: %v", err)
			}

			if len(encrypted) == 0 {
				t.Fatalf("Encrypted data is empty")
			}

			decrypted, err := DecryptWithPrivateKey(encrypted, privateKey)
			if err != nil {
				t.Fatalf("Failed to decrypt data: %v", err)
			}

			if string(decrypted) != string(tc.data) {
				t.Fatalf("Decrypted data doesn't match original. Got: %s, Want: %s", decrypted, tc.data)
			}
		})
	}
}

func TestLoadKeys(t *testing.T) {
	if _, err := os.Stat("../../private_key.pem"); os.IsNotExist(err) {
		t.Skip("Key files not found, skipping test")
	}

	publicKey, err := LoadPublicKey("../../public_key.pem")
	if err != nil {
		t.Fatalf("Failed to load public key: %v", err)
	}
	if publicKey == nil {
		t.Fatal("Loaded public key is nil")
	}

	privateKey, err := LoadPrivateKey("../../private_key.pem")
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}
	if privateKey == nil {
		t.Fatal("Loaded private key is nil")
	}

	testData := []byte("Test message for encryption with loaded keys")
	encrypted, err := EncryptWithPublicKey(testData, publicKey)
	if err != nil {
		t.Fatalf("Failed to encrypt with loaded public key: %v", err)
	}

	decrypted, err := DecryptWithPrivateKey(encrypted, privateKey)
	if err != nil {
		t.Fatalf("Failed to decrypt with loaded private key: %v", err)
	}

	if string(decrypted) != string(testData) {
		t.Fatalf("Decrypted data doesn't match original. Got: %s, Want: %s", decrypted, testData)
	}
}
