package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func deriveKey(masterKey []byte, info string) []byte {
	salt := make([]byte, sha256.Size)
	extractMac := hmac.New(sha256.New, salt)
	extractMac.Write(masterKey)
	prk := extractMac.Sum(nil)

	expandMac := hmac.New(sha256.New, prk)
	expandMac.Write([]byte(info))
	expandMac.Write([]byte{0x01})
	return expandMac.Sum(nil)
}

func encryptWithKey(key []byte, prefix, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return prefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func main() {
	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		connStr = "host=127.0.0.1 port=5432 user=herasat_user password=herasat1405@ dbname=herasat_db sslmode=disable"
	}

	aesKeyStr := os.Getenv("HERASAT_AES_KEY")
	if aesKeyStr == "" {
		aesKeyStr = "HERASAT_SECURE_KEY_32_CHARS_LONG"
	}
	if len(aesKeyStr) != 32 {
		log.Fatalf("AES key must be 32 bytes")
	}

	keyFields := deriveKey([]byte(aesKeyStr), "HERASAT_REPORT_FIELDS_V2")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	categories := []string{
		"financial",
		"administrative",
		"security",
		"complaint",
		"service",
	}

	priorities := []string{
		"normal",
		"urgent",
		"critical",
	}

	classifications := []string{
		"confidential",
		"secret",
		"top_secret",
	}

	sampleSubjects := []string{
		"گزارش مغایرت مالی در صدور احکام اداری منطقه ۲",
		"بررسی مستندات دریافتی غیرمجاز در فرآیند خرید تجهیزات",
		"تخلف در ثبت نمرات و فرآیند ارزیابی آزمون استخدامی",
		"درخواست رسیدگی به سوءاستفاده از موقعیت شغلی در شعبه",
		"گزارش نقص حفاظت فیزیکی و بایگانی اسناد محرمانه آموزشگاه",
		"ابهام در نحوه تخصیص بودجه رفاهی و تسهیلات کارکنان",
		"عدم رعایت ضوابط حراستی در ورود و خروج ارباب رجوع",
		"گزارش جعل امضا در مکاتبات رسمی و قراردادهای پشتیبانی",
		"عدم شفافیت در واگذاری بوفه و امکانات مدارس به بخش خصوصی",
		"گزارش تخلف در توزیع کتب درسی و اقلام اداری اداره کل",
	}

	var adminUsers []string
	rows, err := db.Query("SELECT username FROM admin_users")
	if err == nil {
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err == nil {
				adminUsers = append(adminUsers, u)
			}
		}
		rows.Close()
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	insertedCount := 0

	for i := 1; i <= 50; i++ {
		trackingCode := fmt.Sprintf("PAG%02d%04d", (i*7)%90+10, i*123%9000+1000)
		cat := categories[(i-1)%len(categories)]
		prio := priorities[(i-1)%len(priorities)]
		class := classifications[(i-1)%len(classifications)]
		subj := sampleSubjects[(i-1)%len(sampleSubjects)]

		descPlain := fmt.Sprintf("%s. این گزارش تستی شماره %d جهت بررسی ناوبری و صفحه‌بندی سامانه ثبت شده است. کلیه جزئیات مرتبط با تاریخچه و گردش کار در سامانه ذخیره می‌گردد.", subj, i)
		encDesc, err := encryptWithKey(keyFields, "v2:f:", descPlain)
		if err != nil {
			log.Fatalf("Encryption failed: %v", err)
		}

		var encName, encPhone sql.NullString
		reportType := "anonymous"
		if i%2 == 0 {
			reportType = "identified"
			namePlain := fmt.Sprintf("گزارش‌دهنده تستی %d", i)
			encN, _ := encryptWithKey(keyFields, "v2:f:", namePlain)
			encName = sql.NullString{String: encN, Valid: true}

			phonePlain := fmt.Sprintf("0912%07d", 1000000+i)
			encP, _ := encryptWithKey(keyFields, "v2:f:", phonePlain)
			encPhone = sql.NullString{String: encP, Valid: true}
		}

		hasNewMessage := (i%4 == 0)
		var encResponse sql.NullString
		var respBy sql.NullString
		var respAt *time.Time
		if i%3 == 0 {
			respPlain := fmt.Sprintf("پاسخ حراست برای پرونده شماره %d: گزارش بررسی شد و دستور پیگیری صادر گردید.", i)
			encR, _ := encryptWithKey(keyFields, "v2:f:", respPlain)
			encResponse = sql.NullString{String: encR, Valid: true}
			respBy = sql.NullString{String: "admin_herasat", Valid: true}
			nowTime := time.Now().Add(-time.Duration(51-i) * time.Hour)
			respAt = &nowTime
		}

		createdAt := time.Now().Add(-time.Duration(51-i) * 2 * time.Hour)
		updatedAt := createdAt
		if respAt != nil {
			updatedAt = *respAt
		}

		var reportID int64
		err = tx.QueryRow(`
			INSERT INTO reports (
				tracking_code, category, report_type, priority,
				reporter_name, reporter_phone, description,
				admin_response, responded_by, responded_at,
				has_new_message, classification,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7,
				$8, $9, $10,
				$11, $12,
				$13, $14
			) RETURNING id
		`,
			trackingCode, cat, reportType, prio,
			encName, encPhone, encDesc,
			encResponse, respBy, respAt,
			hasNewMessage, class,
			createdAt, updatedAt,
		).Scan(&reportID)

		if err != nil {
			log.Fatalf("Failed to insert report %d: %v", i, err)
		}

		if i%2 == 1 {
			for _, u := range adminUsers {
				_, _ = tx.Exec(`
					INSERT INTO admin_report_views (report_id, admin_username, viewed_at)
					VALUES ($1, $2, $3)
					ON CONFLICT (report_id, admin_username) DO NOTHING
				`, reportID, u, createdAt.Add(time.Minute*15))
			}
		}

		insertedCount++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Transaction commit failed: %v", err)
	}

	fmt.Printf("Successfully inserted %d test reports.\n", insertedCount)
}
