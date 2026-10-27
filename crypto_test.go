package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"
)

func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid compliant password",
			password: "Herasat@2026Secure",
			wantErr:  false,
		},
		{
			name:     "Valid with symbols and uppercase",
			password: "MySuperSecret!2026",
			wantErr:  false,
		},
		{
			name:     "Too short (< 12 chars)",
			password: "Short@1A",
			wantErr:  true,
		},
		{
			name:     "Missing uppercase letter",
			password: "alllowercasepassword!123",
			wantErr:  true,
		},
		{
			name:     "Missing symbol/special char",
			password: "NoSpecialCharPassword123",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordComplexity(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordComplexity(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestArgon2idDualFormatAndRehash(t *testing.T) {
	password := "Herasat@SecretPassword2026!"

	salt := []byte("1234567890123456")
	legacyHashBytes := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(legacyHashBytes)
	legacyHash := b64Salt + "." + b64Hash

	ok, needsRehash := VerifyPasswordWithRehash(password, legacyHash)
	if !ok {
		t.Fatalf("Expected legacy hash to verify successfully")
	}
	if !needsRehash {
		t.Fatalf("Expected needsRehash to be true for legacy hash")
	}
	if !VerifyPassword(password, legacyHash) {
		t.Fatalf("Expected VerifyPassword to return true for legacy hash")
	}

	phcHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to generate PHC password hash: %v", err)
	}
	if !strings.HasPrefix(phcHash, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Fatalf("Expected PHC hash to start with $argon2id$v=19$m=65536,t=3,p=4$, got %s", phcHash)
	}

	ok, needsRehash = VerifyPasswordWithRehash(password, phcHash)
	if !ok {
		t.Fatalf("Expected new PHC hash to verify successfully")
	}
	if needsRehash {
		t.Fatalf("Expected needsRehash to be false for up-to-date PHC hash")
	}

	badOk, _ := VerifyPasswordWithRehash("WrongPassword123!", phcHash)
	if badOk {
		t.Fatalf("Expected wrong password to fail verification")
	}
}

func TestHKDFKeySeparationAndEncryption(t *testing.T) {
	masterKey := "12345678901234567890123456789012"
	InitCrypto(masterKey)

	if bytes.Equal(keyFields, keyFilePath) || bytes.Equal(keyFields, keyFileBlob) || bytes.Equal(keyFilePath, keyFileBlob) {
		t.Fatalf("Subkeys must be distinct across cryptographic purposes")
	}
	if bytes.Equal(keyFields, aesKey) {
		t.Fatalf("Subkeys must not match master key")
	}

	plainText := "اطلاعات محرمانه گزارش تخلف اداری"
	encField, err := EncryptAES(plainText)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}
	if !strings.HasPrefix(encField, "v2:f:") {
		t.Fatalf("Expected v2:f: prefix, got %s", encField)
	}
	decField, err := DecryptAES(encField)
	if err != nil {
		t.Fatalf("DecryptAES failed: %v", err)
	}
	if decField != plainText {
		t.Fatalf("Expected %q, got %q", plainText, decField)
	}

	pathText := "uploads/2026/09/secret_evidence.pdf"
	encPath, err := EncryptFilePath(pathText)
	if err != nil {
		t.Fatalf("EncryptFilePath failed: %v", err)
	}
	if !strings.HasPrefix(encPath, "v2:p:") {
		t.Fatalf("Expected v2:p: prefix, got %s", encPath)
	}
	decPath, err := DecryptFilePath(encPath)
	if err != nil {
		t.Fatalf("DecryptFilePath failed: %v", err)
	}
	if decPath != pathText {
		t.Fatalf("Expected %q, got %q", pathText, decPath)
	}

	decPathAES, err := DecryptAES(encPath)
	if err != nil {
		t.Fatalf("DecryptAES failed on file path: %v", err)
	}
	if decPathAES != pathText {
		t.Fatalf("Expected %q, got %q", pathText, decPathAES)
	}

	fileBytes := []byte("Sensitive whistleblowing PDF evidence content \x00\xff\xfe")
	encBlob, err := EncryptFileBytes(fileBytes)
	if err != nil {
		t.Fatalf("EncryptFileBytes failed: %v", err)
	}
	if !bytes.HasPrefix(encBlob, []byte("HENC\x02")) {
		t.Fatalf("Expected HENC\\x02 prefix, got %q", encBlob[:5])
	}
	decBlob, err := DecryptFileBytes(encBlob)
	if err != nil {
		t.Fatalf("DecryptFileBytes failed: %v", err)
	}
	if !bytes.Equal(decBlob, fileBytes) {
		t.Fatalf("Decrypted file blob mismatch")
	}

	block, _ := aes.NewCipher(aesKey)
	aesgcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesgcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	legacyCiphertext := aesgcm.Seal(nonce, nonce, []byte(plainText), nil)
	v1Enc := "v1:" + base64.StdEncoding.EncodeToString(legacyCiphertext)

	decV1, err := DecryptAES(v1Enc)
	if err != nil {
		t.Fatalf("DecryptAES failed on legacy v1 ciphertext: %v", err)
	}
	if decV1 != plainText {
		t.Fatalf("Expected %q from v1 ciphertext, got %q", plainText, decV1)
	}

	header := []byte("HENC\x01")
	out := append(header, nonce...)
	legacyBlob := aesgcm.Seal(out, nonce, fileBytes, nil)

	decLegacyBlob, err := DecryptFileBytes(legacyBlob)
	if err != nil {
		t.Fatalf("DecryptFileBytes failed on legacy HENC\\x01: %v", err)
	}
	if !bytes.Equal(decLegacyBlob, fileBytes) {
		t.Fatalf("Decrypted legacy file blob mismatch")
	}

	rawPlain := "plain unencrypted description text"
	_, err = DecryptAES(rawPlain)
	if err == nil {
		t.Fatalf("Expected error for unencrypted text, got nil")
	}
}

func TestCryptoUnicodeEmojiNulBytesAndTamper(t *testing.T) {
	masterKey := "12345678901234567890123456789012"
	InitCrypto(masterKey)

	complexInput := "گزارش حراست: تخلف مالی 💰🚨 و جعل اسناد اداری \x00\x01\x02 پایان پیام"
	enc, err := EncryptAES(complexInput)
	if err != nil {
		t.Fatalf("EncryptAES failed for complex input: %v", err)
	}
	dec, err := DecryptAES(enc)
	if err != nil {
		t.Fatalf("DecryptAES failed for complex input: %v", err)
	}
	if dec != complexInput {
		t.Fatalf("Mismatch in decrypted complex input: expected %q, got %q", complexInput, dec)
	}

	parts := strings.Split(enc, ":")
	if len(parts) == 3 {
		payloadBytes, err := base64.StdEncoding.DecodeString(parts[2])
		if err == nil && len(payloadBytes) > 5 {
			payloadBytes[len(payloadBytes)-1] ^= 0xFF
			tamperedEnc := parts[0] + ":" + parts[1] + ":" + base64.StdEncoding.EncodeToString(payloadBytes)
			_, err = DecryptAES(tamperedEnc)
			if err == nil {
				t.Fatalf("Expected decryption failure on tampered ciphertext, but succeeded")
			}
		}
	}

	blob, err := EncryptFileBytes([]byte("Confidential Document Content"))
	if err != nil {
		t.Fatalf("EncryptFileBytes failed: %v", err)
	}
	if len(blob) > 20 {
		blob[len(blob)-1] ^= 0xFF
		_, err = DecryptFileBytes(blob)
		if err == nil {
			t.Fatalf("Expected error when decrypting tampered file blob, but succeeded")
		}
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	masterKey := []byte("12345678901234567890123456789012")
	InitCrypto(string(masterKey))

	keyA1 := deriveKey(masterKey, "TEST_PURPOSE_PURPOSE")
	keyA2 := deriveKey(masterKey, "TEST_PURPOSE_PURPOSE")
	if !bytes.Equal(keyA1, keyA2) {
		t.Fatalf("deriveKey must be deterministic for identical info strings")
	}

	keyB := deriveKey(masterKey, "DIFFERENT_PURPOSE")
	if bytes.Equal(keyA1, keyB) {
		t.Fatalf("deriveKey must produce different keys for different info strings")
	}
}

func TestCaptchaVerificationSuite(t *testing.T) {
	InitCrypto("12345678901234567890123456789012")
	db := getTestDB(t)
	UsedCaptchas.init(db)

	token := GenerateCaptchaToken(42, 2*time.Minute)
	if !VerifyCaptcha(token, "42") {
		t.Fatalf("Expected captcha verification to succeed for correct answer")
	}

	if VerifyCaptcha(token, "42") {
		t.Fatalf("Expected captcha token reuse to fail (replay protection)")
	}

	expiredToken := GenerateCaptchaToken(10, -5*time.Second)
	if VerifyCaptcha(expiredToken, "10") {
		t.Fatalf("Expected expired captcha token to be rejected")
	}

	parts := strings.Split(token, ".")
	if len(parts) == 2 {
		tamperedToken := parts[0] + ".tamperedSignature=="
		if VerifyCaptcha(tamperedToken, "42") {
			t.Fatalf("Expected tampered captcha signature to be rejected")
		}
	}

	token2 := GenerateCaptchaToken(99, 2*time.Minute)
	if VerifyCaptcha(token2, "100") {
		t.Fatalf("Expected wrong answer to be rejected")
	}

	token3 := GenerateCaptchaToken(15, 2*time.Minute)
	if VerifyCaptcha(token3, "abc") {
		t.Fatalf("Expected non-numeric answer to be rejected")
	}
}

func TestGenerateMathCaptchaSVGAndHandler(t *testing.T) {
	InitCrypto("12345678901234567890123456789012")

	token, dataURI := GenerateMathCaptchaSVG()
	if token == "" {
		t.Fatalf("Expected non-empty token from GenerateMathCaptchaSVG")
	}
	if !strings.HasPrefix(dataURI, "data:image/svg+xml;base64,") {
		t.Fatalf("Expected SVG data URI prefix, got %s", dataURI[:30])
	}

	req := httptest.NewRequest("GET", "/api/captcha", nil)
	rr := httptest.NewRecorder()
	GetCaptchaHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("Expected 200 from GetCaptchaHandler, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "captcha_id") || !strings.Contains(rr.Body.String(), "image_base64") {
		t.Fatalf("Expected JSON with captcha_id and image_base64, got: %s", rr.Body.String())
	}
}
