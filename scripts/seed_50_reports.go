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

func genTrackingCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("PAG%d", time.Now().UnixNano()%100000000)
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return "PAG" + string(b)
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
		"تخلفات مالی",
		"مسائل امنیتی",
		"تخلفات اداری",
		"تخلفات آموزشی",
		"پرسنلی",
		"دانش آموزی",
		"مدارس غیردولتی",
		"آموزش مجازی",
		"فضای مجازی",
		"سایر",
	}

	descriptions := []string{
		"احتراماً به استحضار اداره حراست می‌رساند، در روند تسویه اسناد مالی و فاکتورهای صادره جهت تجهیز کارگاه‌های آموزشی، ارقام غیرواقعی و مغایرت‌های فاحش مشاهده شده است. تقاضای بررسی کارشناسی و تطبیق صورت‌حساب‌های بانکی مورد استدعا می‌باشد.",
		"گزارش واصله حاکی از عدم رعایت پروتکل‌های حفاظت فیزیکی، نقص در عملکرد دوربین‌های نظارتی و نگهداری نامناسب کلیدهای بایگانی اسناد محرمانه آموزشگاه در ساعات غیراداری است. لطفاً اقدامات بازدارنده و مقتضی مبذول گردد.",
		"به استحضار می‌رساند در صدور ابلاغ‌های کارگزینی و تخصیص ساعات اضافه‌کار اداری، ضوابط قانونی و عدالت اداری رعایت نگردیده و تبعیض آشکار در انتصابات اداری مشاهده شده است. استدعای پیگیری و نظارت حراستی را دارد.",
		"در فرآیند نمره‌دهی و ثبت نمرات نهایی مستمر دانش‌آموزان در سامانه سیدا، تغییرات بدون مجوز و تخلف از ضوابط امتحانات توسط عوامل آموزشی صورت پذیرفته است. مستندات جهت رسیدگی اداره سنجش و حراست پیوست می‌باشد.",
		"درخواست رسیدگی به رفتار نامناسب و برخورد غیرحرفه‌ای و خارج از شئونات اداری با همکاران و مراجعین در واحد ستادی، که منجر به ایجاد نارضایتی و اخلال در خدمت‌رسانی مطلوب شده است. تقاضای تذکر جدی و بررسی حراستی داریم.",
		"گزارش در خصوص عدم رعایت ضوابط ایمنی در تردد سرویس‌های ایاب و ذهاب دانش‌آموزی و استفاده از خودروهای فاقد معاینه فنی که موجب مخاطره سلامت فرزندان گردیده است. تقاضای نظارت و الزام شرکت پیمانکار به رفع نقص وجود دارد.",
		"دریافت وجوه مازاد بر تعرفه ابلاغی مصوب آموزش و پرورش تحت عنوان کلاس‌های فوق‌برنامه اجباری و عدم شفافیت در قرارداد با اولیای محترم دانش‌آموزان در این واحد آموزشی. تقاضای اعزام بازرس و عودت مبالغ مازاد را داریم.",
		"مشکلات مکرر در سامانه‌های آموزش برخط، عدم دسترسی پایدار کاربران و قطع مکرر کلاس‌ها که موجب سردرگمی دانش‌آموزان و خانواده‌ها گردیده است. تقاضای بازنگری در زیرساخت فناوری و نظارت بر نحوه ارائه آموزش‌ها را داریم.",
		"ایجاد گروه‌ها و کانال‌های غیررسمی با سوءاستفاده از عنوان مدرسه در بستر شبکه‌های اجتماعی و انتشار محتواهای نامناسب و غیرهمسو با ضوابط تربیتی. تقاضای مسدودسازی و شناسایی عوامل تشکیل‌دهنده از طریق مراجع مربوطه را دارد.",
		"گزارش در مورد نقص فنی مکرر در تأسیسات گرمایشی و سرمایشی آموزشگاه و عدم تخصیص بهنگام اقلام مصرفی بهداشتی که موجب اختلال در فعالیت‌های جاری گردیده است. استدعای پیگیری از مدیریت پشتیبانی و رفاهی را دارد.",
	}

	sampleNames := []string{
		"علی حسینی",
		"مریم رضایی",
		"محمد کریمی",
		"فاطمه ابراهیمی",
		"حسین مرادی",
		"زهرا قاسمی",
		"مهدی صادقی",
		"نرگس نجفی",
		"امیر موسوی",
		"سمیه رحیمی",
	}

	priorities := []string{
		"normal",
		"urgent",
		"critical",
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	insertedCount := 0

	for i := 1; i <= 50; i++ {
		trackingCode := genTrackingCode()
		catIdx := (i - 1) % len(categories)
		cat := categories[catIdx]
		descPlain := fmt.Sprintf("%s (شناسه پیگیری پرونده آزمایشی: %s)", descriptions[catIdx], trackingCode)
		prio := priorities[(i-1)%len(priorities)]

		encDesc, err := encryptWithKey(keyFields, "v2:f:", descPlain)
		if err != nil {
			log.Fatalf("Encryption failed: %v", err)
		}

		var encName, encPhone sql.NullString
		reportType := "anonymous"
		if i%2 == 0 {
			reportType = "identified"
			namePlain := sampleNames[(i-1)%len(sampleNames)]
			encN, _ := encryptWithKey(keyFields, "v2:f:", namePlain)
			encName = sql.NullString{String: encN, Valid: true}

			phonePlain := fmt.Sprintf("0918%07d", 1000000+i*31%900000)
			encP, _ := encryptWithKey(keyFields, "v2:f:", phonePlain)
			encPhone = sql.NullString{String: encP, Valid: true}
		}

		createdAt := time.Now().Add(-time.Duration(51-i) * time.Hour)
		updatedAt := createdAt

		var reportID int64
		err = tx.QueryRow(`
			INSERT INTO reports (
				tracking_code, category, report_type, priority,
				reporter_name, reporter_phone, description,
				admin_response, responded_by, responded_at,
				has_new_message, classification,
				is_read, read_at, read_by,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7,
				NULL, NULL, NULL,
				FALSE, NULL,
				FALSE, NULL, NULL,
				$8, $9
			) RETURNING id
		`,
			trackingCode, cat, reportType, prio,
			encName, encPhone, encDesc,
			createdAt, updatedAt,
		).Scan(&reportID)

		if err != nil {
			log.Fatalf("Failed to insert report %d: %v", i, err)
		}

		insertedCount++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Transaction commit failed: %v", err)
	}

	fmt.Printf("Successfully inserted %d untouched unread test reports.\n", insertedCount)
}
