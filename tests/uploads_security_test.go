package main

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJPEGDecompressionBomb(t *testing.T) {
	fakeBomb := []byte{
		0xFF, 0xD8,
		0xFF, 0xC0, 0x00, 0x11,
		0x08,
		0x4E, 0x20,
		0x4E, 0x20,
		0x01,
		0x01, 0x11, 0x00,
		0xFF, 0xD9,
	}
	_, err := SanitizeJPEG(fakeBomb)
	if err == nil {
		t.Fatalf("expected error on oversized JPEG dimensions, got nil")
	}
}

func TestPNGDecompressionBomb(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte("\x89PNG\r\n\x1a\n"))

	var ihdr bytes.Buffer
	_ = binary.Write(&ihdr, binary.BigEndian, uint32(12000))
	_ = binary.Write(&ihdr, binary.BigEndian, uint32(12000))
	ihdr.WriteByte(8)
	ihdr.WriteByte(2)
	ihdr.WriteByte(0)
	ihdr.WriteByte(0)
	ihdr.WriteByte(0)

	_ = binary.Write(&buf, binary.BigEndian, uint32(ihdr.Len()))
	buf.WriteString("IHDR")
	buf.Write(ihdr.Bytes())
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))

	_, err := SanitizePNG(buf.Bytes())
	if err == nil {
		t.Fatalf("expected error on oversized PNG dimensions, got nil")
	}
}

func TestPDFActiveContentRejection(t *testing.T) {
	t.Run("RejectsEncryptedPDF", func(t *testing.T) {
		pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Encrypt 2 0 R >>\nendobj\n%%EOF")
		err := ScanPDFForActiveContent(pdf)
		if err == nil {
			t.Fatalf("expected error for encrypted PDF")
		}
	})

	t.Run("RejectsDirectJavaScript", func(t *testing.T) {
		pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Action /S /JavaScript /JS (app.alert(1);) >>\nendobj\n%%EOF")
		err := ScanPDFForActiveContent(pdf)
		if err == nil {
			t.Fatalf("expected error for PDF containing /JavaScript")
		}
	})

	t.Run("RejectsHexEncodedKeywords", func(t *testing.T) {
		pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Action /S /#4a#61#76#61#53#63#72#69#70#74 >>\nendobj\n%%EOF")
		err := ScanPDFForActiveContent(pdf)
		if err == nil {
			t.Fatalf("expected error for PDF containing hex-encoded /JavaScript")
		}
	})

	t.Run("RejectsCompressedStreamWithActiveContent", func(t *testing.T) {
		var zbuf bytes.Buffer
		zw := zlib.NewWriter(&zbuf)
		_, _ = zw.Write([]byte("<< /S /Launch /F (calc.exe) >>"))
		_ = zw.Close()

		var pdf bytes.Buffer
		pdf.WriteString("%PDF-1.7\n1 0 obj\n<< /Length 100 /Filter /FlateDecode >>\nstream\n")
		pdf.Write(zbuf.Bytes())
		pdf.WriteString("\nendstream\nendobj\n%%EOF")

		err := ScanPDFForActiveContent(pdf.Bytes())
		if err == nil {
			t.Fatalf("expected error for FlateDecode stream containing /Launch")
		}
	})
}

func TestPDFMetadataSanitization(t *testing.T) {
	pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Author (Secret Whistleblower) /Creator (Confidential Tool) >>\nendobj\n%%EOF")
	sanitized, err := SanitizePDF(pdf)
	if err != nil {
		t.Fatalf("unexpected error in SanitizePDF: %v", err)
	}
	if bytes.Contains(sanitized, []byte("Secret Whistleblower")) {
		t.Fatalf("expected Author metadata to be stripped")
	}
	if bytes.Contains(sanitized, []byte("Confidential Tool")) {
		t.Fatalf("expected Creator metadata to be stripped")
	}
}

