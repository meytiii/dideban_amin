package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	aesKey        []byte // Master key
	keyFields     []byte // Subkey: HERASAT_REPORT_FIELDS_V2
	keyFilePath   []byte // Subkey: HERASAT_FILE_PATH_V2
	keyFileBlob   []byte // Subkey: HERASAT_FILE_BLOB_V2
	hmacKey       []byte // Subkey: HERASAT_CAPTCHA_HMAC_V2
	legacyHmacKey []byte // Legacy HMAC key: HERASAT_CAPTCHA_HMAC_V1
)

// deriveKey implements RFC 5869 HKDF-SHA256 to derive a 32-byte key for a specific purpose.
func deriveKey(masterKey []byte, info string) []byte {
	// HKDF-Extract: PRK = HMAC-Hash(salt, masterKey)
	salt := make([]byte, sha256.Size) // RFC 5869: default salt is HashLen zeros
	extractMac := hmac.New(sha256.New, salt)
	extractMac.Write(masterKey)
	prk := extractMac.Sum(nil)

	// HKDF-Expand: T(1) = HMAC-Hash(PRK, info | 0x01)
	expandMac := hmac.New(sha256.New, prk)
	expandMac.Write([]byte(info))
	expandMac.Write([]byte{0x01})
	return expandMac.Sum(nil)
}

func InitCrypto(key string) {
	if len(key) != 32 {
		panic("AES key must be exactly 32 bytes for AES-256")
	}
	aesKey = []byte(key)

	keyFields = deriveKey(aesKey, "HERASAT_REPORT_FIELDS_V2")
	keyFilePath = deriveKey(aesKey, "HERASAT_FILE_PATH_V2")
	keyFileBlob = deriveKey(aesKey, "HERASAT_FILE_BLOB_V2")
	hmacKey = deriveKey(aesKey, "HERASAT_CAPTCHA_HMAC_V2")

	hLegacy := sha256.Sum256(append(aesKey, []byte("HERASAT_CAPTCHA_HMAC_V1")...))
	legacyHmacKey = hLegacy[:]
}

func GetHMACKey() []byte {
	return hmacKey
}

func GetLegacyHMACKey() []byte {
	return legacyHmacKey
}

func EncryptAES(plaintext string) (string, error) {
	return encryptWithKey(keyFields, "v2:f:", plaintext)
}

func EncryptFilePath(plaintext string) (string, error) {
	return encryptWithKey(keyFilePath, "v2:p:", plaintext)
}

func encryptWithKey(key []byte, prefix, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return prefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

var (
	ErrAuthenticationFailed = errors.New("cryptographic authentication failure")
	ErrInvalidCiphertext   = errors.New("invalid ciphertext format")
)

func DecryptAES(rawCiphertext string) (string, error) {
	if rawCiphertext == "" {
		return "", nil
	}

	if strings.HasPrefix(rawCiphertext, "v2:p:") {
		res, err := decryptWithKey(keyFilePath, strings.TrimPrefix(rawCiphertext, "v2:p:"))
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
		}
		return res, nil
	}

	if strings.HasPrefix(rawCiphertext, "v2:f:") {
		res, err := decryptWithKey(keyFields, strings.TrimPrefix(rawCiphertext, "v2:f:"))
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
		}
		return res, nil
	}
	if strings.HasPrefix(rawCiphertext, "v2:") {
		res, err := decryptWithKey(keyFields, strings.TrimPrefix(rawCiphertext, "v2:"))
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
		}
		return res, nil
	}

	if strings.HasPrefix(rawCiphertext, "v1:") {
		res, err := decryptWithKey(aesKey, strings.TrimPrefix(rawCiphertext, "v1:"))
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
		}
		return res, nil
	}

	decrypted, err := decryptWithKey(aesKey, rawCiphertext)
	if err == nil {
		return decrypted, nil
	}

	return "", ErrAuthenticationFailed
}

func DecryptFilePath(rawCiphertext string) (string, error) {
	return DecryptAES(rawCiphertext)
}

