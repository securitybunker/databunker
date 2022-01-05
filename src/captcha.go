package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/afocus/captcha"
	"github.com/gobuffalo/packr"
	"github.com/julienschmidt/httprouter"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"time"
)

var (
	comic        []byte
	captchaKey   = make([]byte, 16)
	regexCaptcha = regexp.MustCompile("^([a-zA-Z0-9]+):([0-9]+)$")
)

func (e mainEnv) showCaptcha(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := ps.ByName("code")
	if len(code) == 0 {
		err := errors.New("Bad code")
		returnError(w, r, "bad code", 405, err, nil)
		return
	}
	s, err := decryptCaptcha(code)
	if err != nil {
		returnError(w, r, err.Error(), 405, err, nil)
		return
	}
	log.Printf("Decoded captcha: %s", s)
	//box := packr.NewBox("../ui")
	//comic, err := box.Find("site/fonts/comic.ttf")
	//if err != nil {
	//  returnError(w, r, err.Error(), 405, err, nil)
	//  return
	//}
	cap := captcha.New()
	cap.SetSize(128, 64)
	cap.AddFontFromBytes(comic)
	img := cap.CreateCustom(s)
	w.WriteHeader(200)
	png.Encode(w, img)
}

func initCaptcha(h [16]byte) {
	var err error
	copy(captchaKey[:], h[:])
	box := packr.NewBox("../ui")
	comic, err = box.Find("site/fonts/comic.ttf")
	if err != nil {
		log.Fatalf("Failed to load font")
		return
	}
	//captchaKey = h
}

func generateCaptcha() (string, error) {
	code := randNum(6)
	//log.Printf("Generate captcha code: %d", code)
	now := int32(time.Now().Unix())
	s := fmt.Sprintf("%d:%d", code, now)
	plaintext := []byte(s)
	//log.Printf("Going to encrypt %s", plaintext)
	nonce := []byte("$DataBunker$")
	block, err := aes.NewCipher(captchaKey)
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
	result := hex.EncodeToString(ciphertext)
	//log.Printf("Encoded captcha: %s", result)
	//log.Printf("ciphertext : %s", result)
	return result, nil
}

func decryptCaptcha(data string) (string, error) {
	if len(data) > 100 {
		return "", errors.New("Ciphertext too long")
	}
	ciphertext, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	nonce := []byte("$DataBunker$")
	block, err := aes.NewCipher(captchaKey)
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
	match := regexCaptcha.FindStringSubmatch(string(plaintext))
	if len(match) != 3 {
		return "", errors.New("Failed to parse captcha")
	}
	code := match[1]
	t := atoi(match[2])
	// check if time expired
	now := int32(time.Now().Unix())
	if now > (t + 120) {
		return "", errors.New("Captcha expired")
	}
	if t > now {
		return "", errors.New("Captcha from the future")
	}
	return code, nil
}
