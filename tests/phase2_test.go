package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"io"
	"os"
	"strings"
	"testing"
)

func TestAADRowSwapAttack(t *testing.T) {
	key := "01234567890123456789012345678901"
	InitCrypto(key)

	plain := "Sensitive Whistleblower Report Content"
	rowA := int64(42)
	rowB := int64(99)

	ctA, err := EncryptBoundField("reports", "description", rowA, plain)
	if err != nil {
		t.Fatalf("failed to encrypt bound field: %v", err)
	}

	decA, err := DecryptBoundField("reports", "description", rowA, ctA)
	if err != nil {
		t.Fatalf("failed to decrypt valid bound field: %v", err)
	}
	if decA != plain {
		t.Fatalf("expected %q, got %q", plain, decA)
	}

	_, err = DecryptBoundField("reports", "description", rowB, ctA)
	if err == nil {
		t.Fatalf("expected authentication failure when swapping row ID from %d to %d, but decryption succeeded", rowA, rowB)
	}

	_, err = DecryptBoundField("reports", "reporter_name", rowA, ctA)
	if err == nil {
		t.Fatalf("expected authentication failure when column is mismatched, but decryption succeeded")
	}

	_, err = DecryptBoundField("evidence_files", "description", rowA, ctA)
	if err == nil {
		t.Fatalf("expected authentication failure when table is mismatched, but decryption succeeded")
	}
}

func TestUnboundFallbackRejection(t *testing.T) {
	key := "01234567890123456789012345678901"
	InitCrypto(key)

	plain := "Classified Material"
	ctBound, err := EncryptBoundField("reports", "description", 1, plain)
	if err != nil {
		t.Fatalf("failed to encrypt bound field: %v", err)
	}

	_, err = DecryptAES(ctBound)
	if err == nil {
		t.Fatalf("expected DecryptAES to reject bound v3 ciphertext without AAD, but it succeeded")
	}
}

func TestKeyRingRotation(t *testing.T) {
	oldKey := "11111111111111111111111111111111"
	newKey := "22222222222222222222222222222222"

	os.Setenv("HERASAT_KEY_ID", "k-old")
	InitCrypto(oldKey)

	plain := "Encrypted with previous generation key"
	oldCT, err := EncryptBoundField("reports", "description", 555, plain)
	if err != nil {
		t.Fatalf("failed encrypting with old key: %v", err)
	}

	os.Setenv("HERASAT_KEY_ID", "k-new")
	os.Setenv("HERASAT_AES_KEY_PREVIOUS", "k-old:"+oldKey)
	InitCrypto(newKey)

	decrypted, err := DecryptBoundField("reports", "description", 555, oldCT)
	if err != nil {
		t.Fatalf("failed decrypting old ciphertext with key ring: %v", err)
	}
	if decrypted != plain {
		t.Fatalf("expected %q, got %q", plain, decrypted)
	}

	os.Unsetenv("HERASAT_KEY_ID")
	os.Unsetenv("HERASAT_AES_KEY_PREVIOUS")
	InitCrypto("01234567890123456789012345678901")
}

func TestTrackingCodeHMACIndependence(t *testing.T) {
	InitCrypto("01234567890123456789012345678901")

	code := "TRK-2026-998877"
	hash1 := HashTrackingCode(code)
	if len(hash1) != 64 {
		t.Fatalf("expected 64-char hex tracking hash, got length %d", len(hash1))
	}

	hash2 := HashTrackingCode(code)
	if hash1 != hash2 {
		t.Fatalf("tracking hash must be deterministic")
	}

	diffHash := HashTrackingCode("TRK-2026-998878")
	if hash1 == diffHash {
		t.Fatalf("different tracking codes must have distinct hashes")
	}
}

