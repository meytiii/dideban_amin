package main

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

var dangerousArchiveExtensions = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".ps1": true,
	".sh": true, ".vbs": true, ".js": true, ".dll": true,
	".scr": true, ".pif": true, ".jar": true, ".apk": true,
	".msi": true, ".iso": true, ".hta": true, ".cpl": true,
	".com": true, ".vbe": true, ".wsf": true, ".wsh": true,
	".sys": true, ".bin": true, ".elf": true, ".so": true,
	".zip": true, ".rar": true, ".7z": true, ".tar": true,
	".gz": true, ".bz2": true, ".xz": true, ".cab": true,
	".dmg": true, ".tgz": true,
}

func inspectZipArchive(rawBytes []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(rawBytes), int64(len(rawBytes)))
	if err != nil {
		return fmt.Errorf("ساختار فایل فشرده معتبر نیست: %w", err)
	}

	var totalUncompressedSize uint64
	var totalCompressedSize uint64
	const maxAllowedUncompressed = uint64(50 << 20)
	const maxEntries = 1000

	if len(reader.File) > maxEntries {
		return errors.New("تعداد فایل‌های داخل فایل فشرده بیش از حد مجاز است")
	}

	for _, f := range reader.File {
		if f.Flags&0x1 != 0 {
			return errors.New("فایل‌های فشرده رمزگذاری‌شده یا دارای کلمه عبور به دلایل امنیتی غیرمجاز هستند")
		}

		cleanName := filepath.Clean(f.Name)

		if strings.HasPrefix(cleanName, "..") || strings.HasPrefix(f.Name, "/") || strings.HasPrefix(f.Name, "\\") ||
			strings.Contains(f.Name, "../") || strings.Contains(f.Name, "..\\") {
			return errors.New("مسیر فایل در ساختار فشرده نامعتبر و ناامن است (Directory Traversal)")
		}

		innerExt := strings.ToLower(filepath.Ext(cleanName))
		if dangerousArchiveExtensions[innerExt] {
			return fmt.Errorf("فایل فشرده حاوی پسوند غیرمجاز یا آرشیو تودرتو است (%s)", innerExt)
		}

		if f.FileInfo().IsDir() {
			continue
		}

		totalCompressedSize += f.CompressedSize64

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("خطا در باز کردن محتوای آرشیو: %w", err)
		}
		lr := io.LimitReader(rc, int64(maxAllowedUncompressed+1))
		var readCount uint64
		var entryBuf bytes.Buffer
		buf := make([]byte, 8192)
		for {
			n, rErr := lr.Read(buf)
			if n > 0 {
				readCount += uint64(n)
				if innerExt == ".pdf" || innerExt == ".docx" || innerExt == ".xlsx" {
					entryBuf.Write(buf[:n])
				}
			}
			if readCount > maxAllowedUncompressed {
				_ = rc.Close()
				return errors.New("حجم غیرفشرده فایل‌های داخل آرشیو بیش از حد مجاز است (Zip Bomb)")
			}
			if rErr != nil {
				if rErr == io.EOF {
					break
				}
				_ = rc.Close()
				return fmt.Errorf("خطا در خواندن محتوای فشرده: %w", rErr)
			}
		}
		_ = rc.Close()

		totalUncompressedSize += readCount
		if totalUncompressedSize > maxAllowedUncompressed {
			return errors.New("حجم غیرفشرده فایل‌های داخل آرشیو بیش از حد مجاز است (Zip Bomb)")
		}

		if innerExt == ".pdf" {
			if pdfErr := ScanPDFForActiveContent(entryBuf.Bytes()); pdfErr != nil {
				return fmt.Errorf("فایل PDF داخل آرشیو «%s» غیرمجاز است: %w", cleanName, pdfErr)
			}
		}

		if innerExt == ".docx" || innerExt == ".xlsx" {
			if ooxmlErr := InspectOOXML(entryBuf.Bytes()); ooxmlErr != nil {
				return fmt.Errorf("فایل آفیس داخل آرشیو «%s» غیرمجاز است: %w", cleanName, ooxmlErr)
			}
		}
	}

	if totalCompressedSize > 0 && (totalUncompressedSize/totalCompressedSize) > 50 && totalUncompressedSize > (1<<20) {
		return errors.New("نسبت فشرده‌سازی فایل غیرعادی و مشکوک است (Zip Bomb)")
	}

	return nil
}

