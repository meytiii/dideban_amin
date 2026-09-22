package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"
)

func SanitizeJPEG(data []byte) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid jpeg image: %w", err)
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 88})
	if err != nil {
		return nil, fmt.Errorf("failed to encode sanitized jpeg: %w", err)
	}

	return buf.Bytes(), nil
}

func SanitizePNG(data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid png image: %w", err)
	}

	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	err = enc.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sanitized png: %w", err)
	}

	return buf.Bytes(), nil
}

func SanitizeAttachment(filename, mimeType string, data []byte) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".jpg" || ext == ".jpeg" || mimeType == "image/jpeg" {
		return SanitizeJPEG(data)
	}

	if ext == ".png" || mimeType == "image/png" {
		return SanitizePNG(data)
	}

	return data, nil
}
