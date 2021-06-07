package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
)

// shamir secret split
// https://github.com/hashicorp/vault/tree/master/shamir
// https://github.com/kinvolk/go-shamir
// go get github.com/hashicorp/vault/shamir

func generateRecordKey() ([]byte, error) {
	key := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// generate master key - 24 bytes length
func generateMasterKey() ([]byte, error) {
	masterKey := make([]byte, 24)
	_, err := io.ReadFull(rand.Reader, masterKey)
	return masterKey, err
}

func decrypt(masterKey []byte, userKey []byte, data []byte) ([]byte, error) {
	// Load your secret key from a safe place and reuse it across multiple
	// Seal/Open calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	key := append(masterKey, userKey...)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := data[0 : len(data)-12]
	nonce := data[len(data)-12:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	return plaintext, err
}

func encrypt(masterKey []byte, userKey []byte, plaintext []byte) ([]byte, error) {
	// We use 32 byte key (AES-256).
	// comprising 24 master key
	// and 8 bytes record key
	key := append(masterKey, userKey...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	//fmt.Printf("%x\n", ciphertext)
	// apppend random nonce bvalue to the end
	ciphertext = append(ciphertext, nonce...)
	return ciphertext, nil
}

func basicStringEncrypt(plaintext string, masterKey []byte, code []byte) (string, error) {
	//log.Printf("Going to encrypt %s", plaintext)
	nonce := []byte("$DataBunker$")
	key := append(masterKey, code...)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("error in aes.NewCipher %s", err)
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("error in cipher.NewGCM: %s", err)
		return "", err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)
	result := base64.StdEncoding.EncodeToString(ciphertext)
	//log.Printf("ciphertext : %s", result)
	return result, nil
}

func basicStringDecrypt(data string, masterKey []byte, code []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	nonce := []byte("$DataBunker$")
	key := append(masterKey, code...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	//log.Printf("decrypt result : %s", string(plaintext))
	return string(plaintext), err
}
