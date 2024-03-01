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
	namePrivateFile := "oauth-private.key"
	namePublicFile := "oauth-public.key"
	pathFolder := "./storage/secret/"
	bitSize := 4096

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes := encodePublicKeyToPEM(&privateKey.PublicKey)
	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	err = writeKeyToFile(privateKeyBytes, namePrivateFile, pathFolder)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = writeKeyToFile(publicKeyBytes, namePublicFile, pathFolder)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pemKeys.Block
	privateBlock := pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   privateDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privateBlock)

	return privatePEM
}

// encodePublicKeyToPEM encodes Public Key from RSA to PEM format
func encodePublicKeyToPEM(publicKey *rsa.PublicKey) []byte {
	// Get ASN.1 DER format
	publicDER := x509.MarshalPKCS1PublicKey(publicKey)

	// pemKeys.Block
	publicBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicDER,
	}

	// Public key in PEM format
	publicPEM := pem.EncodeToMemory(&publicBlock)

	return publicPEM
}

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, fileName string, pathFolder string) error {
	err := os.MkdirAll(pathFolder, os.ModePerm)
	if err != nil {
		log.Printf("failed create folder")
		return err
	}

	fullPathKey := fmt.Sprintf("%s%s", pathFolder, fileName)
	err = os.WriteFile(fullPathKey, keyBytes, 0600)
	if err != nil {
		log.Printf("failed write file")
		return err
	}

	log.Printf("Key saved to: %s%s", pathFolder, fileName)
	return nil
}
