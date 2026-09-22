package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"
	"unicode/utf8"
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

func ScanPDFForActiveContent(data []byte) error {
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return errors.New("فایل فاقد امضای معتبر PDF است")
	}

	dangerousPatterns := [][]byte{
		[]byte("/JavaScript"),
		[]byte("/JS"),
		[]byte("/Launch"),
		[]byte("/EmbeddedFiles"),
		[]byte("/SubmitForm"),
		[]byte("/ImportData"),
		[]byte("/RichMedia"),
	}

	lowerData := bytes.ToLower(data)
	for _, p := range dangerousPatterns {
		if bytes.Contains(data, p) || bytes.Contains(lowerData, bytes.ToLower(p)) {
			return fmt.Errorf("فایل PDF حاوی عنصر اجرایی یا اسکریپت فعال است (%s)", string(p))
		}
	}
	return nil
}

func ValidateBinaryMagicBytes(buff []byte, filename string) (bool, string) {
	if len(buff) < 4 {
		return false, ""
	}

	ext := strings.ToLower(filepath.Ext(filename))

	if bytes.HasPrefix(buff, []byte("%PDF-")) && ext == ".pdf" {
		return true, "application/pdf"
	}

	if len(buff) >= 8 && bytes.HasPrefix(buff, []byte("\x89PNG\r\n\x1a\n")) && ext == ".png" {
		return true, "image/png"
	}

	if len(buff) >= 3 && buff[0] == 0xFF && buff[1] == 0xD8 && buff[2] == 0xFF && (ext == ".jpg" || ext == ".jpeg") {
		return true, "image/jpeg"
	}

	if len(buff) >= 4 && buff[0] == 'P' && buff[1] == 'K' && (buff[2] == 3 || buff[2] == 5 || buff[2] == 7) {
		if ext == ".zip" {
			return true, "application/zip"
		}
		if ext == ".docx" {
			return true, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		}
		if ext == ".xlsx" {
			return true, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		}
	}

	if bytes.HasPrefix(buff, []byte("Rar!\x1a\x07")) && ext == ".rar" {
		return true, "application/x-rar-compressed"
	}

	if ext == ".txt" && utf8.Valid(buff) {
		for _, b := range buff {
			if b < 0x09 || (b > 0x0D && b < 0x20) {
				return false, ""
			}
		}
		return true, "text/plain; charset=utf-8"
	}

	if len(buff) >= 3 && bytes.HasPrefix(buff, []byte("ID3")) && ext == ".mp3" {
		return true, "audio/mpeg"
	}
	if len(buff) >= 2 && buff[0] == 0xFF && (buff[1]&0xE0) == 0xE0 && ext == ".mp3" {
		return true, "audio/mpeg"
	}

	if len(buff) >= 4 && bytes.HasPrefix(buff, []byte("OggS")) && ext == ".ogg" {
		return true, "audio/ogg"
	}

	if len(buff) >= 12 && bytes.HasPrefix(buff, []byte("RIFF")) && bytes.Equal(buff[8:12], []byte("WAVE")) && ext == ".wav" {
		return true, "audio/wav"
	}

	if len(buff) >= 12 && bytes.Equal(buff[4:8], []byte("ftyp")) && ext == ".mp4" {
		return true, "video/mp4"
	}

	return false, ""
}
