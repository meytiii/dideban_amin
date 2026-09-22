package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestCryptoFailClosedNeverReturnsPlaintext(t *testing.T) {
	masterKey := "12345678901234567890123456789012"
	InitCrypto(masterKey)

	plain := "Sensitive Whistleblower Report Data"
	enc, err := EncryptAES(plain)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}

	dec, err := DecryptAES(enc)
	if err != nil || dec != plain {
		t.Fatalf("DecryptAES valid case failed: %v, got %s", err, dec)
	}

	parts := strings.Split(enc, ":")
	if len(parts) >= 3 {
		payload, err := base64.StdEncoding.DecodeString(parts[len(parts)-1])
		if err != nil {
			t.Fatalf("base64 decode failed: %v", err)
		}
		payload[len(payload)-1] ^= 0x55
		parts[len(parts)-1] = base64.StdEncoding.EncodeToString(payload)
		corruptedEnc := strings.Join(parts, ":")

		corruptedDec, err := DecryptAES(corruptedEnc)
		if err == nil {
			t.Fatalf("expected error on tampered ciphertext, got nil")
		}
		if !errors.Is(err, ErrAuthenticationFailed) {
			t.Fatalf("expected ErrAuthenticationFailed, got: %v", err)
		}
		if corruptedDec != "" {
			t.Fatalf("decrypted value must be empty on failure, got: %q", corruptedDec)
		}
	}

	rawBogus := base64.StdEncoding.EncodeToString([]byte("totally random garbage bytes"))
	bogusDec, err := DecryptAES(rawBogus)
	if err == nil {
		t.Fatalf("expected error on bogus base64 input, got nil")
	}
	if !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("expected ErrAuthenticationFailed for bogus raw base64, got: %v", err)
	}
	if bogusDec != "" {
		t.Fatalf("bogus raw base64 should return empty string, got: %q", bogusDec)
	}

	filePayload := []byte("TOP_SECRET_EVIDENCE_ATTACHMENT_PAYLOAD")
	encBlob, err := EncryptFileBytes(filePayload)
	if err != nil {
		t.Fatalf("EncryptFileBytes failed: %v", err)
	}

	decBlob, err := DecryptFileBytes(encBlob)
	if err != nil || !bytes.Equal(decBlob, filePayload) {
		t.Fatalf("DecryptFileBytes failed on valid blob: %v", err)
	}

	corruptedBlob := make([]byte, len(encBlob))
	copy(corruptedBlob, encBlob)
	corruptedBlob[len(corruptedBlob)-1] ^= 0xAA

	tamperedRes, err := DecryptFileBytes(corruptedBlob)
	if err == nil {
		t.Fatalf("expected decryption error on tampered file blob, got nil")
	}
	if !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("expected ErrAuthenticationFailed on tampered file blob, got: %v", err)
	}
	if tamperedRes != nil {
		t.Fatalf("tampered file blob result must be nil on failure, got %d bytes", len(tamperedRes))
	}

	bogusFile := []byte("INVALID_MAGIC_HEADER_NOT_HENC_0000000000000000000000")
	bogusRes, err := DecryptFileBytes(bogusFile)
	if err == nil {
		t.Fatalf("expected error for bogus file prefix, got nil")
	}
	if !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("expected ErrAuthenticationFailed for non-HENC blob, got: %v", err)
	}
	if bogusRes != nil {
		t.Fatalf("bogus file blob must return nil bytes, got %d bytes", len(bogusRes))
	}
}
