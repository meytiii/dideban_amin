package main

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	maxImagePixels    = 40000000
	maxImageDimension = 10000
)

func SanitizeJPEG(data []byte) ([]byte, error) {
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid jpeg image: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxImageDimension || cfg.Height > maxImageDimension || int64(cfg.Width)*int64(cfg.Height) > maxImagePixels {
		return nil, errors.New("ابعاد تصویر بیش از حد مجاز است (Decompression Bomb Protection)")
	}

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
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid png image: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxImageDimension || cfg.Height > maxImageDimension || int64(cfg.Width)*int64(cfg.Height) > maxImagePixels {
		return nil, errors.New("ابعاد تصویر بیش از حد مجاز است (Decompression Bomb Protection)")
	}

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

func decodePDFHexNames(data []byte) []byte {
	var sb bytes.Buffer
	i := 0
	for i < len(data) {
		if data[i] == '#' && i+2 < len(data) {
			h := data[i+1 : i+3]
			decoded, err := hex.DecodeString(string(h))
			if err == nil && len(decoded) == 1 {
				sb.WriteByte(decoded[0])
				i += 3
				continue
			}
		}
		sb.WriteByte(data[i])
		i++
	}
	return sb.Bytes()
}

var dangerousPDFKeywords = [][]byte{
	[]byte("/JavaScript"),
	[]byte("/JS"),
	[]byte("/Launch"),
	[]byte("/EmbeddedFiles"),
	[]byte("/EmbeddedFile"),
	[]byte("/OpenAction"),
	[]byte("/AA"),
	[]byte("/SubmitForm"),
	[]byte("/ImportData"),
	[]byte("/RichMedia"),
	[]byte("/GoToR"),
}

func scanBufferForPDFKeywords(buf []byte) error {
	norm := decodePDFHexNames(buf)
	lower := bytes.ToLower(buf)
	normLower := bytes.ToLower(norm)

	for _, kw := range dangerousPDFKeywords {
		kwLower := bytes.ToLower(kw)
		if bytes.Contains(buf, kw) || bytes.Contains(lower, kwLower) ||
			bytes.Contains(norm, kw) || bytes.Contains(normLower, kwLower) {
			return fmt.Errorf("فایل PDF حاوی عنصر اجرایی یا اسکریپت فعال است (%s)", string(kw))
		}
	}
	return nil
}

func ScanPDFForActiveContent(data []byte) error {
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return errors.New("فایل فاقد امضای معتبر PDF است")
	}

	if bytes.Contains(data, []byte("/Encrypt")) {
		return errors.New("فایل‌های PDF رمزگذاری‌شده یا دارای کلمه عبور به دلایل امنیتی غیرمجاز هستند")
	}

	if err := scanBufferForPDFKeywords(data); err != nil {
		return err
	}

	streamMarker := []byte("stream")
	endStreamMarker := []byte("endstream")

	idx := 0
	for {
		sIdx := bytes.Index(data[idx:], streamMarker)
		if sIdx == -1 {
			break
		}
		streamStart := idx + sIdx + len(streamMarker)
		if streamStart < len(data) && data[streamStart] == '\r' {
			streamStart++
		}
		if streamStart < len(data) && data[streamStart] == '\n' {
			streamStart++
		}

		eIdx := bytes.Index(data[streamStart:], endStreamMarker)
		if eIdx == -1 {
			break
		}
		streamEnd := streamStart + eIdx

		headerStart := idx + sIdx - 256
		if headerStart < 0 {
			headerStart = 0
		}
		header := data[headerStart : idx+sIdx]

		if bytes.Contains(header, []byte("/FlateDecode")) || bytes.Contains(header, []byte("/Fl")) {
			compressed := data[streamStart:streamEnd]
			zr, err := zlib.NewReader(bytes.NewReader(compressed))
			if err == nil {
				decompressed, rErr := io.ReadAll(io.LimitReader(zr, 10<<20))
				_ = zr.Close()
				if rErr == nil && len(decompressed) > 0 {
					if err := scanBufferForPDFKeywords(decompressed); err != nil {
						return fmt.Errorf("استریم فشرده PDF حاوی محتوای فعال است: %w", err)
					}
				}
			}
		}

		idx = streamEnd + len(endStreamMarker)
	}

	return nil
}

func SanitizePDF(data []byte) ([]byte, error) {
	if err := ScanPDFForActiveContent(data); err != nil {
		return nil, err
	}

	metadataFields := []string{"/Author", "/Creator", "/Producer", "/CreationDate", "/ModDate"}
	res := make([]byte, len(data))
	copy(res, data)

	for _, field := range metadataFields {
		fBytes := []byte(field)
		idx := 0
		for {
			pos := bytes.Index(res[idx:], fBytes)
			if pos == -1 {
				break
			}
			matchPos := idx + pos + len(fBytes)
			for matchPos < len(res) && (res[matchPos] == ' ' || res[matchPos] == '\t') {
				matchPos++
			}
			if matchPos < len(res) && res[matchPos] == '(' {
				closePos := bytes.IndexByte(res[matchPos:], ')')
				if closePos != -1 {
					for k := matchPos + 1; k < matchPos+closePos; k++ {
						res[k] = ' '
					}
				}
			}
			idx = matchPos + 1
		}
	}

	return res, nil
}

