package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	persianProfanities = []string{
		"کیر", "کسکش", "کس کش", "کوسکش", "کوس کش", "کصکش", "کص کش", "کوصکش", "کوص کش",
		"کونی", "کوص", "کص", "کوس", "کون", "جنده", "قحبه",
		"مادرجنده", "مادر جنده", "خواهرجنده", "خواهر جنده", "پدر سگ", "پدرسگ", "پدرصلواتی",
		"دیوث", "بی ناموس", "بی‌ناموس", "بیناموس", "حرامزاده", "حرومزاده", "حروم زاده", "حرام زاده",
		"سیکتیر", "سیک تیر", "گوه", "عن", "لاشی", "لاشخور", "پفیوز", "کثافت", "بی شرف", "بی‌شرف",
		"بیشرف", "مادرقحبه", "مادر قحبه", "مادر قهوه", "مادرقهوه", "کیرم", "کیرت", "کیری",
		"کسخل", "کس خل", "کصخل", "کص خل", "کسمغز", "کس مغز", "کصمغز", "کص مغز",
		"کسشعر", "کس شعر", "کسشر", "کس شر", "کصشعر", "کص شعر", "کصشر", "کص شر", "کصش", "کسش",
		"گه نخور", "خفه شو", "کره خر", "تخم سگ", "تخمسگ", "گاایدم", "گاییدم", "میگام", "بگام",
		"ممه", "کوندوز", "کونده", "اوبی", "بی پدر", "بی‌پدر", "پدرسوخته", "پدر سوخته",
		"کسلیس", "کصلیس", "کونکش", "کون کش",
	}

	finglishProfanities = []string{
		"kos", "koss", "kose", "koseh", "koskesh", "koskesha", "koskeshe", "koskhol", "koskhel",
		"kir", "keer", "kiri", "kiram", "kiret", "kirom", "kiresh", "kire", "koon", "kone", "koone",
		"kooni", "kooneh", "koni", "jende", "jendeh", "jendegi", "madarjende", "madar jende", "madarjendeh",
		"khaharjende", "khahar jendeh", "dayoos", "dayus", "dayous", "dayoose", "dayoosa",
		"binamoos", "bi namoos", "benamoos", "haroomzade", "haromzade", "haroomzadeh", "haramzadeh", "harami",
		"siktir", "sikter", "sik", "seekteer", "seekter", "lashi", "lashkhor", "pofiyooz", "pofyuz",
		"kesafat", "kasafat", "kesafatha", "bisharaf", "bi sharaf", "madarghahbe", "ghahbe", "ghahbeh",
		"pedarsag", "pedar sag", "pedarsookhte", "pedar sookhte", "goh", "gohkhor", "gohkhori",
		"gaeedam", "gayidam", "gayedam", "gayeede", "migam", "begam", "obne", "obneh", "obi",
	}

	englishProfanities = []string{
		"fuck", "fucking", "fucker", "fck", "fuk", "shit", "shitty", "bitch", "bitches", "asshole",
		"ass", "bastard", "cunt", "motherfucker", "dick", "pussy", "whore", "slut", "faggot", "fag",
		"nigger", "nigga", "cock", "blowjob", "prick", "twat", "wanker",
	}

	// Compiled regex patterns for obfuscated, symbol-injected, or wildcard masked vulgarities
	slurRegexes = []*regexp.Regexp{
		// Masked 'کیر' like ک*ر, ک..ر, ک_ر, ک-ر, ک#ر, ک!ر, ک*رم, ک*ری
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ر(?:ی|م|ت|تون|مون|ش|ها)?(?:[^\p{L}\p{N}]|$)`),
		// 'کیر' with arbitrary symbols or spaces between letters (ک*ی*ر, ک.ی.ر, ک_ی_ر, etc.)
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[یiي*]+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر`),
		// Masked 'کسکش' / 'کصکش' / 'کسشعر' like ک*ص*ش, ک*س*ش, کص*ش, کس*ش
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ش(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'کسکش' / 'کصکش' / 'کوسکش' like ک*ص*ک*ش, ک*س*ک*ش, ک*صکش, ک*سکش, کص*کش, کس*کش
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[وo0u*]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ک?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش`),
		// Standalone masked 'کص' / 'کس' like ک*ص, ک*س, ک..ص, ک_س
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+[صس](?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'کسخل' / 'کصخل' like ک*ص*خ*ل, ک*صخل, ک*سخل
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*خ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ل`),
		// Obfuscated 'کسشعر' / 'کصشعر' / 'کسشر' / 'کصشر'
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[ع*]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر`),
		// Masked 'کون' / 'کونی' like ک*ن, ک*نی, ک..ن, ک_نی
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ن[یiي*]?(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'جنده' / 'ج*ده' / 'ج*نده' / 'ج*ن*ده'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*](?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'قحبه' / 'ق*حبه' / 'ق*ح*به' / 'ق*به'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ق[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ح?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*](?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'دیوث' / 'د*وث' / 'د*یوث' / 'د*ی*وث'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ث(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'سیکتیر' / 'س*کتیر' / 'س*ی*ک*ت*ی*ر'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])س[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ت[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'پدرسگ' / 'پ*درسگ' / 'پ*د*ر*س*گ'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])پ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*س[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*گ(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'بیناموس' / 'بی ناموس' / 'ب*ناموس' / 'ب*ی*ناموس'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*س(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'بیشرف' / 'بی شرف' / 'ب*شرف'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ف(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'حرامزاده' / 'حرومزاده' / 'ح*رامزاده'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ح[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[اوu]*[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ز[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*](?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'مادرجنده' / 'م*ادرجنده' / 'مادر*نده'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*](?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'خواهرجنده'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])خ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ه[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*](?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'گاییدم' / 'گ*اییدم' / 'گ*ییدم'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])گ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م(?:[^\p{L}\p{N}]|$)`),
		// Obfuscated 'لاشی' / 'ل*شی'
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ل[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[یiي*]+(?:[^\p{L}\p{N}]|$)`),

		// Finglish & English obfuscations
		regexp.MustCompile(`(?i)\bk+[i|e|y]+r+\b`),
		regexp.MustCompile(`(?i)\bk+[o|u]+s+\b`),
		regexp.MustCompile(`(?i)\bk+[o|u]+n+\b`),
		regexp.MustCompile(`(?i)\bf+[\s*._~^#@$%&/\\|!?\-+=0-9]*[u*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*[c*]*[\s*._~^#@$%&/\\|!?\-+=0-9]*k+\b`),
		regexp.MustCompile(`(?i)\bs+[\s*._~^#@$%&/\\|!?\-+=0-9]*[h*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*[i*!1]+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+\b`),
		regexp.MustCompile(`(?i)\bb+[\s*._~^#@$%&/\\|!?\-+=0-9]*[i*!1]+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+[\s*._~^#@$%&/\\|!?\-+=0-9]*c+[\s*._~^#@$%&/\\|!?\-+=0-9]*h+\b`),
		regexp.MustCompile(`(?i)\bc+[\s*._~^#@$%&/\\|!?\-+=0-9]*[u*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*n+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+\b`),
		regexp.MustCompile(`(?i)\ba+[\s*._~^#@$%&/\\|!?\-+=0-9]*s+[\s*._~^#@$%&/\\|!?\-+=0-9]*s+[\s*._~^#@$%&/\\|!?\-+=0-9]*h+[\s*._~^#@$%&/\\|!?\-+=0-9]*o+[\s*._~^#@$%&/\\|!?\-+=0-9]*l+[\s*._~^#@$%&/\\|!?\-+=0-9]*e+\b`),
		regexp.MustCompile(`(?i)\bj+e+n+d+e+h?`),
		regexp.MustCompile(`(?i)\bd+a+y+[o|u]+s+`),
		regexp.MustCompile(`(?i)\bk+o+s+k+e+s+h+`),
	}
)

