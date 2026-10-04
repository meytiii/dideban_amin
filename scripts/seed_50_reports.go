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
	bytesArr := make([]byte, 16)
	if _, err := rand.Read(bytesArr); err != nil {
		panic("crypto/rand failure")
	}
	for i, b := range bytesArr {
		bytesArr[i] = charset[b%byte(len(charset))]
	}
	return string(bytesArr)
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
		"احتراماً به استحضار اداره حراست می‌رساند، در روند تسویه اسناد مالی و فاکتورهای صادره جهت تجهیز کارگاه‌های آموزشی هنرستان، ارقام غیرواقعی و مغایرت‌های فاحش مشاهده شده است. تقاضای بررسی کارشناسی و تطبیق صورت‌حساب‌های بانکی و اسناد مالی را داریم.",
		"گزارش دریافتی از اولیا حاکی از اخذ مبالغ اجباری خارج از فیش رسمی بانکی به بهانه کمک به مدرسه و هزینه‌های جاری در زمان ثبت‌نام و توزیع کارنامه است. تقاضای اعزام بازرسان حراست جهت رسیدگی و استرداد وجوه به این واحد آموزشی را داریم.",
		"بررسی دفاتر مالی و قراردادهای بوفه و سالن ورزشی نشان‌دهنده عدم واریز درآمدهای اختصاصی به حساب متمرکز دولتی آموزش و پرورش و واریز مبالغ به حساب‌های متفرقه شخصی می‌باشد. استدعای پیگیری فوری و حسابرسی دقیق را دارد.",
		"گزارش واصله حاکی از عدم رعایت پروتکل‌های حفاظت فیزیکی، نقص در عملکرد سیستم دوربین‌های مداربسته نظارتی و نگهداری نامناسب کلیدهای بایگانی اسناد محرمانه آموزشگاه در ساعات غیراداری است. لطفاً اقدامات بازدارنده و مقتضی مبذول گردد.",
		"مشاهده تردد افراد ناشناس و فاقد کارت شناسایی معتبر در محوطه داخلی مدرسه در ساعات آموزشی بدون هماهنگی با نگهبانی و مسئولین آموزشگاه، که موجب نگرانی کادر آموزشی و اولیا شده است. تقاضای تقویت تدابیر حراستی و تذکر اکید به مدیریت واحد را داریم.",
		"مفقود شدن تعدادی از مهرهای رسمی اداری و اوراق سربرگ‌دار بایگانی شده در بخش دبیرخانه که احتمال سوءاستفاده و جعل اسناد را افزایش می‌دهد. تقاضای ورود فوری کارشناسان حراست جهت ردیابی و پیشگیری از هرگونه تخلف احتمالی را دارد.",
		"به استحضار می‌رساند در صدور ابلاغ‌های کارگزینی و تخصیص ساعات اضافه‌کار اداری، ضوابط قانونی و عدالت اداری رعایت نگردیده و تبعیض آشکار در انتصابات اداری مشاهده شده است. استدعای پیگیری و نظارت حراستی را دارد.",
		"عدم حضور منظم و ترک محل خدمت توسط برخی از عوامل اجرایی در ساعات اداری که منجر به سردرگمی ارباب‌رجوع و معطل ماندن پرونده‌های خدمتی همکاران فرهنگی گردیده است. خواهشمند است موضوع مورد رسیدگی اداری قرار گیرد.",
		"استفاده غیرمجاز و شخصی از خودروها و تجهیزات نقلیه سازمانی در ساعات غیراداری و ایام تعطیل بدون مأموریت اداری مصوب. تقاضای بررسی سامانه‌های ردیاب خودرویی و نظارت دقیق‌تر بر واحد نقلیه اداره را داریم.",
		"در فرآیند نمره‌دهی و ثبت نمرات نهایی مستمر دانش‌آموزان در سامانه سیدا، تغییرات بدون مجوز و تخلف از ضوابط امتحانات توسط عوامل آموزشی صورت پذیرفته است. مستندات جهت رسیدگی اداره سنجش و حراست پیوست می‌باشد.",
		"عدم تشکیل منظم کلاس‌های درس در پایه دوازدهم و اجبار دانش‌آموزان به شرکت در کلاس‌های خصوصی با مبالغ گزاف خارج از مدرسه توسط دبیران مربوطه. استدعای پیگیری و برخورد قانونی با ناقضان آیین‌نامه‌های آموزشی را داریم.",
		"تخلف در نحوه برگزاری امتحانات هماهنگ نهایی و رعایت نکردن قرنطینه سوالات در حوزه امتحانی مربوطه که شائبه افشای سوالات پیش از آزمون را به همراه داشته است. تقاضای اعزام فوری ناظر مقیم از اداره را دارد.",
		"درخواست رسیدگی به رفتار نامناسب و برخورد غیرحرفه‌ای و خارج از شئونات اداری با همکاران و مراجعین در واحد ستادی، که منجر به ایجاد نارضایتی و اخلال در خدمت‌رسانی مطلوب شده است. تقاضای تذکر جدی و بررسی حراستی داریم.",
		"امتناع غیرموجه از رسیدگی به درخواست‌های قانونی نقل و انتقال همکاران و ایجاد موانع سلیقه‌ای بر خلاف بخشنامه جامع نقل و انتقالات فرهنگیان استان. خواهشمند است نسبت به رفع تبعیض اقدام لازم معمول گردد.",
		"گزارش در خصوص ادعای مدارک تحصیلی و دوره‌های ضمن خدمت نامعتبر ارائه شده توسط برخی کارکنان جهت ارتقای رتبه شغلی در طرح رتبه‌بندی معلمان. استدعای استعلام صحت‌سنجی از مراجع ذی‌صلاح صادرکننده را دارد.",
		"گزارش در خصوص عدم رعایت ضوابط ایمنی در تردد سرویس‌های ایاب و ذهاب دانش‌آموزی و استفاده از خودروهای فاقد معاینه فنی که موجب مخاطره سلامت فرزندان گردیده است. تقاضای نظارت و الزام شرکت پیمانکار به رفع نقص وجود دارد.",
		"بروز اختلافات مکرر و رفتارهای پرخطر و خشونت‌آمیز در زنگ‌های تفریح و غیبت مشاور تربیتی در آموزشگاه که نیازمند مداخله کارشناسی و تدابیر انضباطی پیشگیرانه می‌باشد.",
		"فروش اقلام خوراکی غیربهداشتی و تاریخ‌گذشته در پایگاه تغذیه سالم مدرسه بدون رعایت دستورالعمل‌های بهداشت و سلامت مدارس. استدعای بازدید میدانی کارشناسان بهداشت و حراست را دارد.",
		"دریافت وجوه مازاد بر تعرفه ابلاغی مصوب آموزش و پرورش تحت عنوان کلاس‌های فوق‌برنامه اجباری و عدم شفافیت در قرارداد با اولیای محترم دانش‌آموزان در این واحد آموزشی. تقاضای اعزام بازرس و عودت مبالغ مازاد را داریم.",
		"به‌کارگیری نیروهای آموزشی فاقد گزینش حراستی و عدم رعایت حداقل حقوق مصوب قانون کار برای معلمان آزاد شاغل در این مجموعه غیردولتی. تقاضای بازرسی و ساماندهی قراردادهای پرسنلی را داریم.",
		"تغییر کاربری غیراستاندارد ساختمان مسکونی به مدرسه بدون اخذ تاییدیه‌های نوسازی مدارس، استحکام بنا و استانداردهای آتش‌نشانی که جان دانش‌آموزان را تهدید می‌نماید. تقاضای توقف فعالیت تا ایمن‌سازی کامل را دارد.",
		"مشکلات مکرر در سامانه‌های آموزش برخط، عدم دسترسی پایدار کاربران و قطع مکرر کلاس‌ها که موجب سردرگمی دانش‌آموزان و خانواده‌ها گردیده است. تقاضای بازنگری در زیرساخت فناوری و نظارت بر نحوه ارائه آموزش‌ها را داریم.",
		"عدم بارگذاری سرفصل‌های تدریس مجازی طبق تقویم زمان‌بندی و دریافت شهریه دوره‌های هوشمندسازی بدون ارائه خدمات تعریف‌شده در سامانه شاد و پرتال آموزشگاه. تقاضای پیگیری و ارزیابی کیفیت آموزش را داریم.",
		"ایجاد گروه‌ها و کانال‌های غیررسمی با سوءاستفاده از عنوان مدرسه در بستر شبکه‌های اجتماعی و انتشار محتواهای نامناسب و غیرهمسو با ضوابط تربیتی. تقاضای مسدودسازی و شناسایی عوامل تشکیل‌دهنده از طریق مراجع مربوطه را دارد.",
		"انتشار تصاویر و فیلم‌های حریم خصوصی دانش‌آموزان و معلمان بدون رضایت در صفحات مجازی به قصد تمسخر یا تخریب جایگاه فرهنگیان. تقاضای برخورد قانونی و هماهنگی با پلیس فتا جهت حذف محتواها را دارد.",
		"گزارش در مورد نقص فنی مکرر در تأسیسات گرمایشی و سرمایشی آموزشگاه و عدم تخصیص بهنگام اقلام مصرفی بهداشتی که موجب اختلال در فعالیت‌های جاری گردیده است. استدعای پیگیری از مدیریت پشتیبانی و رفاهی را دارد.",
		"مشکلات ترافیکی شدید و نبود پل عابر پیاده یا گذرگاه امن در برابر درب ورودی مجتمع آموزشی که خطرات جانی روزمره برای دانش‌آموزان هنگام خروج از مدرسه ایجاد نموده است. تقاضای مکاتبه با شهرداری و راهور را دارد.",
	}

	sampleNames := []string{
		"علیرضا احمدی",
		"فاطمه رضایی",
		"محمدحسین مرادی",
		"زهرا کریمی",
		"مهدی سلیمانی",
		"نرگس باقری",
		"احسان محمدی",
		"مریم کاظمی",
		"سعید اکبری",
		"سمیرا رحیمی",
		"علی حسینی",
		"زهرا قاسمی",
		"حسین کرمی",
		"الهام موسوی",
		"امیرحسین نجفی",
	}

	phonePrefixes := []string{"0918", "0912", "0935", "0919", "0990"}

	now := time.Now()

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	insertedCount := 0

	for i := 1; i <= 50; i++ {
		trackingCode := genTrackingCode()
		cat := categories[(i-1)%len(categories)]
		descPlain := descriptions[(i-1)%len(descriptions)]

		prio := "normal"
		if i%5 == 0 {
			prio = "urgent"
		} else if i%10 == 0 {
			prio = "critical"
		}

		encDesc, err := encryptWithKey(keyFields, "v2:f:", descPlain)
		if err != nil {
			log.Fatalf("Encryption failed: %v", err)
		}

		var encName, encPhone sql.NullString
		reportType := "anonymous"
		if i%2 == 0 {
			reportType = "identified"
			namePlain := sampleNames[((i-1)/2)%len(sampleNames)]
			encN, _ := encryptWithKey(keyFields, "v2:f:", namePlain)
			encName = sql.NullString{String: encN, Valid: true}

			pref := phonePrefixes[(i-1)%len(phonePrefixes)]
			phonePlain := fmt.Sprintf("%s%07d", pref, 1000000+(i*3719)%9000000)
			encP, _ := encryptWithKey(keyFields, "v2:f:", phonePlain)
			encPhone = sql.NullString{String: encP, Valid: true}
		}

		hoursAgo := float64(50-i) * (720.0 / 50.0)
		createdAt := now.Add(-time.Duration(hoursAgo*float64(time.Hour)) - time.Duration((i*7)%59)*time.Minute)
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

		_, err = tx.Exec(`
			INSERT INTO admin_notifications (
				notif_type, title, message, target_report_id, target_tracking_code, is_read, created_at
			) VALUES (
				'new_report', 'گزارش جدید ثبت شد', $1, $2, $3, FALSE, $4
			)
		`, "موضوع: "+cat+" | شماره پرونده: "+trackingCode, reportID, trackingCode, createdAt)

		if err != nil {
			log.Fatalf("Failed to insert notification for report %d: %v", i, err)
		}

		insertedCount++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Transaction commit failed: %v", err)
	}

	fmt.Printf("Successfully inserted %d untouched unread test reports with matching tracking codes and categories.\n", insertedCount)
}