func InspectOOXML(data []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("ساختار فایل آفیس معتبر نیست: %w", err)
	}

	dangerousOOXMLParts := []string{
		"vbaproject.bin",
		".bin",
		"externallink",
		"oleobject",
	}

	for _, f := range zr.File {
		cleanName := strings.ToLower(filepath.ToSlash(filepath.Clean(f.Name)))
		if strings.HasPrefix(cleanName, "..") || strings.HasPrefix(cleanName, "/") || strings.HasPrefix(cleanName, "\\") {
			return errors.New("مسیر نامعتبر در ساختار آفیس")
		}

		for _, part := range dangerousOOXMLParts {
			if strings.Contains(cleanName, part) {
				return fmt.Errorf("فایل آفیس حاوی ماکرو، ارجاع خارجی یا شیء غیرمجاز است (%s)", part)
			}
		}

		if strings.HasSuffix(cleanName, ".rels") {
			rc, err := f.Open()
			if err == nil {
				content, rErr := io.ReadAll(io.LimitReader(rc, 2<<20))
				_ = rc.Close()
				if rErr == nil {
					lowerContent := strings.ToLower(string(content))
					if strings.Contains(lowerContent, "attachedtemplate") ||
						strings.Contains(lowerContent, "targetmode=\"external\"") {
						return errors.New("فایل آفیس حاوی ارجاع قالب یا اسکریپت خارجی غیرمجاز است")
					}
				}
			}
		}
	}

	return nil
}

func SanitizeOOXML(data []byte) ([]byte, error) {
	if err := InspectOOXML(data); err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	minimalCore := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/coreProperties" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title></dc:title><dc:creator>Anonymous</dc:creator><cp:lastModifiedBy>Anonymous</cp:lastModifiedBy></cp:coreProperties>`)
	minimalApp := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"><Application>Office</Application></Properties>`)

	for _, f := range zr.File {
		clean := strings.ToLower(filepath.ToSlash(filepath.Clean(f.Name)))
		if clean == "docprops/custom.xml" {
			continue
		}

		w, err := zw.Create(f.Name)
		if err != nil {
			return nil, err
		}

		if clean == "docprops/core.xml" {
			if _, err := w.Write(minimalCore); err != nil {
				return nil, err
			}
			continue
		}

		if clean == "docprops/app.xml" {
			if _, err := w.Write(minimalApp); err != nil {
				return nil, err
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(w, rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func SanitizeMP3(data []byte) ([]byte, error) {
	out := data
	if len(out) >= 10 && bytes.HasPrefix(out, []byte("ID3")) {
		tagSize := int(out[6]&0x7F)<<21 | int(out[7]&0x7F)<<14 | int(out[8]&0x7F)<<7 | int(out[9]&0x7F)
		headerSize := 10 + tagSize
		if out[5]&0x10 != 0 {
			headerSize += 10
		}
		if headerSize < len(out) {
			out = out[headerSize:]
		}
	}

	if len(out) >= 128 && bytes.HasPrefix(out[len(out)-128:], []byte("TAG")) {
		out = out[:len(out)-128]
	}

	return out, nil
}

func SanitizeWAV(data []byte) ([]byte, error) {
	if len(data) < 12 || !bytes.HasPrefix(data, []byte("RIFF")) || !bytes.Equal(data[8:12], []byte("WAVE")) {
		return data, nil
	}

	var cleanChunks bytes.Buffer
	idx := 12
	for idx+8 <= len(data) {
		chunkID := string(data[idx : idx+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[idx+4 : idx+8]))
		chunkEnd := idx + 8 + chunkSize

		if chunkEnd > len(data) {
			break
		}

		if chunkID == "fmt " || chunkID == "data" {
			cleanChunks.Write(data[idx:chunkEnd])
		}

		if chunkSize%2 != 0 {
			chunkEnd++
		}
		idx = chunkEnd
	}

	if cleanChunks.Len() == 0 {
		return data, nil
	}

	var res bytes.Buffer
	res.Write([]byte("RIFF"))
	var totalLen [4]byte
	binary.LittleEndian.PutUint32(totalLen[:], uint32(cleanChunks.Len()+4))
	res.Write(totalLen[:])
	res.Write([]byte("WAVE"))
	res.Write(cleanChunks.Bytes())

	return res.Bytes(), nil
}

func SanitizeAttachment(filename, mimeType string, data []byte) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".jpg" || ext == ".jpeg" || mimeType == "image/jpeg" {
		return SanitizeJPEG(data)
	}

	if ext == ".png" || mimeType == "image/png" {
		return SanitizePNG(data)
	}

	if ext == ".pdf" || mimeType == "application/pdf" {
		return SanitizePDF(data)
	}

	if ext == ".docx" || ext == ".xlsx" {
		return SanitizeOOXML(data)
	}

	if ext == ".mp3" || mimeType == "audio/mpeg" {
		return SanitizeMP3(data)
	}

	if ext == ".wav" || mimeType == "audio/wav" {
		return SanitizeWAV(data)
	}

	return data, nil
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

	if len(buff) >= 12 && bytes.HasPrefix(buff, []byte("RIFF")) && bytes.Equal(buff[8:12], []byte("WAVE")) && ext == ".wav" {
		return true, "audio/wav"
	}

	return false, ""
}
