package utils

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

var (
	cryptTestKey  = []byte("0123456789abcdefghijklmn")
	cryptTestCode = []byte("codecode")
)

func cryptRaw(t *testing.T, enc string) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(enc, stringPrefix))
	if err != nil {
		t.Fatalf("bad base64 %q: %v", enc, err)
	}
	return raw
}

func TestBasicString2RoundTrip(t *testing.T) {
	for _, plain := range []string{"", "a", "user@example.com", "Zoë 日本", "v2:looks-encrypted", strings.Repeat("x", 5000)} {
		enc, err := BasicStringEncrypt2(plain, cryptTestKey, cryptTestCode)
		if err != nil {
			t.Fatalf("encrypt %.20q: %v", plain, err)
		}
		again, _ := BasicStringEncrypt2(plain, cryptTestKey, cryptTestCode)
		if enc != again {
			t.Errorf("%.20q must encrypt to the same value every time", plain)
		}
		got, err := BasicStringDecrypt2(enc, cryptTestKey, cryptTestCode)
		if err != nil || got != plain {
			t.Errorf("%.20q: got %.20q, %v", plain, got, err)
		}
	}
}

// Pins the stored format: a change here would orphan saved data and lookups.
func TestBasicString2StableFormat(t *testing.T) {
	const want = "v2:GlHvtiN0uVffdJtU60k1YP3yYz+RxRJBUo3Cc8ATXQedhCaHzmKddodS28k="
	enc, _ := BasicStringEncrypt2("user@example.com", cryptTestKey, cryptTestCode)
	if enc != want {
		t.Fatalf("stored format changed: %s", enc)
	}
}

// Different values must not share a keystream.
func TestBasicString2Keystream(t *testing.T) {
	a, b := "known-user@example.com", "other-user@example.com"
	encA, _ := BasicStringEncrypt2(a, cryptTestKey, cryptTestCode)
	encB, _ := BasicStringEncrypt2(b, cryptTestKey, cryptTestCode)
	rawA, rawB := cryptRaw(t, encA)[12:], cryptRaw(t, encB)[12:]
	same := 0
	for i := range a {
		if rawA[i]^a[i] == rawB[i]^b[i] {
			same++
		}
	}
	if same > len(a)/2 {
		t.Fatalf("%d of %d keystream bytes are shared", same, len(a))
	}
}

// Nonces must differ per value, key and tenant code.
func TestBasicString2Nonces(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 20000; i++ {
		enc, _ := BasicStringEncrypt2(strings.Repeat("a", i%50)+string(rune('A'+i%26))+base64.StdEncoding.EncodeToString([]byte{byte(i), byte(i >> 8)}), cryptTestKey, cryptTestCode)
		seen[string(cryptRaw(t, enc)[:12])] = true
	}
	if len(seen) != 20000 {
		t.Errorf("expected 20000 distinct nonces, got %d", len(seen))
	}
	a, _ := BasicStringEncrypt2("same", cryptTestKey, cryptTestCode)
	b, _ := BasicStringEncrypt2("same", cryptTestKey, []byte("codecod2"))
	c, _ := BasicStringEncrypt2("same", []byte("0123456789abcdefghijklmX"), cryptTestCode)
	na, nb, nc := cryptRaw(t, a)[:12], cryptRaw(t, b)[:12], cryptRaw(t, c)[:12]
	if bytes.Equal(na, nb) || bytes.Equal(na, nc) || bytes.Equal(nb, nc) {
		t.Errorf("tenants or keys share a nonce")
	}
}

func TestBasicString2ReadsOldFormat(t *testing.T) {
	for _, plain := range []string{"", "user@example.com", "v2:x", strings.Repeat("y", 300)} {
		old, _ := BasicStringEncrypt(plain, cryptTestKey, cryptTestCode)
		if strings.HasPrefix(old, stringPrefix) {
			t.Fatalf("old values must never look like the new format")
		}
		got, err := BasicStringDecrypt2(old, cryptTestKey, cryptTestCode)
		if err != nil || got != plain {
			t.Errorf("old value %.20q: got %.20q, %v", plain, got, err)
		}
	}
}

func TestBasicString2RejectsBadInput(t *testing.T) {
	enc, _ := BasicStringEncrypt2("user@example.com", cryptTestKey, cryptTestCode)
	if _, err := BasicStringDecrypt2(enc, cryptTestKey, []byte("codecod2")); err == nil {
		t.Errorf("another tenant's code must not decrypt")
	}
	if _, err := BasicStringDecrypt2(enc, []byte("0123456789abcdefghijklmX"), cryptTestCode); err == nil {
		t.Errorf("another master key must not decrypt")
	}
	raw := cryptRaw(t, enc)
	for i := range raw {
		bad := append([]byte{}, raw...)
		bad[i] ^= 1
		if _, err := BasicStringDecrypt2(stringPrefix+base64.StdEncoding.EncodeToString(bad), cryptTestKey, cryptTestCode); err == nil {
			t.Errorf("changed byte %d accepted", i)
		}
	}
	for n := 0; n < len(raw); n++ {
		if _, err := BasicStringDecrypt2(stringPrefix+base64.StdEncoding.EncodeToString(raw[:n]), cryptTestKey, cryptTestCode); err == nil {
			t.Errorf("value cut to %d bytes accepted", n)
		}
	}
	for _, bad := range []string{"", "v2:", "v2:!!!", "v2", "!!!"} {
		if _, err := BasicStringDecrypt2(bad, cryptTestKey, cryptTestCode); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	for _, key := range [][]byte{nil, {}, []byte("short"), make([]byte, 32)} {
		if _, err := BasicStringEncrypt2("x", key, cryptTestCode); err == nil {
			t.Errorf("master key of %d bytes accepted", len(key))
		}
	}
}
