package main

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func TestDocxSandboxMaliciousPayload(t *testing.T) {

	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not found in PATH, skipping docx test")
	}

	db := getTestDB(t)
	oldDB := DB
	DB = db
	defer func() { DB = oldDB }()

	InitCrypto("01234567890123456789012345678901")

	maliciousPayload := "__import__('os').system('rm -rf /'); alert('xss'); {{7*7}}"
	encDesc, err := EncryptAES(maliciousPayload)
	if err != nil {
		t.Fatalf("Failed to encrypt malicious payload: %v", err)
	}

	trackingCode := "TEST-DOCX-SEC-1"
	var reportID int64
	err = db.QueryRow(`
		INSERT INTO reports (tracking_code, category, report_type, priority, description, created_at, updated_at)
		VALUES ($1, 'administrative', 'complaint', 'normal', $2, NOW(), NOW())
		RETURNING id
	`, trackingCode, encDesc).Scan(&reportID)
	if err != nil {
		t.Fatalf("Failed to insert report: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/report/export-docx?id="+strconv.FormatInt(reportID, 10), nil)
	adminSession := AdminSession{
		Username:    "superuser_test",
		FullName:    "Super User",
		Title:       "Director",
		Role:        "superuser",
		IsSuperuser: true,
	}
	ctx := contextWithAdmin(req.Context(), adminSession)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	ExportReportDocxHandler(rec, req)

	if rec.Code != http.StatusOK {
		if _, err := os.Stat("templates/report_template.docx"); os.IsNotExist(err) {
			t.Skip("Skipping TestDocxSandboxMaliciousPayload: templates/report_template.docx is not present on this machine")
			return
		}
		t.Fatalf("ExportReportDocxHandler returned status %d: %s", rec.Code, rec.Body.String())
	}

	docxBytes := rec.Body.Bytes()
	if len(docxBytes) == 0 {
		t.Fatalf("Expected non-empty docx bytes")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("Failed to open output as zip (docx): %v", err)
	}

	var foundDocumentXML bool
	var docXMLContent string
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			foundDocumentXML = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open word/document.xml: %v", err)
			}
			data, _ := io.ReadAll(rc)
			rc.Close()
			docXMLContent = string(data)
			break
		}
	}

	if !foundDocumentXML {
		t.Fatalf("word/document.xml not found in generated docx")
	}

	if !strings.Contains(docXMLContent, "__import__") || !strings.Contains(docXMLContent, "rm -rf") {
		t.Fatalf("Expected document.xml to contain payload literally without code execution")
	}
}
