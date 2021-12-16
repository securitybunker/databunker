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
	// DO NOT USE THE FOLLOWING LINE. It is broken!!!
	//key := append(masterKey, userKey...)
	la := len(masterKey)
	key := make([]byte, la + len(userKey))
	copy(key, masterKey)
	copy(key[la:], userKey)

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
	//fmt.Printf("full key: %x, mkey %x, ukey: %x, data: %x\n", key, masterKey, userKey, data)
	//fmt.Printf("nonce: %x, ciphertext: %x\n", nonce, ciphertext)
	return plaintext, err
}

func encrypt(masterKey []byte, userKey []byte, plaintext []byte) ([]byte, error) {
	// We use 32 byte key (AES-256).
	// comprising 24 master key
	// and 8 bytes record key
	la := len(masterKey)
	key := make([]byte, la + len(userKey))
	copy(key, masterKey)
	copy(key[la:], userKey)

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
	ciphertext0 := aesgcm.Seal(nil, nonce, plaintext, nil)
	//fmt.Printf("%x\n", ciphertext)
	// apppend random nonce bvalue to the end
	//ciphertext := append(ciphertext0, nonce...)
	la = len(ciphertext0)
	ciphertext := make([]byte, la + len(nonce))
	copy(ciphertext, ciphertext0)
	copy(ciphertext[la:], nonce)
	return ciphertext, nil
}

func basicStringEncrypt(plaintext string, masterKey []byte, code []byte) (string, error) {
	//log.Printf("Going to encrypt %s", plaintext)
	nonce := []byte("$DataBunker$")
	la := len(masterKey)
	key := make([]byte, la + len(code))
	copy(key, masterKey)
	copy(key[la:], code)

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
	la := len(masterKey)
	key := make([]byte, la + len(code))
	copy(key, masterKey)
	copy(key[la:], code)
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
