package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
)

const (
	ivSize = 16
)

func generateIV() []byte {
	asciiChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	iv := make([]byte, ivSize)
	for i := 0; i < ivSize; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(asciiChars))))
		iv[i] = asciiChars[num.Int64()]
	}
	return iv
}

// PKCS7 Padding
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7 Unpadding
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding value")
	}
	for _, v := range data[len(data)-padding:] {
		if int(v) != padding {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:len(data)-padding], nil
}

func Encrypt(plain string) (string, error) {
	key := os.Getenv("SECURITY_KEY")
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// Generate ASCII IV
	iv := generateIV()

	// CBC mode encrypter
	mode := cipher.NewCBCEncrypter(block, iv)

	// Pad plaintext to block size
	padded := pkcs7Pad([]byte(plain), aes.BlockSize)
	encrypted := make([]byte, len(padded))
	mode.CryptBlocks(encrypted, padded)

	// Combine IV + ciphertext
	combined := append(iv, encrypted...)
	return hex.EncodeToString(combined), nil
}

func Decrypt(cipherHex string) (string, error) {
	key := os.Getenv("SECURITY_KEY")
	combined, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	if len(combined) < ivSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := combined[:ivSize]
	encrypted := combined[ivSize:]

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Unpad
	unpadded, err := pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return "", err
	}
	return string(unpadded), nil
}
