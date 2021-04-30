package aes256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func Encrypt(data []byte, secret []byte, iv []byte) (encryptedData []byte, err error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	encryptedData = aesgcm.Seal(nil, iv, data, nil)
	return
}

func Decrypt(ciphertext []byte, secret []byte, iv []byte) (decryptedData []byte, err error) {

	block, err := aes.NewCipher(secret)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	decryptedData, err = aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return
	}

	return
}

func GenerateSecret() (secret []byte, err error) {
	secret = make([]byte, 32)

	_, err = rand.Read(secret)
	return
}

func GenerateIV() (iv []byte, err error) {
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	//nonce := make([]byte, 12)
	//if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
	//	return
	//}
	iv = make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, iv)

	return
}