func TestScopedSSEAccessControl(t *testing.T) {
	superAdmin := AdminSession{
		Username:    "admin_it",
		Role:        "superuser",
		IsSuperuser: true,
	}
	edareKolAdmin := AdminSession{
		Username: "admin_general",
		Role:     "edare_kol",
	}

	if !canAdminAccessReport(superAdmin, 999999) {
		t.Fatalf("superuser should have access to reports")
	}
	if !canAdminAccessReport(edareKolAdmin, 999999) {
		t.Fatalf("edare_kol should have access to reports")
	}
}

func TestMetadataDeepSanitizationPDF(t *testing.T) {
	rawPDF := []byte("%PDF-1.4\n1 0 obj\n<< /Title (Confidential) /Author (John Doe) /Producer (EvilWriter) >>\nendobj\n2 0 obj\n<< /Type /Metadata /Subtype /XML >>\nstream\n<?xpacket begin=\"\" id=\"W5M0MpCehiHzreSzNTczkc9d\"?>\n<x:xmpmeta>secret-leak</x:xmpmeta>\n<?xpacket end=\"w\"?>\nendstream\nendobj\ntrailer\n<< /Root 1 0 R /Info 1 0 R >>\n%%EOF")

	sanitized, err := SanitizePDF(rawPDF)
	if err != nil {
		t.Fatalf("SanitizePDF failed: %v", err)
	}

	if bytes.Contains(sanitized, []byte("/Info")) {
		t.Fatalf("expected /Info reference to be stripped from PDF trailer")
	}
	if bytes.Contains(sanitized, []byte("secret-leak")) {
		t.Fatalf("expected XMP packet content to be stripped from PDF metadata stream")
	}
}

func TestMetadataDeepSanitizationOOXML(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	w, _ := zw.Create("[Content_Types].xml")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`))

	w, _ = zw.Create("docProps/core.xml")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"><dc:creator>John Doe</dc:creator></cp:coreProperties>`))

	w, _ = zw.Create("word/comments.xml")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><w:comments><w:comment w:id="1">leaked comment</w:comment></w:comments>`))

	w, _ = zw.Create("word/document.xml")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><w:document w:rsidR="00112233"><w:body><w:p><w:commentReference w:id="1"/><w:r><w:t>Hello World</w:t></w:r></w:p></w:body></w:document>`))

	_ = zw.Close()

	sanitized, err := SanitizeOOXML(buf.Bytes())
	if err != nil {
		t.Fatalf("SanitizeOOXMLMetadata failed: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(sanitized), int64(len(sanitized)))
	if err != nil {
		t.Fatalf("failed to open sanitized OOXML zip: %v", err)
	}

	for _, f := range zr.File {
		if strings.Contains(f.Name, "comments.xml") {
			t.Fatalf("comments.xml was not stripped from OOXML package")
		}
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			_ = rc.Close()
			if strings.Contains(string(content), "rsid") {
				t.Fatalf("expected rsid revision attributes to be stripped from document.xml")
			}
			if strings.Contains(string(content), "commentReference") {
				t.Fatalf("expected commentReference to be stripped from document.xml")
			}
		}
	}
}

func TestDatabaseSchemaMigrationsAndWORM(t *testing.T) {
	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		if os.Getenv("HERASAT_REQUIRE_DB_TESTS") == "1" {
			t.Fatalf("HERASAT_REQUIRE_DB_TESTS=1 but HERASAT_DB_CONN is unset")
		}
		t.Skip("Skipping DB test: HERASAT_DB_CONN unset")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("db ping failed: %v", err)
	}

	if err := VerifyWORMTriggers(db); err != nil {
		t.Fatalf("VerifyWORMTriggers failed: %v", err)
	}

	if err := VerifyAppRolePrivileges(db); err != nil {
		t.Fatalf("VerifyAppRolePrivileges failed: %v", err)
	}

	var versionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&versionCount)
	if err != nil {
		t.Fatalf("schema_migrations table missing or unreadable: %v", err)
	}
	if versionCount < 2 {
		t.Fatalf("expected at least 2 migrations applied, found: %d", versionCount)
	}
}