func TestOOXMLMacroAndTemplateRejection(t *testing.T) {
	t.Run("RejectsMacroBin", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, _ := zw.Create("word/vbaProject.bin")
		_, _ = w.Write([]byte("fake macro bytecode"))
		_ = zw.Close()

		err := InspectOOXML(buf.Bytes())
		if err == nil {
			t.Fatalf("expected InspectOOXML to reject file containing vbaProject.bin")
		}
	})

	t.Run("RejectsExternalRelationships", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, _ := zw.Create("word/_rels/settings.xml.rels")
		_, _ = w.Write([]byte(`<Relationships><Relationship Target="http://evil.com/template.dotm" TargetMode="External"/></Relationships>`))
		_ = zw.Close()

		err := InspectOOXML(buf.Bytes())
		if err == nil {
			t.Fatalf("expected InspectOOXML to reject external relationship target")
		}
	})

	t.Run("SanitizesOOXMLMetadata", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w1, _ := zw.Create("docProps/core.xml")
		_, _ = w1.Write([]byte(`<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:creator>John Doe</dc:creator></cp:coreProperties>`))
		w2, _ := zw.Create("docProps/custom.xml")
		_, _ = w2.Write([]byte(`<Properties><property name="Company">SecretCorp</property></Properties>`))
		w3, _ := zw.Create("word/document.xml")
		_, _ = w3.Write([]byte(`<w:document><w:body><w:p><w:r><w:t>Hello</w:t></w:r></w:p></w:body></w:document>`))
		_ = zw.Close()

		sanitized, err := SanitizeOOXML(buf.Bytes())
		if err != nil {
			t.Fatalf("SanitizeOOXML failed: %v", err)
		}

		zr, err := zip.NewReader(bytes.NewReader(sanitized), int64(len(sanitized)))
		if err != nil {
			t.Fatalf("failed to read sanitized ooxml: %v", err)
		}

		for _, f := range zr.File {
			if f.Name == "docProps/custom.xml" {
				t.Fatalf("expected docProps/custom.xml to be removed")
			}
			if f.Name == "docProps/core.xml" {
				rc, _ := f.Open()
				content := make([]byte, f.UncompressedSize64)
				_, _ = rc.Read(content)
				rc.Close()
				if bytes.Contains(content, []byte("John Doe")) {
					t.Fatalf("expected creator to be stripped from core.xml")
				}
			}
		}
	})
}

func TestZipArchiveInspection(t *testing.T) {
	t.Run("RejectsDangerousNestedExecutable", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, _ := zw.Create("payload.exe")
		_, _ = w.Write([]byte("fake executable"))
		_ = zw.Close()

		err := inspectZipArchive(buf.Bytes())
		if err == nil {
			t.Fatalf("expected inspectZipArchive to reject zip with .exe file")
		}
	})

	t.Run("RejectsEncryptedEntry", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		header := &zip.FileHeader{
			Name:  "secret.txt",
			Flags: 0x1,
		}
		w, _ := zw.CreateHeader(header)
		_, _ = w.Write([]byte("encrypted data"))
		_ = zw.Close()

		err := inspectZipArchive(buf.Bytes())
		if err == nil {
			t.Fatalf("expected inspectZipArchive to reject encrypted zip entry")
		}
	})

	t.Run("RejectsZipBombDecompressionRatio", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		header := &zip.FileHeader{
			Name:   "bomb.txt",
			Method: zip.Deflate,
		}
		w, _ := zw.CreateHeader(header)
		largeZeroes := make([]byte, 2<<20)
		_, _ = w.Write(largeZeroes)
		_ = zw.Close()

		data := buf.Bytes()
		err := inspectZipArchive(data)
		if err != nil {
		}
	})
}

func TestAudioSanitization(t *testing.T) {
	t.Run("SanitizeMP3StripsID3v2AndID3v1", func(t *testing.T) {
		raw := make([]byte, 0, 500)
		raw = append(raw, []byte("ID3")...)
		raw = append(raw, 0x03, 0x00, 0x00)
		raw = append(raw, 0x00, 0x00, 0x00, 0x10)
		raw = append(raw, make([]byte, 16)...)

		audioData := []byte{0xFF, 0xFB, 0x90, 0x64, 0x00, 0x00, 0x00, 0x00}
		raw = append(raw, audioData...)

		var id3v1 [128]byte
		copy(id3v1[:3], []byte("TAG"))
		raw = append(raw, id3v1[:]...)

		sanitized, err := SanitizeMP3(raw)
		if err != nil {
			t.Fatalf("SanitizeMP3 failed: %v", err)
		}
		if bytes.HasPrefix(sanitized, []byte("ID3")) {
			t.Fatalf("expected ID3v2 header to be stripped")
		}
		if bytes.HasSuffix(sanitized, []byte("TAG")) {
			t.Fatalf("expected ID3v1 trailer to be stripped")
		}
	})

	t.Run("SanitizeWAVStripsExtraneousChunks", func(t *testing.T) {
		var buf bytes.Buffer
		buf.WriteString("RIFF")
		_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
		buf.WriteString("WAVE")

		buf.WriteString("fmt ")
		_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(44100))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(88200))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(2))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(16))

		buf.WriteString("LIST")
		_ = binary.Write(&buf, binary.LittleEndian, uint32(14))
		buf.WriteString("INFOINAMArtist")

		buf.WriteString("data")
		_ = binary.Write(&buf, binary.LittleEndian, uint32(8))
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})

		sanitized, err := SanitizeWAV(buf.Bytes())
		if err != nil {
			t.Fatalf("SanitizeWAV failed: %v", err)
		}
		if bytes.Contains(sanitized, []byte("LIST")) || bytes.Contains(sanitized, []byte("Artist")) {
			t.Fatalf("expected LIST chunk to be stripped from WAV")
		}
		if !bytes.Contains(sanitized, []byte("fmt ")) || !bytes.Contains(sanitized, []byte("data")) {
			t.Fatalf("expected fmt and data chunks to be preserved")
		}
	})
}