func cleanRune(r rune) rune {
	switch r {
	case 'ي', 'ئ':
		return 'ی'
	case 'ك':
		return 'ک'
	case 'ة':
		return 'ه'
	case 'ؤ':
		return 'و'
	case '@':
		return 'a'
	case '$':
		return 's'
	case '0':
		return 'o'
	case '1', '!':
		return 'i'
	case '3':
		return 'e'
	case '4':
		return 'a'
	case '5':
		return 's'
	}
	return r
}

func isDiacriticOrInvisible(r rune) bool {
	// Arabic/Persian diacritics (Fatha, Damma, Kasra, Tashdid, Tanwin, Sukun) & Tatweel
	if (r >= 0x064B && r <= 0x065F) || r == 0x0640 || r == 0x0670 {
		return true
	}
	// Invisible markers, ZWNJ, ZWSP, BOM
	if r == 0x200C || r == 0x200B || r == 0xFEFF || r == 0x200E || r == 0x200F {
		return true
	}
	return false
}

func isSymbolOrPunct(r rune) bool {
	if unicode.IsPunct(r) || unicode.IsSymbol(r) {
		return true
	}
	switch r {
	case '*', '.', '_', '-', '+', '=', '~', '^', '#', '@', '$', '%', '&', '/', '\\', '|', '!', '?', ',', ':', ';', '"', '\'', '«', '»', '‐', '–', '—', '…', '·', '•', '×', '÷':
		return true
	}
	return false
}

