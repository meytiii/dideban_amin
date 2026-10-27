package main

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestUploadLimiterCapacity(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		uploadLimiter <- struct{}{}
	}

	req := httptest.NewRequest("POST", "/api/reports/submit", strings.NewReader("dummy"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=something")
	rec := httptest.NewRecorder()

	SubmitReportHandler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected HTTP 503 Service Unavailable when upload semaphore exhausted, got %d", rec.Code)
	}

	for i := 0; i < 10; i++ {
		<-uploadLimiter
	}
	wg.Wait()
}

func TestZipInspectionNestedMaliciousPDF(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("report.pdf")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}

	maliciousPDF := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R /OpenAction << /S /JavaScript /JS (app.alert(1)) >> >>\nendobj\nxref\ntrailer\n<< /Root 1 0 R >>\n%%EOF")
	if _, err := w.Write(maliciousPDF); err != nil {
		t.Fatalf("failed to write nested pdf: %v", err)
	}
	_ = zw.Close()

	err = inspectZipArchive(buf.Bytes())
	if err == nil {
		t.Fatalf("expected inspectZipArchive to reject zip containing nested active PDF")
	}
	if !strings.Contains(err.Error(), "PDF") {
		t.Fatalf("expected error message to mention PDF active content, got: %v", err)
	}
}

func TestZipInspectionSmallTextNotFalsePositive(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("notes.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}

	smallRepeating := bytes.Repeat([]byte("A"), 4096)
	if _, err := w.Write(smallRepeating); err != nil {
		t.Fatalf("failed to write text: %v", err)
	}
	_ = zw.Close()

	err = inspectZipArchive(buf.Bytes())
	if err != nil {
		t.Fatalf("expected small repetitive text to NOT be falsely flagged as zip bomb, got: %v", err)
	}
}

func TestZipInspectionLargeRatioTrigger(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	header := &zip.FileHeader{
		Name:   "huge_bomb.txt",
		Method: zip.Deflate,
	}
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	largeZeroes := make([]byte, 2<<20)
	if _, err := w.Write(largeZeroes); err != nil {
		t.Fatalf("failed to write zeroes: %v", err)
	}
	_ = zw.Close()

	err = inspectZipArchive(buf.Bytes())
	if err == nil {
		t.Fatalf("expected 2MB highly-compressed zero archive to be rejected as zip bomb")
	}
}

func TestDownloadResponseControllerDeadline(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/admin/download?id=1", nil)
	rec := httptest.NewRecorder()

	DownloadAttachmentHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without auth context, got %d", rec.Code)
	}
}
