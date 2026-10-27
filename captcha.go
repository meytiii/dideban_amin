package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CaptchaPayload struct {
	Answer    int   `json:"a"`
	ExpiresAt int64 `json:"e"`
}

type CaptchaResponse struct {
	CaptchaID   string `json:"captcha_id"`
	ImageBase64 string `json:"image_base64"`
}

type UsedCaptchaStore struct {
	mu       sync.Mutex
	db       *sql.DB
	initOnce sync.Once
	tokens   map[string]int64
	attempts map[string]int
}

func NewUsedCaptchaStore(db *sql.DB) *UsedCaptchaStore {
	return &UsedCaptchaStore{
		db:       db,
		tokens:   make(map[string]int64),
		attempts: make(map[string]int),
	}
}

var UsedCaptchas = NewUsedCaptchaStore(nil)

func (s *UsedCaptchaStore) init(db *sql.DB) {
	s.initOnce.Do(func() {
		s.mu.Lock()
		s.db = db
		s.mu.Unlock()
	})
}

func (s *UsedCaptchaStore) getDB() *sql.DB {
	s.mu.Lock()
	db := s.db
	s.mu.Unlock()
	if db != nil {
		return db
	}
	if s != UsedCaptchas {
		panic("FATAL: UsedCaptchaStore called before database initialization")
	}
	if DB != nil {
		s.init(DB)
		return DB
	}
	if os.Getenv("HERASAT_DEV_MODE") == "1" || flag.Lookup("test.v") != nil {
		return nil
	}
	panic("FATAL: UsedCaptchaStore called before database initialization")
}

func (s *UsedCaptchaStore) RecordFailure(signature string, expiresAt int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[signature]++
	if s.attempts[signature] >= 2 {
		delete(s.attempts, signature)
		s.tokens[signature] = expiresAt
		if db := s.db; db != nil {
			_, _ = db.Exec(`
				INSERT INTO used_captcha_tokens (signature_b64, expires_at)
				VALUES ($1, TO_TIMESTAMP($2))
				ON CONFLICT (signature_b64) DO NOTHING`, signature, expiresAt)
		}
		return true
	}
	return false
}

func (s *UsedCaptchaStore) TryRedeem(signature string, expiresAt int64) bool {
	now := time.Now().Unix()
	if now > expiresAt {
		return false
	}

	s.mu.Lock()
	delete(s.attempts, signature)
	if exp, exists := s.tokens[signature]; exists {
		s.mu.Unlock()
		if now <= exp {
			return false
		}
	}
	s.mu.Unlock()

	if db := s.getDB(); db != nil {
		var returned int
		err := db.QueryRow(`
			INSERT INTO used_captcha_tokens (signature_b64, expires_at)
			VALUES ($1, TO_TIMESTAMP($2))
			ON CONFLICT (signature_b64) DO NOTHING
			RETURNING 1`, signature, expiresAt).Scan(&returned)
		if err != nil || returned != 1 {
			s.mu.Lock()
			s.tokens[signature] = expiresAt
			s.mu.Unlock()
			return false
		}

		s.mu.Lock()
		s.tokens[signature] = expiresAt
		s.mu.Unlock()
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tokens[signature]; exists {
		return false
	}

	s.tokens[signature] = expiresAt
	return true
}

func (s *UsedCaptchaStore) StartReaper(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			now := time.Now()

			if db := s.getDB(); db != nil {
				_, _ = db.Exec(`DELETE FROM used_captcha_tokens WHERE expires_at < NOW()`)
			}

			s.mu.Lock()
			for sig, exp := range s.tokens {
				if now.Unix() > exp {
					delete(s.tokens, sig)
				}
			}
			s.mu.Unlock()
		}
	}()
}