func collapseRepeatedRunes(s string) string {
	var b strings.Builder
	var last rune
	for _, r := range s {
		if r != last {
			b.WriteRune(r)
			last = r
		}
	}
	return b.String()
}

func stripSymbolsAndDiacritics(s string) string {
	var b strings.Builder
	for _, r := range s {
		r = cleanRune(r)
		if isDiacriticOrInvisible(r) || isSymbolOrPunct(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func toPhoneticPersian(s string) string {
	return strings.ReplaceAll(s, "ص", "س")
}

func containsBoundedPhrase(text, term string) bool {
	idx := 0
	for {
		i := strings.Index(text[idx:], term)
		if i == -1 {
			return false
		}
		matchStart := idx + i
		matchEnd := matchStart + len(term)

		validBefore := true
		if matchStart > 0 {
			rBefore, _ := utf8.DecodeLastRuneInString(text[:matchStart])
			if unicode.IsLetter(rBefore) || unicode.IsDigit(rBefore) {
				validBefore = false
			}
		}

		validAfter := true
		if matchEnd < len(text) {
			rAfter, _ := utf8.DecodeRuneInString(text[matchEnd:])
			if unicode.IsLetter(rAfter) || unicode.IsDigit(rAfter) {
				validAfter = false
			}
		}

		if validBefore && validAfter {
			return true
		}

		idx = matchStart + 1
		if idx >= len(text) {
			break
		}
	}
	return false
}

func hasVulgarPrefix(token, termClean string) bool {
	if strings.HasPrefix(token, termClean) {
		suffix := token[len(termClean):]
		switch suffix {
		case "", "ها", "هاش", "های", "ه", "ی", "م", "ت", "ش", "مان", "تان", "شان", "تون", "مون", "ان", "ا":
			return true
		}
	}
	return false
}

func ContainsVulgarLanguage(text string) (bool, string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false, ""
	}

	lowered := strings.ToLower(trimmed)

	// 1. Check Regex Patterns (Catches wildcard replacements like ک*ر, ک*ص*ش, and symbols between letters)
	for _, re := range slurRegexes {
		if re.MatchString(lowered) {
			return true, "obfuscated_slur"
		}
	}

	// 2. Token-level analysis (handles internal symbols, repeated stretching like کییییییر, ک*ی*ر, ک.ص.ک.ش)
	rawTokens := strings.Fields(lowered)
	var singleCharSeq strings.Builder

	for _, token := range rawTokens {
		stripped := stripSymbolsAndDiacritics(token)
		collapsed := collapseRepeatedRunes(stripped)
		collapsedPhonetic := toPhoneticPersian(collapsed)

		// Check against Persian profanities
		for _, term := range persianProfanities {
			termClean := stripSymbolsAndDiacritics(term)
			termPhonetic := toPhoneticPersian(termClean)

			if stripped == termClean || collapsed == termClean || collapsedPhonetic == termPhonetic {
				return true, term
			}

			// Check if token starts with vulgar word + valid suffix (e.g. کسکش‌ها, کونی‌ها, لاشی‌ها)
			if hasVulgarPrefix(stripped, termClean) || hasVulgarPrefix(collapsed, termClean) || hasVulgarPrefix(collapsedPhonetic, termPhonetic) {
				return true, term
			}
		}

		// Check against Finglish profanities
		for _, term := range finglishProfanities {
			if stripped == term || collapsed == term {
				return true, term
			}
			if len(term) >= 4 && (strings.HasPrefix(stripped, term) || strings.HasPrefix(collapsed, term)) {
				return true, term
			}
		}

		// Check against English profanities
		for _, term := range englishProfanities {
			if stripped == term || collapsed == term {
				return true, term
			}
			if len(term) >= 4 && (strings.HasPrefix(stripped, term) || strings.HasPrefix(collapsed, term)) {
				return true, term
			}
		}

		// Collect single-character runs for spaced-out tokens (e.g. "ک" "ی" "ر" or "ک" "ص" "ک" "ش")
		runes := []rune(stripped)
		if len(runes) == 1 {
			singleCharSeq.WriteRune(runes[0])
		} else {
			if singleCharSeq.Len() > 0 {
				seqStr := singleCharSeq.String()
				seqCollapsed := collapseRepeatedRunes(seqStr)
				seqPhonetic := toPhoneticPersian(seqCollapsed)
				for _, term := range persianProfanities {
					termClean := stripSymbolsAndDiacritics(term)
					termPhonetic := toPhoneticPersian(termClean)
					if seqStr == termClean || seqCollapsed == termClean || seqPhonetic == termPhonetic {
						return true, term
					}
				}
				singleCharSeq.Reset()
			}
		}
	}

	// Flush trailing single-character sequence if any
	if singleCharSeq.Len() > 0 {
		seqStr := singleCharSeq.String()
		seqCollapsed := collapseRepeatedRunes(seqStr)
		seqPhonetic := toPhoneticPersian(seqCollapsed)
		for _, term := range persianProfanities {
			termClean := stripSymbolsAndDiacritics(term)
			termPhonetic := toPhoneticPersian(termClean)
			if seqStr == termClean || seqCollapsed == termClean || seqPhonetic == termPhonetic {
				return true, term
			}
		}
	}

	// 3. Multi-word phrase matching with flexible spacing (e.g. "مادر جنده", "بی ناموس", "پدر سگ")
	var textBuilder strings.Builder
	for _, r := range lowered {
		r = cleanRune(r)
		if isDiacriticOrInvisible(r) {
			continue
		}
		if isSymbolOrPunct(r) || unicode.IsSpace(r) {
			textBuilder.WriteRune(' ')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			textBuilder.WriteRune(r)
		}
	}
	normalizedSentence := strings.Join(strings.Fields(textBuilder.String()), " ")

	for _, term := range persianProfanities {
		termClean := stripSymbolsAndDiacritics(term)
		termSquashed := strings.ReplaceAll(termClean, " ", "")
		termSquashedPhonetic := toPhoneticPersian(termSquashed)

		// Multi-word phrases (contain spaces like "مادر جنده", "خواهر جنده", "بی ناموس", "پدر سگ")
		if strings.Contains(term, " ") {
			if containsBoundedPhrase(normalizedSentence, termClean) ||
				containsBoundedPhrase(normalizedSentence, termSquashed) ||
				containsBoundedPhrase(toPhoneticPersian(normalizedSentence), termSquashedPhonetic) {
				return true, term
			}
		}
	}

	return false, ""
}

const VulgarityLegalWarningMessage = "اخطار حقوقی و امنیتی: متن ارسالی حاوی الفاظ نامتعارف، توهین‌آمیز یا مصادیق توهین به اشخاص و کارکنان دولت می‌باشد. بر اساس ماده ۶۰۸ و ۶۰۹ قانون مجازات اسلامی و مقررات جرایم رایانه‌ای، هرگونه توهین، افترا و هتک حیثیت در سامانه‌های رسمی کشور، پیگرد قانونی به همراه داشته و مشخصات فنی مرتکب جهت رسیدگی به مراجع ذی‌صلاح قضایی و امنیتی منعکس خواهد گردید. لطفاً گزارش خود را با رعایت موازین قانونی و ادب بازنویسی فرمایید."

// GetProfanityRulesHandler serves synchronized vulgarity detection lists with client caching (M-3)
func GetProfanityRulesHandler(w http.ResponseWriter, r *http.Request) {
	etag := `"herasat-profanity-rules-v1"`
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := map[string]interface{}{
		"persian_profanities":  persianProfanities,
		"finglish_profanities": finglishProfanities,
		"english_profanities":  englishProfanities,
	}
	_ = json.NewEncoder(w).Encode(resp)
}