func decryptWithKey(key []byte, b64Ciphertext string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(b64Ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func EncryptFileBytes(rawBytes []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyFileBlob)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	header := []byte("HENC\x02")
	out := append(header, nonce...)
	encrypted := aesgcm.Seal(out, nonce, rawBytes, nil)
	return encrypted, nil
}

func DecryptFileBytes(encryptedBytes []byte) ([]byte, error) {
	if len(encryptedBytes) == 0 {
		return nil, ErrAuthenticationFailed
	}
	if bytes.HasPrefix(encryptedBytes, []byte("HENC\x02")) {
		return decryptFilePayload(keyFileBlob, encryptedBytes, 5)
	}
	if bytes.HasPrefix(encryptedBytes, []byte("HENC\x01")) {
		return decryptFilePayload(aesKey, encryptedBytes, 5)
	}
	return nil, ErrAuthenticationFailed
}

func decryptFilePayload(key []byte, encryptedBytes []byte, headerLen int) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(encryptedBytes) < headerLen+nonceSize {
		return nil, errors.New("encrypted file payload is too short")
	}

	nonce := encryptedBytes[headerLen : headerLen+nonceSize]
	ciphertext := encryptedBytes[headerLen+nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	return plaintext, nil
}

const (
	CurrentArgon2TimeCost   = uint32(3)
	CurrentArgon2MemoryCost = uint32(64 * 1024)
	CurrentArgon2Threads    = uint8(4)
	CurrentArgon2KeyLen     = uint32(32)
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, CurrentArgon2TimeCost, CurrentArgon2MemoryCost, CurrentArgon2Threads, CurrentArgon2KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		CurrentArgon2MemoryCost, CurrentArgon2TimeCost, CurrentArgon2Threads, b64Salt, b64Hash), nil
}

func VerifyPasswordWithRehash(password, encodedHash string) (ok bool, needsRehash bool) {
	if strings.HasPrefix(encodedHash, "$argon2id$") {
		parts := strings.Split(encodedHash, "$")
		if len(parts) != 6 {
			return false, false
		}

		var m, t uint32
		var p uint8
		if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
			return false, false
		}

		salt, err := base64.RawStdEncoding.DecodeString(parts[4])
		if err != nil {
			return false, false
		}

		hash, err := base64.RawStdEncoding.DecodeString(parts[5])
		if err != nil {
			return false, false
		}

		computedHash := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(hash)))
		if subtle.ConstantTimeCompare(hash, computedHash) != 1 {
			return false, false
		}

		needsRehash = (t < CurrentArgon2TimeCost || m < CurrentArgon2MemoryCost || p < CurrentArgon2Threads)
		return true, needsRehash
	}

	// Legacy format: salt.hash (base64 raw-std, t=1)
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		return false, false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, false
	}

	legacyTimeCost := uint32(1)
	legacyMemoryCost := uint32(64 * 1024)
	legacyThreads := uint8(4)
	legacyKeyLen := uint32(32)

	computedHash := argon2.IDKey([]byte(password), salt, legacyTimeCost, legacyMemoryCost, legacyThreads, legacyKeyLen)
	if subtle.ConstantTimeCompare(hash, computedHash) != 1 {
		return false, false
	}

	return true, true
}

func VerifyPassword(password, encodedHash string) bool {
	ok, _ := VerifyPasswordWithRehash(password, encodedHash)
	return ok
}

func ValidatePasswordComplexity(password string) error {
	if len(password) < 12 {
		return errors.New("رمز عبور باید حداقل ۱۲ کاراکتر باشد.")
	}

	hasUpper := false
	hasSymbol := false

	for _, r := range password {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}
		if strings.ContainsRune("!@#$%^&*()_+-=[]{};':\"|,.<>/?`~\\", r) {
			hasSymbol = true
		}
	}

	if !hasUpper {
		return errors.New("رمز عبور باید شامل حداقل یک حرف بزرگ انگلیسی (A-Z) باشد.")
	}
	if !hasSymbol {
		return errors.New("رمز عبور باید شامل حداقل یک نماد یا کاراکتر خاص (!@#$%^&*...) باشد.")
	}

	return nil
}