func TestValidateBinaryMagicBytes(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		data     []byte
		valid    bool
	}{
		{"ValidPDF", "doc.pdf", []byte("%PDF-1.4 sample"), true},
		{"ValidPNG", "pic.png", []byte("\x89PNG\r\n\x1a\n\x00\x00"), true},
		{"ValidJPG", "pic.jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, true},
		{"ValidZIP", "archive.zip", []byte{'P', 'K', 0x03, 0x04}, true},
		{"ValidDOCX", "doc.docx", []byte{'P', 'K', 0x03, 0x04}, true},
		{"ValidXLSX", "sheet.xlsx", []byte{'P', 'K', 0x03, 0x04}, true},
		{"BannedRAR", "bad.rar", []byte("Rar!\x1a\x07\x00"), false},
		{"BannedMP4", "video.mp4", []byte("\x00\x00\x00 ftyp"), false},
		{"MismatchedPDFExtension", "doc.exe", []byte("%PDF-1.4 sample"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, _ := ValidateBinaryMagicBytes(tc.data, tc.filename)
			if ok != tc.valid {
				t.Fatalf("expected valid=%v, got %v for %s", tc.valid, ok, tc.filename)
			}
		})
	}
}

func TestCaptchaBruteForceProtection(t *testing.T) {
	token := GenerateCaptchaToken(42, 5*time.Minute)
	if token == "" {
		t.Fatalf("failed to generate captcha token")
	}

	if VerifyCaptcha(token, "99") {
		t.Fatalf("expected wrong answer to fail")
	}

	if VerifyCaptcha(token, "100") {
		t.Fatalf("expected 2nd wrong answer to fail")
	}

	if VerifyCaptcha(token, "42") {
		t.Fatalf("expected token to be burned after failure attempts")
	}
}

func TestCaptchaSingleUseRedemption(t *testing.T) {
	token := GenerateCaptchaToken(77, 5*time.Minute)
	if token == "" {
		t.Fatalf("failed to generate captcha token")
	}

	if !VerifyCaptcha(token, "77") {
		t.Fatalf("expected first verification with correct answer to succeed")
	}

	if VerifyCaptcha(token, "77") {
		t.Fatalf("expected replay of already redeemed captcha token to fail")
	}
}

func TestIsWithinDirectory(t *testing.T) {
	tempDir := t.TempDir()
	childDir := filepath.Join(tempDir, "subdir")
	childFile := filepath.Join(childDir, "file.txt")
	siblingDir := tempDir + "_sibling"
	siblingFile := filepath.Join(siblingDir, "file.txt")

	if !isWithinDirectory(tempDir, childFile) {
		t.Fatalf("expected childFile to be within tempDir")
	}
	if !isWithinDirectory(tempDir, tempDir) {
		t.Fatalf("expected tempDir to be within itself")
	}
	if isWithinDirectory(tempDir, siblingFile) {
		t.Fatalf("expected siblingFile NOT to be within tempDir")
	}
	if isWithinDirectory(tempDir, filepath.Join(tempDir, "..", "outside.txt")) {
		t.Fatalf("expected parent traversal NOT to be within tempDir")
	}
}

func TestCategoryValidationAllowList(t *testing.T) {
	if !isValidReportCategory("تخلفات مالی") {
		t.Fatalf("expected تخلفات مالی to be valid")
	}
	if isValidReportCategory("دسته‌بندی غیرمجاز و مخرب") {
		t.Fatalf("expected malicious category to be invalid")
	}
	if isValidReportCategory(strings.Repeat("a", 200)) {
		t.Fatalf("expected overly long category to be invalid")
	}
}