func GenerateCaptchaToken(answer int, ttl time.Duration) string {
	payload := CaptchaPayload{
		Answer:    answer,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	payloadBytes, _ := json.Marshal(payload)
	encPayload, err := EncryptCaptchaPayload(string(payloadBytes))
	if err != nil {
		return ""
	}

	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(encPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", encPayload, signature)
}

func VerifyCaptcha(token, userAnswerStr string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	encPayload, sigB64 := parts[0], parts[1]

	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(encPayload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigB64), []byte(expectedSig)) {
		macLegacy := hmac.New(sha256.New, GetLegacyHMACKey())
		macLegacy.Write([]byte(encPayload))
		legacyExpectedSig := base64.RawURLEncoding.EncodeToString(macLegacy.Sum(nil))
		if !hmac.Equal([]byte(sigB64), []byte(legacyExpectedSig)) {
			return false
		}
	}

	decrypted, err := DecryptCaptchaPayload(encPayload)
	if err != nil {
		return false
	}

	var payload CaptchaPayload
	if err := json.Unmarshal([]byte(decrypted), &payload); err != nil {
		return false
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return false
	}

	userAnswer, err := strconv.Atoi(strings.TrimSpace(userAnswerStr))
	if err != nil {
		UsedCaptchas.RecordFailure(sigB64, payload.ExpiresAt)
		return false
	}

	if userAnswer != payload.Answer {
		UsedCaptchas.RecordFailure(sigB64, payload.ExpiresAt)
		return false
	}

	return UsedCaptchas.TryRedeem(sigB64, payload.ExpiresAt)
}

func cryptoRandInt(min, max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return int(nBig.Int64()) + min
}

func generateNoiseElements() string {
	var noise strings.Builder
	colors := []string{"#38bdf8", "#818cf8", "#f43f5e", "#10b981", "#fbbf24", "#64748b"}

	for i := 0; i < 28; i++ {
		cx := cryptoRandInt(5, 175)
		cy := cryptoRandInt(5, 55)
		r := cryptoRandInt(1, 2)
		color := colors[cryptoRandInt(0, len(colors)-1)]
		noise.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" opacity="0.30"/>`, cx, cy, r, color))
	}

	for i := 0; i < 3; i++ {
		x1 := cryptoRandInt(5, 30)
		y1 := cryptoRandInt(10, 50)
		cx := cryptoRandInt(50, 130)
		cy := cryptoRandInt(5, 55)
		x2 := cryptoRandInt(150, 175)
		y2 := cryptoRandInt(10, 50)
		color := colors[cryptoRandInt(0, len(colors)-1)]
		noise.WriteString(fmt.Sprintf(`<path d="M%d,%d Q%d,%d %d,%d" stroke="%s" stroke-width="1.2" fill="none" opacity="0.35"/>`,
			x1, y1, cx, cy, x2, y2, color))
	}

	return noise.String()
}

var glyphPaths = map[rune]string{
	'0': "M 3,3 C 3,1 9,1 9,3 L 9,13 C 9,15 3,15 3,13 Z",
	'1': "M 4,4 L 7,2 L 7,15 M 4,15 L 10,15",
	'2': "M 3,5 C 3,2 9,2 9,5 C 9,9 3,11 3,15 L 10,15",
	'3': "M 3,3 L 9,3 L 6,7 C 9,7 9,14 5,14 C 3,14 3,12 3,12",
	'4': "M 8,15 L 8,2 L 2,10 L 10,10",
	'5': "M 9,2 L 3,2 L 3,7 C 5,6 9,6 9,10 C 9,14 3,14 3,12",
	'6': "M 8,2 C 4,3 3,7 3,10 C 3,14 9,14 9,10 C 9,7 3,8 3,10",
	'7': "M 2,2 L 9,2 L 5,15",
	'8': "M 6,2 C 9,2 9,7 6,7 C 3,7 3,2 6,2 M 6,7 C 9,7 9,14 6,14 C 3,14 3,7 6,7",
	'9': "M 4,14 C 8,13 9,9 9,6 C 9,2 3,2 3,6 C 3,9 9,8 9,6",
	'+': "M 2,8 L 10,8 M 6,4 L 6,12",
	'-': "M 2,8 L 10,8",
	'=': "M 2,6 L 10,6 M 2,10 L 10,10",
	'?': "M 3,4 C 3,1 9,1 9,5 C 9,8 6,8 6,11 M 6,14.5 L 6,15",
}

func renderObfuscatedChars(expr string) string {
	var sb strings.Builder
	charColors := []string{"#38bdf8", "#60a5fa", "#93c5fd", "#e0f2fe", "#7dd3fc"}

	step := 16
	spaceWidth := 8
	totalWidth := 0
	for _, ch := range expr {
		if ch == ' ' {
			totalWidth += spaceWidth
		} else {
			totalWidth += step
		}
	}

	currentX := (180 - totalWidth) / 2

	for _, ch := range expr {
		if ch == ' ' {
			currentX += spaceWidth
			continue
		}

		pathData, exists := glyphPaths[ch]
		if !exists {
			currentX += step
			continue
		}

		rot := cryptoRandInt(-6, 6)
		yJitter := cryptoRandInt(18, 21)
		color := charColors[cryptoRandInt(0, len(charColors)-1)]

		sb.WriteString(fmt.Sprintf(
			`<g transform="translate(%d,%d) rotate(%d 6 8) scale(1.38)"><path d="%s" stroke="%s" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" fill="none"/></g>`,
			currentX, yJitter, rot, pathData, color,
		))

		currentX += step
	}

	return sb.String()
}

func GenerateMathCaptchaSVG() (string, string) {
	num1 := cryptoRandInt(25, 95)
	num2 := cryptoRandInt(10, 45)
	isAddition := cryptoRandInt(0, 1) == 1

	var text string
	var answer int

	if isAddition {
		answer = num1 + num2
		text = fmt.Sprintf("%d + %d = ?", num1, num2)
	} else {
		if num1 < num2 {
			num1, num2 = num2, num1
		}
		answer = num1 - num2
		text = fmt.Sprintf("%d - %d = ?", num1, num2)
	}

	token := GenerateCaptchaToken(answer, 2*time.Minute)
	noise := generateNoiseElements()
	renderedChars := renderObfuscatedChars(text)
	seed := cryptoRandInt(1, 999)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="180" height="60" viewBox="0 0 180 60">
  <defs>
    <filter id="distortionFilter" x="-10%%" y="-10%%" width="120%%" height="120%%">
      <feTurbulence type="turbulence" baseFrequency="0.03 0.03" numOctaves="2" seed="%d" result="noise" />
      <feDisplacementMap in="SourceGraphic" in2="noise" scale="2" xChannelSelector="R" yChannelSelector="G" />
    </filter>
  </defs>
  <rect width="100%%" height="100%%" fill="#0b1329" rx="8" stroke="#1e293b" stroke-width="1.5"/>
  <g filter="url(#distortionFilter)">
    %s
    %s
  </g>
</svg>`, seed, noise, renderedChars)

	base64URI := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	return token, base64URI
}

func GetCaptchaHandler(w http.ResponseWriter, r *http.Request) {
	token, svgDataURI := GenerateMathCaptchaSVG()

	resp := CaptchaResponse{
		CaptchaID:   token,
		ImageBase64: svgDataURI,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	json.NewEncoder(w).Encode(resp)
}
