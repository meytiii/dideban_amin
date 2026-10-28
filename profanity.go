package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type compiledTerm struct {
	raw              string
	clean            string
	collapsed        string
	phonetic         string
	squashed         string
	squashedPhonetic string
	isExactSade      bool
}

var (
	persianProfanities = []string{
		"کیر", "کسکش", "کس کش", "کوسکش", "کوس کش", "کصکش", "کص کش", "کوصکش", "کوص کش",
		"کونی", "کوص", "کص", "کوس", "کون", "جنده", "قحبه",
		"مادرجنده", "مادر جنده", "خواهرجنده", "خواهر جنده", "پدر سگ", "پدرسگ", "پدرصلواتی",
		"دیوث", "دیوس", "بی ناموس", "بیناموس", "حرامزاده", "حرومزاده", "حروم زاده", "حرام زاده",
		"سیکتیر", "سیک تیر", "گوه", "عن", "لاشی", "لاشخور", "پفیوز", "کثافت", "بی شرف",
		"بیشرف", "مادرقحبه", "مادر قحبه", "مادر قهوه", "مادرقهوه", "کیرم", "کیرت", "کیری",
		"کسخل", "کس خل", "کصخل", "کص خل", "کسمغز", "کس مغز", "کصمغز", "کص مغز",
		"کسشعر", "کس شعر", "کسشر", "کس شر", "کصشعر", "کص شعر", "کصشر", "کص شر", "کصش",
		"گه نخور", "خفه شو", "کره خر", "تخم سگ", "تخمسگ", "گاایدم", "گاییدم", "میگام", "بگام",
		"ممه", "کوندوز", "کونده", "اوبی", "بی پدر", "بیپدر", "پدرسوخته", "پدر سوخته",
		"کسلیس", "کصلیس", "کونکش", "کون کش",
	}

	finglishProfanities = []string{
		"kos", "koss", "kose", "koseh", "koskesh", "koskesha", "koskeshe", "koskhol", "koskhel",
		"kir", "keer", "kiri", "kiram", "kiret", "kirom", "kiresh", "koon", "koone",
		"kooni", "kooneh", "jende", "jendeh", "jendegi", "madarjende", "madar jende", "madarjendeh",
		"khaharjende", "khahar jendeh", "dayoos", "dayus", "dayous", "dayoose", "dayoosa",
		"binamoos", "bi namoos", "benamoos", "haroomzade", "haromzade", "haroomzadeh", "haramzadeh", "harami",
		"siktir", "sikter", "sik", "seekteer", "seekter", "lashi", "lashkhor", "pofiyooz", "pofyuz",
		"kesafat", "kasafat", "kesafatha", "bisharaf", "bi sharaf", "madarghahbe", "ghahbe", "ghahbeh",
		"pedarsag", "pedar sag", "pedarsookhte", "pedar sookhte", "goh", "gohkhor", "gohkhori",
		"gaeedam", "gayidam", "gayedam", "gayeede", "begam", "obne", "obneh", "obi",
	}

	englishProfanities = []string{
		"fuck", "fucking", "fucker", "fck", "fuk", "shit", "shitty", "bitch", "bitches", "asshole",
		"ass", "bastard", "cunt", "motherfucker", "dick", "pussy", "whore", "slut", "faggot", "fag",
		"nigger", "nigga", "cock", "blowjob", "prick", "twat", "wanker",
	}

	compiledPersian  []compiledTerm
	compiledFinglish []compiledTerm
	compiledEnglish  []compiledTerm

	slurRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ر(?:ی|م|ت|تون|مون|ش|ها)?(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[یi*]+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ش(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[وo0u*]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ک?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+[صس](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*خ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ل`),
		regexp.MustCompile(`(?i)ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[ع*]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ن[یi*]?(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هa*](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ق[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[حه]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هa*](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[ثسص](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])س[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[تط][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])پ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*س[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*گ(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[سص](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ف(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])[حه][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[اوu]*[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[زذضظ][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هa*](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هa*](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])خ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ه[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هa*](?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])گ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])ل[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[یi*]+(?:[^\p{L}\p{N}]|$)`),
		regexp.MustCompile(`(?i)\bk+[iey]+r+\b`),
		regexp.MustCompile(`(?i)\bk+[ou]+s+\b`),
		regexp.MustCompile(`(?i)\bk+o{2,}n+\b`),
		regexp.MustCompile(`(?i)\bf+[\s*._~^#@$%&/\\|!?\-+=0-9]*[u*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*[c*]*[\s*._~^#@$%&/\\|!?\-+=0-9]*k+\b`),
		regexp.MustCompile(`(?i)\bs+[\s*._~^#@$%&/\\|!?\-+=0-9]*[h*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*[i*!1]+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+\b`),
		regexp.MustCompile(`(?i)\bb+[\s*._~^#@$%&/\\|!?\-+=0-9]*[i*!1]+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+[\s*._~^#@$%&/\\|!?\-+=0-9]*c+[\s*._~^#@$%&/\\|!?\-+=0-9]*h+\b`),
		regexp.MustCompile(`(?i)\bc+[\s*._~^#@$%&/\\|!?\-+=0-9]*[u*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*n+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+\b`),
		regexp.MustCompile(`(?i)\ba+[\s*._~^#@$%&/\\|!?\-+=0-9]*s+[\s*._~^#@$%&/\\|!?\-+=0-9]*s+[\s*._~^#@$%&/\\|!?\-+=0-9]*h+[\s*._~^#@$%&/\\|!?\-+=0-9]*o+[\s*._~^#@$%&/\\|!?\-+=0-9]*l+[\s*._~^#@$%&/\\|!?\-+=0-9]*e+\b`),
		regexp.MustCompile(`(?i)\bj+e+n+d+e+h?`),
		regexp.MustCompile(`(?i)\bd+a+y+[ou]+s+`),
		regexp.MustCompile(`(?i)\bk+o+s+k+e+s+h+`),
	}
)

func init() {
	compileList := func(list []string, isPersian bool) []compiledTerm {
		out := make([]compiledTerm, 0, len(list))
		for _, raw := range list {
			clean := stripSymbolsAndDiacritics(strings.ToLower(raw))
			collapsed := collapseRepeatedRunes(clean)
			squashed := strings.ReplaceAll(collapsed, " ", "")
			ct := compiledTerm{
				raw:       raw,
				clean:     clean,
				collapsed: collapsed,
				squashed:  squashed,
			}
			if isPersian {
				// Prevent standalone "کص" or "کوص" from phonetically matching normal "کس" (person)
				if squashed == "کص" || squashed == "کوص" || squashed == "کصش" {
					ct.isExactSade = true
					ct.phonetic = collapsed
					ct.squashedPhonetic = squashed
				} else {
					ct.phonetic = toPhoneticPersian(collapsed)
					ct.squashedPhonetic = toPhoneticPersian(squashed)
				}
			}
			out = append(out, ct)
		}
		return out
	}

	compiledPersian = compileList(persianProfanities, true)
	compiledFinglish = compileList(finglishProfanities, false)
	compiledEnglish = compileList(englishProfanities, false)
}

func cleanRune(r rune) rune {
	switch r {
	case 'ي', 'ئ', 'ى', 'ے', 'ۍ', 'ې':
		return 'ی'
	case 'ك', 'ڪ':
		return 'ک'
	case 'ة', 'ۀ', 'ھ', 'ە':
		return 'ه'
	case 'ؤ':
		return 'و'
	case 'أ', 'إ', 'آ', 'ٱ':
		return 'ا'
	case '@', '4':
		return 'a'
	case '$', '5':
		return 's'
	case '0':
		return 'o'
	case '1', '!':
		return 'i'
	case '3':
		return 'e'
	}
	return r
}

func isDiacriticOrInvisible(r rune) bool {
	if (r >= 0x064B && r <= 0x065F) || r == 0x0640 || r == 0x0670 {
		return true
	}
	if (r >= 0x200B && r <= 0x200F) || r == 0xFEFF || r == 0x00AD {
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

func preCleanForRegex(s string) string {
	var b strings.Builder
	for _, r := range s {
		if isDiacriticOrInvisible(r) {
			continue
		}
		switch r {
		case 'ي', 'ئ', 'ى', 'ے':
			r = 'ی'
		case 'ك', 'ڪ':
			r = 'ک'
		case 'ة', 'ۀ', 'ھ', 'ە':
			r = 'ه'
		case 'أ', 'إ', 'آ', 'ٱ':
			r = 'ا'
		}
		b.WriteRune(r)
	}
	return b.String()
}

func toPhoneticPersian(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'ص', 'ث':
			b.WriteRune('س')
		case 'ذ', 'ض', 'ظ':
			b.WriteRune('ز')
		case 'ط':
			b.WriteRune('ت')
		case 'غ':
			b.WriteRune('ق')
		case 'ح':
			b.WriteRune('ه')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func containsBoundedPhrase(text, term string) bool {
	if term == "" {
		return false
	}
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

func hasVulgarPersianSuffix(token, termClean string) bool {
	if !strings.HasPrefix(token, termClean) {
		return false
	}
	suffix := token[len(termClean):]
	// Avoid false positives on "عن" ("عنان", "عنه", "عنها")
	if termClean == "عن" {
		switch suffix {
		case "", "م", "ت", "ش", "ی", "تر":
			return true
		}
		return false
	}
	switch suffix {
	case "", "ها", "هاش", "های", "ه", "ی", "م", "ت", "ش", "مان", "تان", "شان", "تون", "مون", "ان", "ا":
		return true
	}
	return false
}

func hasVulgarLatinSuffix(token, term string) bool {
	if !strings.HasPrefix(token, term) {
		return false
	}
	suffix := token[len(term):]
	switch suffix {
	case "", "s", "es", "ed", "ing", "er", "ers", "y", "e", "eh", "a", "ha", "am", "et", "esh", "i", "o", "oo":
		return true
	}
	return false
}

func matchSingleToken(stripped, collapsed, collapsedPhonetic string) (bool, string) {
	if stripped == "" {
		return false, ""
	}

	for _, ct := range compiledPersian {
		if strings.Contains(ct.clean, " ") {
			continue
		}
		if stripped == ct.clean || collapsed == ct.collapsed {
			return true, ct.raw
		}
		if hasVulgarPersianSuffix(stripped, ct.clean) || hasVulgarPersianSuffix(collapsed, ct.collapsed) {
			return true, ct.raw
		}
		if !ct.isExactSade {
			if collapsedPhonetic == ct.phonetic || hasVulgarPersianSuffix(collapsedPhonetic, ct.phonetic) {
				return true, ct.raw
			}
		}
	}

	for _, ct := range compiledFinglish {
		if strings.Contains(ct.clean, " ") {
			continue
		}
		// Require double 'o' on "koon"/"kooni" so normal Finglish "kone"/"koni" is not blocked
		if ct.clean == "koon" || ct.clean == "kooni" || ct.clean == "koone" || ct.clean == "kooneh" {
			if stripped == ct.clean || (len(ct.clean) >= 4 && hasVulgarLatinSuffix(stripped, ct.clean)) {
				return true, ct.raw
			}
			continue
		}
		if stripped == ct.clean || collapsed == ct.collapsed {
			return true, ct.raw
		}
		if len(ct.collapsed) >= 4 && (hasVulgarLatinSuffix(stripped, ct.clean) || hasVulgarLatinSuffix(collapsed, ct.collapsed)) {
			return true, ct.raw
		}
	}

	for _, ct := range compiledEnglish {
		if stripped == ct.clean || collapsed == ct.collapsed {
			return true, ct.raw
		}
		if len(ct.collapsed) >= 4 && (hasVulgarLatinSuffix(stripped, ct.clean) || hasVulgarLatinSuffix(collapsed, ct.collapsed)) {
			return true, ct.raw
		}
	}

	return false, ""
}

func ContainsVulgarLanguage(text string) (bool, string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false, ""
	}

	lowered := strings.ToLower(trimmed)
	regexReady := preCleanForRegex(lowered)

	for _, re := range slurRegexes {
		if re.MatchString(regexReady) {
			return true, "obfuscated_slur"
		}
	}

	rawTokens := strings.Fields(lowered)
	cleanedTokens := make([]string, 0, len(rawTokens))
	var singleCharSeq strings.Builder

	flushSingleCharSeq := func() (bool, string) {
		if singleCharSeq.Len() == 0 {
			return false, ""
		}
		seqStr := singleCharSeq.String()
		singleCharSeq.Reset()
		seqCollapsed := collapseRepeatedRunes(seqStr)
		seqPhonetic := toPhoneticPersian(seqCollapsed)
		return matchSingleToken(seqStr, seqCollapsed, seqPhonetic)
	}

	for _, token := range rawTokens {
		stripped := stripSymbolsAndDiacritics(token)
		if stripped == "" {
			continue
		}
		cleanedTokens = append(cleanedTokens, stripped)

		collapsed := collapseRepeatedRunes(stripped)
		collapsedPhonetic := toPhoneticPersian(collapsed)

		if matched, term := matchSingleToken(stripped, collapsed, collapsedPhonetic); matched {
			return true, term
		}

		runes := []rune(stripped)
		if len(runes) == 1 {
			singleCharSeq.WriteRune(runes[0])
		} else {
			if matched, term := flushSingleCharSeq(); matched {
				return true, term
			}
		}
	}

	if matched, term := flushSingleCharSeq(); matched {
		return true, term
	}

	// Sliding window (2 to 4 adjacent tokens) to catch split-word bypasses like "مادر ج نده" or "b i namoos"
	for i := 0; i < len(cleanedTokens); i++ {
		var windowBuilder strings.Builder
		windowBuilder.WriteString(cleanedTokens[i])
		for j := i + 1; j < len(cleanedTokens) && j <= i+3; j++ {
			windowBuilder.WriteString(cleanedTokens[j])
			joined := windowBuilder.String()
			joinedCollapsed := collapseRepeatedRunes(joined)
			joinedPhonetic := toPhoneticPersian(joinedCollapsed)

			for _, ct := range compiledPersian {
				if joined == ct.squashed || joinedCollapsed == ct.squashed {
					return true, ct.raw
				}
				if !ct.isExactSade && joinedPhonetic == ct.squashedPhonetic {
					return true, ct.raw
				}
			}
			for _, ct := range compiledFinglish {
				if ct.clean == "koon" || ct.clean == "kooni" || ct.clean == "koone" || ct.clean == "kooneh" {
					continue
				}
				if len(ct.squashed) >= 5 && (joined == ct.squashed || joinedCollapsed == ct.squashed) {
					return true, ct.raw
				}
			}
			for _, ct := range compiledEnglish {
				if len(ct.squashed) >= 4 && (joined == ct.squashed || joinedCollapsed == ct.squashed) {
					return true, ct.raw
				}
			}
		}
	}

	// Normalized sentence check for multi-word phrases
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
	collapsedSentence := collapseRepeatedRunes(normalizedSentence)
	phoneticSentence := toPhoneticPersian(collapsedSentence)

	for _, ct := range compiledPersian {
		if strings.Contains(ct.clean, " ") {
			if containsBoundedPhrase(normalizedSentence, ct.clean) ||
				containsBoundedPhrase(collapsedSentence, ct.collapsed) ||
				containsBoundedPhrase(collapsedSentence, ct.squashed) ||
				(!ct.isExactSade && containsBoundedPhrase(phoneticSentence, ct.phonetic)) {
				return true, ct.raw
			}
		}
	}

	for _, ct := range compiledFinglish {
		if strings.Contains(ct.clean, " ") {
			if containsBoundedPhrase(normalizedSentence, ct.clean) ||
				containsBoundedPhrase(collapsedSentence, ct.collapsed) ||
				containsBoundedPhrase(collapsedSentence, ct.squashed) {
				return true, ct.raw
			}
		}
	}

	return false, ""
}

const VulgarityLegalWarningMessage = "اخطار حقوقی و امنیتی: متن ارسالی حاوی الفاظ نامتعارف، توهین‌آمیز یا مصادیق توهین به اشخاص و کارکنان دولت می‌باشد. بر اساس ماده ۶۰۸ و ۶۰۹ قانون مجازات اسلامی و مقررات جرایم رایانه‌ای، هرگونه توهین، افترا و هتک حیثیت در سامانه‌های رسمی کشور، پیگرد قانونی به همراه داشته و مشخصات فنی مرتکب جهت رسیدگی به مراجع ذی‌صلاح قضایی و امنیتی منعکس خواهد گردید. لطفاً گزارش خود را با رعایت موازین قانونی و ادب بازنویسی فرمایید."

func GetProfanityRulesHandler(w http.ResponseWriter, r *http.Request) {
	etag := `"herasat-profanity-rules-v1"`
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := map[string]interface{}{
		"persian_profanities":  persianProfanities,
		"finglish_profanities": finglishProfanities,
		"english_profanities":  englishProfanities,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
