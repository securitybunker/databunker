package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"strings"
)

// stringPrefix marks BasicStringEncrypt2 values; BasicStringEncrypt output never contains a colon.
const stringPrefix = "v2:"

// shamir secret split
// https://github.com/hashicorp/vault/tree/master/shamir
// https://github.com/kinvolk/go-shamir
// go get github.com/hashicorp/vault/shamir

func GenerateRecordKey() ([]byte, error) {
	key := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// generate master key - 24 bytes length
func GenerateMasterKey() ([]byte, error) {
	masterKey := make([]byte, 24)
	_, err := io.ReadFull(rand.Reader, masterKey)
	return masterKey, err
}

func Decrypt(masterKey []byte, userKey []byte, data []byte) ([]byte, error) {
	// DO NOT USE THE FOLLOWING LINE. It is broken!!!
	//key := append(masterKey, userKey...)
	la := len(masterKey)
	key := make([]byte, la+len(userKey))
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

func Encrypt(masterKey []byte, userKey []byte, plaintext []byte) ([]byte, error) {
	// We use 32 byte key (AES-256).
	// comprising 24 master key
	// and 8 bytes record key
	la := len(masterKey)
	key := make([]byte, la+len(userKey))
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
	ciphertext := make([]byte, la+len(nonce))
	copy(ciphertext, ciphertext0)
	copy(ciphertext[la:], nonce)
	return ciphertext, nil
}

// Deprecated: use BasicStringEncrypt2.
func BasicStringEncrypt(plaintext string, masterKey []byte, code []byte) (string, error) {
	//log.Printf("Going to utils.Encrypt %s", plaintext)
	nonce := []byte("$DataBunker$")
	la := len(masterKey)
	key := make([]byte, la+len(code))
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

func BasicStringDecrypt(data string, masterKey []byte, code []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	nonce := []byte("$DataBunker$")
	la := len(masterKey)
	key := make([]byte, la+len(code))
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
	//log.Printf("utils.Decrypt result : %s", string(plaintext))
	return string(plaintext), err
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// stringCipher derives separate encryption and nonce keys, not shared with other functions.
func stringCipher(masterKey []byte, code []byte) (cipher.AEAD, []byte, error) {
	if len(masterKey) != 24 {
		return nil, nil, errors.New("bad master key length")
	}
	key := make([]byte, 0, len(masterKey)+len(code))
	key = append(key, masterKey...)
	key = append(key, code...)
	block, err := aes.NewCipher(hmacSHA256(key, "databunker v2 encryption"))
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	return aesgcm, hmacSHA256(key, "databunker v2 nonce"), err
}

// BasicStringEncrypt2 derives the nonce from the value: equal values match, so they can be looked up.
func BasicStringEncrypt2(plaintext string, masterKey []byte, code []byte) (string, error) {
	aesgcm, nonceKey, err := stringCipher(masterKey, code)
	if err != nil {
		return "", err
	}
	nonce := hmacSHA256(nonceKey, plaintext)[:aesgcm.NonceSize()]
	data := append([]byte{}, nonce...)
	data = aesgcm.Seal(data, nonce, []byte(plaintext), nil)
	return stringPrefix + base64.StdEncoding.EncodeToString(data), nil
}

// BasicStringDecrypt2 reads BasicStringEncrypt2 values and older BasicStringEncrypt ones.
func BasicStringDecrypt2(data string, masterKey []byte, code []byte) (string, error) {
	encoded, ok := strings.CutPrefix(data, stringPrefix)
	if !ok {
		return BasicStringDecrypt(data, masterKey, code)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	aesgcm, _, err := stringCipher(masterKey, code)
	if err != nil {
		return "", err
	}
	if len(raw) < aesgcm.NonceSize() {
		return "", errors.New("encrypted value is too short")
	}
	plaintext, err := aesgcm.Open(nil, raw[:aesgcm.NonceSize()], raw[aesgcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
