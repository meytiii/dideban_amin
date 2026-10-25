# Technical Evidence: CAPTCHA Token Answer Leakage Analysis

## Overview
- **Vulnerability ID**: DA-SEC-001
- **Severity**: High (CVSS 4.0: 8.7)
- **Target File**: `captcha.go`
- **Affected Functions**: `GenerateCaptchaToken` (line 134), `VerifyCaptcha` (line 150)
- **CWE**: CWE-312 (Cleartext Transmission of Sensitive Information), CWE-807 (Reliance on Untrusted Inputs)

## Execution Path Trace

In `captcha.go`:
```go
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
```

1. The server generates mathematical arithmetic (e.g. `12 + 7 = 19`).
2. The integer answer `19` is placed into `CaptchaPayload.Answer`.
3. `payloadBytes` is `{"a":19,"e":1729845000}`.
4. `payloadB64` is the standard Base64 representation of `{"a":19,"e":1729845000}` -> `eyJhIjoxOSwiZSI6MTcyOTg0NTAwMH0`.
5. The signature is computed over `payloadB64`.
6. The client receives `{"captcha_id": "eyJhIjoxOSwiZSI6MTcyOTg0NTAwMH0.<signature>", "image_base64": "..."}`.

## Client-Side Extraction Proof

Any automated script or attacker can solve the CAPTCHA with 100% accuracy in zero milliseconds without computer vision or OCR:

```javascript
function solveCaptchaWithoutLooking(captchaId) {
    const rawPayload = atob(captchaId.split('.')[0]);
    const payload = JSON.parse(rawPayload);
    return payload.a;
}
```

Because `VerifyCaptcha` takes `captcha_id` and `userAnswerStr`, validates the HMAC signature on `payloadB64`, extracts `payload.Answer`, and matches `userAnswer == payload.Answer`, the submitted answer matches perfectly and the token is redeemed.

## Root Cause
HMAC (Hash-based Message Authentication Code) provides **integrity and authenticity**, not **confidentiality**. Signing sensitive data without encrypting it leaves it exposed to anyone holding the token.

## Remediation Strategy
1. **Option A (Encrypted Token)**: Encrypt `CaptchaPayload` using `EncryptAES(string(payloadBytes))` before sending to client, or
2. **Option B (Server-Side Storage)**: Store `captcha_id -> answer` in PostgreSQL or an in-memory TTL map with single-use eviction, sending only an opaque random ID to the client.

## Remediation Resolution & Verification (DA-SEC-001)
- **Status**: **REMEDIATED & VERIFIED**
- **Implementation**: Adopted Option A with dedicated domain key separation: `keyCaptcha = deriveKey(aesKey, "HERASAT_CAPTCHA_ENC_V2")`.
- `crypto.go:183-225`: Implemented `EncryptCaptchaPayload` and `DecryptCaptchaPayload` using authenticated AES-256-GCM encryption with cryptographically random nonces.
- `captcha.go:134-195`: `GenerateCaptchaToken` encrypts the payload prior to calculating the HMAC signature. `VerifyCaptcha` validates the HMAC signature, decrypts the payload, checks expiration, and redeems the signature in `used_captcha_tokens`.
- **Automated Verification**: `TestRemediationCaptchaConfidentialityAndIntegrity` passed in 0.13s, confirming that:
  - Base64 payload decoding does not yield JSON.
  - Plaintext answers cannot be recovered without the server's master key.
  - Replay attempts are rejected.
  - Bit-flipped ciphertexts fail cryptographic authentication.
