package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func createTestJPEGWithEXIF() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 7), G: uint8(y * 7), B: 120, A: 255})
		}
	}

	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	encoded := buf.Bytes()

	fakeExifApp1 := []byte{
		0xFF, 0xE1, 0x00, 0x1A,
		'E', 'x', 'i', 'f', 0x00, 0x00,
		'I', 'I', 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x25, 0x88, 0x04, 0x00, 0x01, 0x00,
		0x00, 0x00, 0x1A, 0x00,
	}

	if len(encoded) > 2 && encoded[0] == 0xFF && encoded[1] == 0xD8 {
		res := make([]byte, 0, len(encoded)+len(fakeExifApp1))
		res = append(res, encoded[:2]...)
		res = append(res, fakeExifApp1...)
		res = append(res, encoded[2:]...)
		return res
	}
	return encoded
}

func createTestPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 10), G: 150, B: uint8(y * 10), A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestMetadataSanitizationJPEG(t *testing.T) {
	rawJpeg := createTestJPEGWithEXIF()
	if !bytes.Contains(rawJpeg, []byte("Exif")) {
		t.Fatalf("test JPEG setup failed: missing Exif marker in raw test data")
	}

	sanitized, err := SanitizeJPEG(rawJpeg)
	if err != nil {
		t.Fatalf("SanitizeJPEG failed: %v", err)
	}

	if bytes.Contains(sanitized, []byte("Exif")) {
		t.Fatalf("SanitizeJPEG failed: output still contains Exif metadata")
	}

	img, err := jpeg.Decode(bytes.NewReader(sanitized))
	if err != nil {
		t.Fatalf("Sanitized JPEG is not valid: %v", err)
	}
	if img.Bounds().Dx() != 32 || img.Bounds().Dy() != 32 {
		t.Fatalf("Sanitized JPEG dimensions mismatch: %v", img.Bounds())
	}

	rawHash := sha256.Sum256(rawJpeg)
	sanitizedHash := sha256.Sum256(sanitized)
	if hex.EncodeToString(rawHash[:]) == hex.EncodeToString(sanitizedHash[:]) {
		t.Fatalf("Sanitized payload hash must differ from un-sanitized payload with metadata")
	}
}

func TestMetadataSanitizationPNG(t *testing.T) {
	rawPng := createTestPNG()
	sanitized, err := SanitizePNG(rawPng)
	if err != nil {
		t.Fatalf("SanitizePNG failed: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(sanitized))
	if err != nil {
		t.Fatalf("Sanitized PNG is invalid: %v", err)
	}
	if img.Bounds().Dx() != 24 || img.Bounds().Dy() != 24 {
		t.Fatalf("Sanitized PNG dimensions mismatch: %v", img.Bounds())
	}
}

func TestSanitizeAttachmentPassthrough(t *testing.T) {
	docPayload := []byte("%PDF-1.4 header and content")
	sanitized, err := SanitizeAttachment("report.pdf", "application/pdf", docPayload)
	if err != nil {
		t.Fatalf("SanitizeAttachment on pdf failed: %v", err)
	}
	if !bytes.Equal(sanitized, docPayload) {
		t.Fatalf("Non-image payload should be passed through unmodified")
	}
}