var allowedReportCategories = map[string]bool{
	"تخلفات مالی":                      true,
	"مسائل امنیتی":                     true,
	"تخلفات اداری":                     true,
	"تخلفات قانونی و اداری":            true,
	"تخلفات آموزشی":                    true,
	"پرسنلی":                           true,
	"مسائل پرسنلی":                     true,
	"دانش آموزی":                       true,
	"مسائل دانش آموزی":                 true,
	"مدارس غیردولتی":                   true,
	"مدارس و آموزشگاه‌های غیردولتی":    true,
	"آموزش مجازی":                      true,
	"مسائل مربوط به آموزش مجازی":       true,
	"فضای مجازی":                       true,
	"هنجارشکنی و تخلفات در فضای مجازی": true,
	"سایر":                 true,
	"عمومی و سایر موضوعات": true,
	"عمومی":                true,
}

func isValidReportCategory(cat string) bool {
	return allowedReportCategories[cat]
}

func validateMagicBytes(buff []byte, filename string) (bool, string) {
	return ValidateBinaryMagicBytes(buff, filename)
}

func generateTrackingCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	bytesArr := make([]byte, 16)
	if _, err := rand.Read(bytesArr); err != nil {
		panic("crypto/rand failure")
	}
	for i, b := range bytesArr {
		bytesArr[i] = charset[b%byte(len(charset))]
	}
	return string(bytesArr)
}

var uploadLimiter = make(chan struct{}, 10)

func SubmitReportHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case uploadLimiter <- struct{}{}:
		defer func() { <-uploadLimiter }()
	default:
		http.Error(w, "سامانه در حال حاضر تحت بار کاری بالا است. لطفاً چند لحظه بعد مجدداً تلاش فرمایید.", http.StatusServiceUnavailable)
		return
	}

	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(time.Now().Add(5 * time.Minute))

	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, "Payload too large or malformed", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	if r.MultipartForm != nil && len(r.MultipartForm.File["evidence"]) > 5 {
		http.Error(w, "حداکثر ۵ فایل پیوست مجاز است.", http.StatusBadRequest)
		return
	}

	captchaID := r.FormValue("captcha_id")
	captchaAnswer := r.FormValue("captcha_answer")

	if !VerifyCaptcha(captchaID, captchaAnswer) {
		http.Error(w, "کد امنیتی اشتباه است یا منقضی شده است.", http.StatusBadRequest)
		return
	}

	category := strings.TrimSpace(stripNul(r.FormValue("category")))
	description := strings.TrimSpace(stripNul(r.FormValue("description")))
	rawName := strings.TrimSpace(stripNul(r.FormValue("reporter_name")))
	rawPhone := strings.TrimSpace(stripNul(r.FormValue("reporter_phone")))
	if rawPhone != "" {
		rawPhone = sanitizePhone(rawPhone)
	}

	if description == "" || category == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(category) > 64 || len(category) > 64 {
		http.Error(w, "موضوع انتخاب‌شده نامعتبر یا طولانی‌تر از حد مجاز است.", http.StatusBadRequest)
		return
	}
	if !isValidReportCategory(category) {
		http.Error(w, "موضوع انتخاب‌شده برای گزارش معتبر نمی‌باشد.", http.StatusBadRequest)
		return
	}

	descRunes := utf8.RuneCountInString(description)
	if descRunes < 50 {
		http.Error(w, "متن شرح گزارش باید حداقل ۵۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}
	if descRunes > 5000 {
		http.Error(w, "متن شرح گزارش نمی‌تواند بیش از ۵۰۰۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	if rawName != "" && utf8.RuneCountInString(rawName) > 100 {
		http.Error(w, "نام گزارش‌دهنده نمی‌تواند بیش از ۱۰۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	if rawPhone != "" && utf8.RuneCountInString(rawPhone) > 32 {
		http.Error(w, "شماره تماس نمی‌تواند بیش از ۳۲ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	profanityAction := strings.ToLower(strings.TrimSpace(os.Getenv("HERASAT_PROFANITY_ACTION")))
	if profanityAction == "" {
		profanityAction = "flag"
	}

	hasVulgarDesc, _ := ContainsVulgarLanguage(description)
	hasVulgarName, _ := ContainsVulgarLanguage(rawName)

	if hasVulgarName {
		logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_VULGAR_NAME", nil, "anonymous")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(VulgarityLegalWarningMessage))
		return
	}

	flaggedVulgar := false
	if hasVulgarDesc {
		if profanityAction == "reject" {
			logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_VULGAR_CONTENT", nil, "anonymous")
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(VulgarityLegalWarningMessage))
			return
		}
		flaggedVulgar = true
		logAudit("anonymous_user", "PUBLIC_REPORTER", "FLAGGED_PROFANITY_CONTENT", nil, "anonymous")
	}

	reportType := "anonymous"
	if rawName != "" || rawPhone != "" {
		reportType = "identified"
	}

	trackingCode := generateTrackingCode()
	trackingHash := HashTrackingCode(trackingCode)
	encTrackingCode, err := EncryptAES(trackingCode)
	if err != nil {
		http.Error(w, "Encryption engine failure", http.StatusInternalServerError)
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		http.Error(w, "Database transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var reportID int64
	err = tx.QueryRow("SELECT nextval(pg_get_serial_sequence('reports', 'id'))").Scan(&reportID)
	if err != nil {
		http.Error(w, "Failed to allocate report identifier", http.StatusInternalServerError)
		return
	}

	encDesc, err := EncryptBoundField("reports", "description", reportID, description)
	if err != nil {
		http.Error(w, "Encryption engine failure", http.StatusInternalServerError)
		return
	}

	var encName, encPhone sql.NullString
	if reportType == "identified" {
		if rawName != "" {
			if name, err := EncryptBoundField("reports", "reporter_name", reportID, rawName); err == nil {
				encName = sql.NullString{String: name, Valid: true}
			}
		}
		if rawPhone != "" {
			if phone, err := EncryptBoundField("reports", "reporter_phone", reportID, rawPhone); err == nil {
				encPhone = sql.NullString{String: phone, Valid: true}
			}
		}
	}

	initialPriority := "normal"
	var initialNotes sql.NullString
	if flaggedVulgar {
		initialPriority = "high"
		if notes, err := EncryptBoundField("reports", "internal_notes", reportID, "سیستم: حاوی الفاظ نامتعارف، علامت‌گذاری جهت بررسی توسط کارشناس."); err == nil {
			initialNotes = sql.NullString{String: notes, Valid: true}
		}
	}

	_, err = tx.Exec(`
		INSERT INTO reports (id, tracking_code, tracking_code_hash, category, report_type, priority, internal_notes, reporter_name, reporter_phone, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		reportID, encTrackingCode, trackingHash, category, reportType, initialPriority, initialNotes, encName, encPhone, encDesc)
	if err != nil {
		http.Error(w, "Failed to save report data", http.StatusInternalServerError)
		return
	}

	var savedFilesOnDisk []string
	cleanupDiskFiles := func() {
		for _, p := range savedFilesOnDisk {
			_ = os.Remove(p)
		}
	}

	if r.MultipartForm != nil && len(r.MultipartForm.File["evidence"]) > 0 {
		evidenceHeaders := r.MultipartForm.File["evidence"]

		var totalEvidenceSize int64
		for _, fh := range evidenceHeaders {
			totalEvidenceSize += fh.Size
		}
		if totalEvidenceSize > 20<<20 {
			http.Error(w, "مجموع حجم فایل‌های پیوست نمی‌تواند بیشتر از ۲۰ مگابایت باشد.", http.StatusBadRequest)
			return
		}

		os.MkdirAll("./secure_uploads", 0700)

		for _, header := range evidenceHeaders {
			if header.Size == 0 {
				continue
			}

			file, err := header.Open()
			if err != nil {
				cleanupDiskFiles()
				http.Error(w, "Error reading attached file", http.StatusInternalServerError)
				return
			}

			sniffBuf := make([]byte, 512)
			n, sErr := io.ReadFull(file, sniffBuf)
			if sErr != nil && sErr != io.EOF && sErr != io.ErrUnexpectedEOF {
				file.Close()
				cleanupDiskFiles()
				http.Error(w, "Error reading file signature", http.StatusInternalServerError)
				return
			}
			sniffBuf = sniffBuf[:n]

			if len(sniffBuf) == 0 {
				file.Close()
				continue
			}

			isValidSignature, detectedMime := validateMagicBytes(sniffBuf, header.Filename)
			if !isValidSignature {
				file.Close()
				cleanupDiskFiles()
				http.Error(w, fmt.Sprintf("فرمت فایل «%s» غیرمجاز است یا ساختار باینری فایل با پسوند آن همخوانی ندارد.", header.Filename), http.StatusUnsupportedMediaType)
				return
			}
			mimeType := detectedMime
			ext := strings.ToLower(filepath.Ext(header.Filename))

			if _, err := file.Seek(0, io.SeekStart); err != nil {
				file.Close()
				cleanupDiskFiles()
				http.Error(w, "Error rewinding file stream", http.StatusInternalServerError)
				return
			}

			clam := NewClamAVScanner()
			if clam.IsEnabled() {
				clean, _, scanErr := clam.ScanStream(file)
				if scanErr != nil {
					file.Close()
					if errors.Is(scanErr, ErrMalwareDetected) {
						logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_MALICIOUS_FILE", nil, "anonymous")
						cleanupDiskFiles()
						http.Error(w, "فایل ارسالی آلوده به بدافزار تشخیص داده شد", http.StatusBadRequest)
						return
					}
					logAudit("system", "ANTIVIRUS", "CLAMAV_SERVICE_UNAVAILABLE", nil, "anonymous")
					cleanupDiskFiles()
					http.Error(w, "سامانه اسکن امنیتی موقتاً در دسترس نیست.", http.StatusServiceUnavailable)
					return
				}
				if !clean {
					file.Close()
					logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_MALICIOUS_FILE", nil, "anonymous")
					cleanupDiskFiles()
					http.Error(w, "فایل ارسالی آلوده به بدافزار تشخیص داده شد", http.StatusBadRequest)
					return
				}
				if _, err := file.Seek(0, io.SeekStart); err != nil {
					file.Close()
					cleanupDiskFiles()
					http.Error(w, "Error rewinding file stream", http.StatusInternalServerError)
					return
				}
			}

			rawBytes, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				cleanupDiskFiles()
				http.Error(w, "Error reading file content", http.StatusInternalServerError)
				return
			}

			if mimeType == "application/pdf" || ext == ".pdf" {
				if pdfErr := ScanPDFForActiveContent(rawBytes); pdfErr != nil {
					logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_ACTIVE_PDF_PAYLOAD", nil, "anonymous")
					cleanupDiskFiles()
					http.Error(w, fmt.Sprintf("فایل «%s» غیرمجاز است: %v", header.Filename, pdfErr), http.StatusBadRequest)
					return
				}
			}

			if mimeType == "application/zip" || ext == ".zip" {
				if zipErr := inspectZipArchive(rawBytes); zipErr != nil {
					logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_UNSAFE_ZIP_PAYLOAD", nil, "anonymous")
					cleanupDiskFiles()
					http.Error(w, fmt.Sprintf("فایل فشرده «%s» نامعتبر یا غیرمجاز است: %v", header.Filename, zipErr), http.StatusBadRequest)
					return
				}
			}

			if ext == ".docx" || ext == ".xlsx" {
				if ooxmlErr := InspectOOXML(rawBytes); ooxmlErr != nil {
					logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_UNSAFE_OOXML_PAYLOAD", nil, "anonymous")
					cleanupDiskFiles()
					http.Error(w, fmt.Sprintf("فایل آفیس «%s» غیرمجاز است: %v", header.Filename, ooxmlErr), http.StatusBadRequest)
					return
				}
			}

			sanitizedBytes, err := SanitizeAttachment(header.Filename, mimeType, rawBytes)
			if err != nil {
				cleanupDiskFiles()
				http.Error(w, "خطا در پردازش و اعتبارسنجی ساختار فایل", http.StatusBadRequest)
				return
			}
			canonicalBytes := sanitizedBytes

			hasher := sha256.New()
			hasher.Write(canonicalBytes)
			fileHash := hex.EncodeToString(hasher.Sum(nil))

			encryptedBytes, err := EncryptFileBytes(canonicalBytes)
			if err != nil {
				cleanupDiskFiles()
				http.Error(w, "File encryption failure", http.StatusInternalServerError)
				return
			}

			safeFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), generateTrackingCode()) + filepath.Ext(header.Filename)
			savePath := filepath.Join("./secure_uploads", safeFileName)

			if err := os.WriteFile(savePath, encryptedBytes, 0600); err != nil {
				cleanupDiskFiles()
				http.Error(w, "Internal storage failure", http.StatusInternalServerError)
				return
			}
			savedFilesOnDisk = append(savedFilesOnDisk, savePath)

			var fileID int64
			err = tx.QueryRow("SELECT nextval(pg_get_serial_sequence('evidence_files', 'id'))").Scan(&fileID)
			if err != nil {
				cleanupDiskFiles()
				http.Error(w, "Failed to allocate file identifier", http.StatusInternalServerError)
				return
			}

			encPath, _ := EncryptBoundField("evidence_files", "file_path", fileID, savePath)
			encFileName, _ := EncryptBoundField("evidence_files", "file_name", fileID, header.Filename)

			_, err = tx.Exec(`
				INSERT INTO evidence_files (id, report_id, file_path, file_hash, mime_type, file_name, file_size)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				fileID, reportID, encPath, fileHash, mimeType, encFileName, int64(len(canonicalBytes)))
			if err != nil {
				cleanupDiskFiles()
				http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		cleanupDiskFiles()
		http.Error(w, "Failed to commit report", http.StatusInternalServerError)
		return
	}

	GlobalSSEBroker.BroadcastNewReport(reportID)

	_, _ = DB.Exec(`
		INSERT INTO admin_notifications (notif_type, title, message, target_report_id)
		VALUES ($1, $2, $3, $4)`,
		"new_report",
		"گزارش محرمانه جدید در کارتابل",
		"موضوع: "+category,
		reportID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":        "success",
		"message":       "گزارش شما با موفقیت ثبت شد",
		"tracking_code": trackingCode,
	})
}

type ReportMessageItem struct {
	ID            int64  `json:"id"`
	SenderType    string `json:"sender_type"`
	Message       string `json:"message"`
	AdminUsername string `json:"admin_username,omitempty"`
	AdminFullName string `json:"admin_full_name,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type PublicReportStatus struct {
	TrackingCode  string              `json:"tracking_code"`
	Category      string              `json:"category"`
	CreatedAt     string              `json:"created_at"`
	IsRead        bool                `json:"is_read"`
	AdminResponse string              `json:"admin_response"`
	RespondedAt   string              `json:"responded_at"`
	Messages      []ReportMessageItem `json:"messages"`
}

func TrackReportHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TrackingCode  string `json:"tracking_code"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !VerifyCaptcha(req.CaptchaID, req.CaptchaAnswer) {
		http.Error(w, "کد امنیتی اشتباه است یا منقضی شده است.", http.StatusBadRequest)
		return
	}

	req.TrackingCode = strings.TrimSpace(stripNul(req.TrackingCode))
	if !isValidTrackingCode(req.TrackingCode) {
		http.Error(w, "کد پیگیری نامعتبر است.", http.StatusBadRequest)
		return
	}

	var status PublicReportStatus
	var encResponse sql.NullString
	var createdAt time.Time
	var respAt sql.NullTime
	var reportID int64

	trackingHash := HashTrackingCode(req.TrackingCode)
	err := DB.QueryRow(`
		SELECT id, tracking_code, category, is_read, created_at, admin_response, responded_at
		FROM reports
		WHERE tracking_code_hash = $1`, trackingHash).
		Scan(&reportID, &status.TrackingCode, &status.Category, &status.IsRead, &createdAt, &encResponse, &respAt)

	if err == sql.ErrNoRows {
		http.Error(w, "گزارشی با این کد پیگیری یافت نشد.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	status.TrackingCode = req.TrackingCode

	status.CreatedAt = createdAt.Format(time.RFC3339)
	if respAt.Valid {
		status.RespondedAt = respAt.Time.Format(time.RFC3339)
	}

	if encResponse.Valid && encResponse.String != "" {
		decResp, err := DecryptBoundField("reports", "admin_response", reportID, encResponse.String)
		if err == nil {
			status.AdminResponse = decResp
		}
	}

	msgRows, err := DB.Query(`
		SELECT id, sender_type, message_text, COALESCE(admin_username, ''), created_at
		FROM report_messages
		WHERE report_id = $1
		ORDER BY created_at ASC, id ASC`, reportID)
	status.Messages = []ReportMessageItem{}
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var m ReportMessageItem
			var encMsg string
			var msgCreatedAt time.Time
			if err := msgRows.Scan(&m.ID, &m.SenderType, &encMsg, &m.AdminUsername, &msgCreatedAt); err == nil {
				decMsg, err := DecryptBoundField("report_messages", "message", m.ID, encMsg)
				if err == nil {
					m.Message = decMsg
				} else {
					m.Message = "[پیام رمزنگاری‌شده]"
				}
				m.CreatedAt = msgCreatedAt.Format(time.RFC3339)
				if m.SenderType == "admin" {
					m.AdminUsername = ""
				}
				status.Messages = append(status.Messages, m)
			}
		}
		if err := msgRows.Err(); err != nil {
			log.Printf("Error iterating report messages: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func SubmitFollowUpHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TrackingCode  string `json:"tracking_code"`
		Message       string `json:"message"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !VerifyCaptcha(req.CaptchaID, req.CaptchaAnswer) {
		http.Error(w, "کد امنیتی اشتباه است یا منقضی شده است.", http.StatusBadRequest)
		return
	}

	req.TrackingCode = strings.TrimSpace(stripNul(req.TrackingCode))
	req.Message = strings.TrimSpace(stripNul(req.Message))
	if req.TrackingCode == "" || req.Message == "" {
		http.Error(w, "کد پیگیری و متن پیام تکمیلی الزامی است.", http.StatusBadRequest)
		return
	}

	if !isValidTrackingCode(req.TrackingCode) {
		http.Error(w, "کد پیگیری نامعتبر است.", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(req.Message) > 4000 {
		http.Error(w, "متن پیام تکمیلی نمی‌تواند بیش از ۴۰۰۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	flaggedFollowUpVulgar := false
	if hasVulgar, _ := ContainsVulgarLanguage(req.Message); hasVulgar {
		action := strings.ToLower(strings.TrimSpace(os.Getenv("HERASAT_PROFANITY_ACTION")))
		if action == "reject" {
			logAudit("anonymous_user", "PUBLIC_REPORTER", "REJECTED_VULGAR_FOLLOWUP", nil, "anonymous")
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(VulgarityLegalWarningMessage))
			return
		}
		flaggedFollowUpVulgar = true
		logAudit("anonymous_user", "PUBLIC_REPORTER", "FLAGGED_VULGAR_FOLLOWUP", nil, "anonymous")
	}

	trackingHash := HashTrackingCode(req.TrackingCode)
	var reportID int64
	err := DB.QueryRow("SELECT id FROM reports WHERE tracking_code_hash = $1", trackingHash).Scan(&reportID)
	if err == sql.ErrNoRows {
		http.Error(w, "گزارشی با این کد پیگیری یافت نشد.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var lastMsgID int64
	var lastMsgEnc string
	var lastMsgCreatedAt time.Time
	err = DB.QueryRow(`
		SELECT id, message_text, created_at FROM report_messages
		WHERE report_id = $1 AND sender_type = 'reporter'
		ORDER BY id DESC LIMIT 1`, reportID).Scan(&lastMsgID, &lastMsgEnc, &lastMsgCreatedAt)
	if err == nil && time.Since(lastMsgCreatedAt) < 30*time.Second {
		if dec, err := DecryptBoundField("report_messages", "message", lastMsgID, lastMsgEnc); err == nil && dec == req.Message {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status": "success", "message": "پیام تکمیلی شما با موفقیت ثبت شد"}`))
			return
		}
	}

	var msgID int64
	err = DB.QueryRow("SELECT nextval(pg_get_serial_sequence('report_messages', 'id'))").Scan(&msgID)
	if err != nil {
		http.Error(w, "Failed to allocate message identifier", http.StatusInternalServerError)
		return
	}

	encMsg, err := EncryptBoundField("report_messages", "message", msgID, req.Message)
	if err != nil {
		http.Error(w, "Encryption failure", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(`
		INSERT INTO report_messages (id, report_id, sender_type, message_text)
		VALUES ($1, $2, 'reporter', $3)`,
		msgID, reportID, encMsg)
	if err != nil {
		http.Error(w, "Failed to save follow-up message", http.StatusInternalServerError)
		return
	}

	if flaggedFollowUpVulgar {
		_, _ = DB.Exec("UPDATE reports SET has_new_message = true, priority = 'high', updated_at = NOW() WHERE id = $1", reportID)
	} else {
		_, _ = DB.Exec("UPDATE reports SET has_new_message = true, updated_at = NOW() WHERE id = $1", reportID)
	}

	GlobalSSEBroker.BroadcastNewMessage(reportID)

	_, _ = DB.Exec(`
		INSERT INTO admin_notifications (notif_type, title, message, target_report_id, recipient_username)
		VALUES ($1, $2, $3, $4, NULL)`,
		"new_reply",
		"پاسخ و پیام تکمیلی از گزارش‌دهنده",
		"پیام جدید از طرف گزارش‌دهنده دریافت شد.",
		reportID,
	)

	fwdRows, err := DB.Query("SELECT recipient_username FROM report_forwardings WHERE report_id = $1", reportID)
	if err == nil {
		defer fwdRows.Close()
		for fwdRows.Next() {
			var recipUser string
			if scanErr := fwdRows.Scan(&recipUser); scanErr == nil && recipUser != "" {
				_, _ = DB.Exec(`
					INSERT INTO admin_notifications (notif_type, title, message, target_report_id, recipient_username)
					VALUES ($1, $2, $3, $4, $5)`,
					"new_reply",
					"پاسخ و پیام تکمیلی از گزارش‌دهنده",
					"پیام جدید در پرونده ارجاعی دریافت شد.",
					reportID,
					recipUser,
				)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "پیام تکمیلی شما با موفقیت در بستری امن ثبت گردید."}`))
}
