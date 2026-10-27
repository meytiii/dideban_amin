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
	"regexp"
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

var dangerousPDFNames = map[string]bool{
	"JavaScript":    true,
	"JS":            true,
	"Launch":        true,
	"EmbeddedFile":  true,
	"EmbeddedFiles": true,
	"OpenAction":    true,
	"AA":            true,
	"SubmitForm":    true,
	"ImportData":    true,
	"RichMedia":     true,
	"GoToR":         true,
	"Encrypt":       true,
}

func isPDFDelimiter(b byte) bool {
	switch b {
	case ' ', '\t', '\r', '\n', '\x00', '\f',
		'(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}

func decodePDFNameToken(raw []byte) string {
	var sb strings.Builder
	for i := 0; i < len(raw); {
		if raw[i] == '#' && i+2 < len(raw) {
			h, err := hex.DecodeString(string(raw[i+1 : i+3]))
			if err == nil && len(h) == 1 {
				sb.WriteByte(h[0])
				i += 3
				continue
			}
		}
		sb.WriteByte(raw[i])
		i++
	}
	return sb.String()
}

func scanPDFBufferTokens(buf []byte) error {
	for i := 0; i < len(buf); {
		if buf[i] == '/' {
			start := i + 1
			end := start
			for end < len(buf) && !isPDFDelimiter(buf[end]) {
				end++
			}
			if end > start {
				name := decodePDFNameToken(buf[start:end])
				if dangerousPDFNames[name] {
					if name == "Encrypt" {
						return errors.New("فایل‌های PDF رمزگذاری‌شده یا دارای کلمه عبور به دلایل امنیتی غیرمجاز هستند")
					}
					return fmt.Errorf("فایل PDF حاوی عنصر اجرایی یا اسکریپت فعال است (/%s)", name)
				}
			}
			i = end
			continue
		}
		i++
	}
	return nil
}

func ScanPDFForActiveContent(data []byte) error {
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return errors.New("فایل فاقد امضای معتبر PDF است")
	}

	streamMarker := []byte("stream")
	endStreamMarker := []byte("endstream")

	idx := 0
	for idx < len(data) {
		sIdx := bytes.Index(data[idx:], streamMarker)
		if sIdx == -1 {
			if err := scanPDFBufferTokens(data[idx:]); err != nil {
				return err
			}
			break
		}

		streamPos := idx + sIdx
		if err := scanPDFBufferTokens(data[idx:streamPos]); err != nil {
			return err
		}

		streamStart := streamPos + len(streamMarker)
		if streamStart < len(data) && data[streamStart] == '\r' {
			streamStart++
		}
		if streamStart < len(data) && data[streamStart] == '\n' {
			streamStart++
		}

		eIdx := bytes.Index(data[streamStart:], endStreamMarker)
		if eIdx == -1 {
			if err := scanPDFBufferTokens(data[streamStart:]); err != nil {
				return err
			}
			break
		}
		streamEnd := streamStart + eIdx

		headerStart := streamPos - 512
		if headerStart < 0 {
			headerStart = 0
		}
		header := data[headerStart:streamPos]

		if bytes.Contains(header, []byte("/FlateDecode")) || bytes.Contains(header, []byte("/Fl")) {
			compressed := data[streamStart:streamEnd]
			zr, err := zlib.NewReader(bytes.NewReader(compressed))
			if err == nil {
				decompressed, rErr := io.ReadAll(io.LimitReader(zr, 10<<20))
				_ = zr.Close()
				if rErr == nil && len(decompressed) > 0 {
					if err := scanPDFBufferTokens(decompressed); err != nil {
						return fmt.Errorf("استریم فشرده PDF حاوی محتوای فعال است: %w", err)
					}
				}
			}
		}

		idx = streamEnd + len(endStreamMarker)
	}

	return nil
}

var (
	rsidAttrRe   = regexp.MustCompile(`\s*w:rsid[A-Za-z0-9]*="[^"]*"`)
	commentTagRe = regexp.MustCompile(`<w:comment(RangeStart|RangeEnd|Reference)[^>]*/>`)
	trackRevRe   = regexp.MustCompile(`<(w:trackRevisions|w:documentProtection|w:writeProtection)[^>]*/>`)
	relsDropRe   = regexp.MustCompile(`<Relationship [^>]*Target="[^"]*(comments|people\.xml|customXml)[^"]*"[^>]*/>`)
)

func SanitizePDF(data []byte) ([]byte, error) {
	if err := ScanPDFForActiveContent(data); err != nil {
		return nil, err
	}

	res := make([]byte, len(data))
	copy(res, data)

	metadataFields := []string{
		"/Author", "/Creator", "/Producer", "/CreationDate", "/ModDate",
		"/Title", "/Subject", "/Keywords", "/Company", "/SourceModified",
	}

	for _, field := range metadataFields {
		fBytes := []byte(field)
		idx := 0
		for {
			pos := bytes.Index(res[idx:], fBytes)
			if pos == -1 {
				break
			}
			matchPos := idx + pos + len(fBytes)
			for matchPos < len(res) && (res[matchPos] == ' ' || res[matchPos] == '\t' || res[matchPos] == '\r' || res[matchPos] == '\n') {
				matchPos++
			}
			if matchPos < len(res) && res[matchPos] == '(' {
				depth := 1
				k := matchPos + 1
				for k < len(res) && depth > 0 {
					if res[k] == '\\' && k+1 < len(res) {
						res[k] = ' '
						res[k+1] = ' '
						k += 2
						continue
					}
					if res[k] == '(' {
						depth++
					} else if res[k] == ')' {
						depth--
					}
					if depth > 0 {
						res[k] = ' '
					}
					k++
				}
			} else if matchPos < len(res) && res[matchPos] == '<' {
				closePos := bytes.IndexByte(res[matchPos:], '>')
				if closePos != -1 {
					for k := matchPos + 1; k < matchPos+closePos; k++ {
						res[k] = '0'
					}
				}
			}
			idx = matchPos + 1
		}
	}

	infoRefRe := regexp.MustCompile(`/Info\s+\d+\s+\d+\s+R`)
	res = infoRefRe.ReplaceAllFunc(res, func(m []byte) []byte {
		return bytes.Repeat([]byte(" "), len(m))
	})

	metaRefRe := regexp.MustCompile(`/Metadata\s+\d+\s+\d+\s+R`)
	res = metaRefRe.ReplaceAllFunc(res, func(m []byte) []byte {
		return bytes.Repeat([]byte(" "), len(m))
	})

	xmpStart := []byte("<?xpacket begin")
	xmpEnd := []byte("<?xpacket end")
	idx := 0
	for {
		sPos := bytes.Index(res[idx:], xmpStart)
		if sPos == -1 {
			break
		}
		actualStart := idx + sPos
		ePos := bytes.Index(res[actualStart:], xmpEnd)
		if ePos == -1 {
			break
		}
		actualEnd := actualStart + ePos + len(xmpEnd)
		for actualEnd < len(res) && res[actualEnd] != '>' {
			actualEnd++
		}
		if actualEnd < len(res) {
			actualEnd++
		}
		for k := actualStart; k < actualEnd; k++ {
			res[k] = ' '
		}
		idx = actualEnd
	}

	eofMarker := []byte("%%EOF")
	firstEOF := bytes.Index(res, eofMarker)
	lastEOF := bytes.LastIndex(res, eofMarker)
	if firstEOF != -1 && lastEOF != -1 && firstEOF < lastEOF {
		prevTrailerRe := regexp.MustCompile(`trailer\s*<<[\s\S]*?>>`)
		earlierPortion := res[:lastEOF]
		cleared := prevTrailerRe.ReplaceAllFunc(earlierPortion, func(m []byte) []byte {
			return bytes.Repeat([]byte(" "), len(m))
		})
		copy(res[:lastEOF], cleared)
	}

	return res, nil
}

func InspectOOXML(data []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("ساختار فایل آفیس معتبر نیست: %w", err)
	}

	for _, f := range zr.File {
		cleanName := strings.ToLower(filepath.ToSlash(filepath.Clean(f.Name)))
		if strings.HasPrefix(cleanName, "..") || strings.HasPrefix(cleanName, "/") || strings.HasPrefix(cleanName, "\\") {
			return errors.New("مسیر نامعتبر در ساختار آفیس")
		}

		baseName := filepath.Base(cleanName)

		if baseName == "vbaproject.bin" || strings.HasSuffix(cleanName, "/vbaproject.bin") {
			return fmt.Errorf("فایل آفیس حاوی ماکرو غیرمجاز است (%s)", baseName)
		}

		if strings.Contains(cleanName, "/embeddings/") || strings.HasPrefix(cleanName, "embeddings/") {
			return fmt.Errorf("فایل آفیس حاوی شیء جاسازی‌شده غیرمجاز است (%s)", cleanName)
		}

		if strings.HasPrefix(baseName, "oleobject") && strings.HasSuffix(baseName, ".bin") {
			return fmt.Errorf("فایل آفیس حاوی شیء OLE غیرمجاز است (%s)", baseName)
		}

		if strings.Contains(cleanName, "/externallinks/") ||
			strings.HasPrefix(cleanName, "xl/externallinks/") ||
			strings.HasPrefix(cleanName, "word/externallinks/") ||
			strings.Contains(cleanName, "externallink") {
			return fmt.Errorf("فایل آفیس حاوی پیوند خارجی غیرمجاز است (%s)", cleanName)
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
		if clean == "docprops/custom.xml" ||
			clean == "word/comments.xml" ||
			clean == "word/commentsextended.xml" ||
			clean == "word/commentsids.xml" ||
			clean == "word/people.xml" ||
			clean == "word/commentauthors.xml" ||
			strings.HasPrefix(clean, "customxml/") ||
			strings.Contains(clean, "/customxml/") {
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

		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}

		if clean == "word/document.xml" {
			content = rsidAttrRe.ReplaceAll(content, nil)
			content = commentTagRe.ReplaceAll(content, nil)
		} else if clean == "word/settings.xml" {
			content = trackRevRe.ReplaceAll(content, nil)
		} else if strings.HasSuffix(clean, ".rels") {
			content = relsDropRe.ReplaceAll(content, nil)
		}

		if _, err := w.Write(content); err != nil {
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
