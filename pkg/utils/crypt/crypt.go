package crypt

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func GeneratePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password due to error %w", err)
	}

	return string(hash), nil
}

func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GetSHA512(key, secret string) string {
	s := fmt.Sprintf("%s%s", key, secret)
	sha := sha512.Sum512([]byte(s))
	return fmt.Sprintf("%x", sha)
}

func EncryptWithPublicKey(data []byte, pK *rsa.PublicKey) (string, error) {
	hash := sha512.New()
	var chunkSize = pK.N.BitLen()/8 - 2*len(data) - 2

	var result []byte
	chunks := chunkBy[byte](data, chunkSize)

	for _, chunk := range chunks {
		ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pK, chunk, nil)
		if err != nil {
			return "", err
		}
		result = append(result, ciphertext...)
	}

	key := base64.StdEncoding.EncodeToString(result)
	replace := map[string]string{
		"+": "-",
		"/": "_",
	}
	key = strReplace(key, replace)
	key = strings.TrimRight(key, "=")

	return key, nil
}

func DecryptWithPrivateKey(key string, pK *rsa.PrivateKey) ([]byte, error) {

	if len(key)%4 != 0 {
		for i := 1; i < len(key)%4; i++ {
			key += "="
		}
	}

	replace := map[string]string{
		"-": "+",
		"_": "/",
	}
	key = strReplace(key, replace)
	keyByte, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return []byte{}, err
	}

	decKey := []byte("")
	hash := sha512.New()

	for _, chunk := range chunkBy[byte](keyByte, pK.N.BitLen()/8) {
		plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, pK, chunk, nil)
		if err != nil {
			return []byte{}, err
		}
		decKey = append(decKey, plaintext...)
	}

	return decKey, nil
}

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func strReplace(str string, replace map[string]string) string {
	if len(replace) == 0 || len(str) == 0 {
		return str
	}
	for from, to := range replace {
		str = strings.ReplaceAll(str, from, to)
	}
	return str
}
