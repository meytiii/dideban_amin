package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
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
}

func NewUsedCaptchaStore(db *sql.DB) *UsedCaptchaStore {
	return &UsedCaptchaStore{
		db:     db,
		tokens: make(map[string]int64),
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
	if s == UsedCaptchas && DB != nil {
		s.init(DB)
		return DB
	}
	panic("FATAL: UsedCaptchaStore called before database initialization")
}

func (s *UsedCaptchaStore) TryRedeem(signature string, expiresAt int64) bool {
	now := time.Now().Unix()
	if now > expiresAt {
		return false
	}

	s.mu.Lock()
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
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(payloadB64))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", payloadB64, signature)
}

func VerifyCaptcha(token, userAnswerStr string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	payloadB64, sigB64 := parts[0], parts[1]

	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(payloadB64))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigB64), []byte(expectedSig)) {
		macLegacy := hmac.New(sha256.New, GetLegacyHMACKey())
		macLegacy.Write([]byte(payloadB64))
		legacyExpectedSig := base64.RawURLEncoding.EncodeToString(macLegacy.Sum(nil))
		if !hmac.Equal([]byte(sigB64), []byte(legacyExpectedSig)) {
			return false
		}
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return false
	}

	var payload CaptchaPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return false
	}

	userAnswer, err := strconv.Atoi(strings.TrimSpace(userAnswerStr))
	if err != nil {
		return false
	}

	if userAnswer != payload.Answer {
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

	for i := 0; i < 45; i++ {
		cx := cryptoRandInt(5, 175)
		cy := cryptoRandInt(5, 55)
		r := cryptoRandInt(1, 2)
		color := colors[cryptoRandInt(0, len(colors)-1)]
		noise.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" opacity="0.45"/>`, cx, cy, r, color))
	}

	for i := 0; i < 6; i++ {
		x1 := cryptoRandInt(5, 30)
		y1 := cryptoRandInt(5, 55)
		cx := cryptoRandInt(50, 130)
		cy := cryptoRandInt(0, 60)
		x2 := cryptoRandInt(150, 175)
		y2 := cryptoRandInt(5, 55)
		strokeWidth := cryptoRandInt(1, 2)
		color := colors[cryptoRandInt(0, len(colors)-1)]
		noise.WriteString(fmt.Sprintf(`<path d="M%d,%d Q%d,%d %d,%d" stroke="%s" stroke-width="%d" fill="none" opacity="0.65"/>`,
			x1, y1, cx, cy, x2, y2, color, strokeWidth))
	}

	return noise.String()
}

func renderObfuscatedChars(expr string) string {
	var sb strings.Builder
	charColors := []string{"#38bdf8", "#60a5fa", "#93c5fd", "#e0f2fe", "#7dd3fc"}

	currentX := cryptoRandInt(14, 20)
	step := 17

	for _, ch := range expr {
		if ch == ' ' {
			currentX += 8
			continue
		}

		rot := cryptoRandInt(-20, 20)
		yJitter := cryptoRandInt(34, 42)
		fontSize := cryptoRandInt(22, 27)
		color := charColors[cryptoRandInt(0, len(charColors)-1)]

		sb.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d" font-family="'Courier New', Courier, monospace, sans-serif" font-size="%d" font-weight="900" fill="%s" transform="rotate(%d %d %d)">%c</text>`,
			currentX, yJitter, fontSize, color, rot, currentX, yJitter, ch,
		))

		currentX += step + cryptoRandInt(-2, 3)
	}

	return sb.String()
}

func GenerateMathCaptchaSVG() (string, string) {
	num1 := cryptoRandInt(12, 49)
	num2 := cryptoRandInt(1, 9)
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

	baseFreqX := fmt.Sprintf("0.0%d", cryptoRandInt(3, 7))
	baseFreqY := fmt.Sprintf("0.0%d", cryptoRandInt(4, 8))
	scale := cryptoRandInt(3, 5)
	seed := cryptoRandInt(1, 999)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="180" height="60" viewBox="0 0 180 60">
  <defs>
    <filter id="distortionFilter" x="-10%%" y="-10%%" width="120%%" height="120%%">
      <feTurbulence type="turbulence" baseFrequency="%s %s" numOctaves="2" seed="%d" result="noise" />
      <feDisplacementMap in="SourceGraphic" in2="noise" scale="%d" xChannelSelector="R" yChannelSelector="G" />
    </filter>
  </defs>
  <rect width="100%%" height="100%%" fill="#0b1329" rx="8" stroke="#1e293b" stroke-width="1.5"/>
  <g filter="url(#distortionFilter)">
    %s
    %s
  </g>
</svg>`, baseFreqX, baseFreqY, seed, scale, noise, renderedChars)

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
