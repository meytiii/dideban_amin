# Electron Security Guide

> Read this reference when auditing Electron desktop applications.

---

## Critical Security Settings

Every `BrowserWindow` must use these webPreferences:

```typescript
// ✅ Secure defaults
const securePreferences = {
  nodeIntegration: false,
  contextIsolation: true,
  sandbox: true,
  webSecurity: true,
  allowRunningInsecureContent: false,
  enableRemoteModule: false,
  experimentalFeatures: false,
}
```

| Setting | Secure Value | Risk if Wrong |
|---------|-------------|---------------|
| `nodeIntegration` | `false` | XSS → Full RCE (access to `require`, `fs`, `child_process`) |
| `contextIsolation` | `true` | Renderer can access preload's Node.js context |
| `sandbox` | `true` | Renderer has unrestricted OS access |
| `webSecurity` | `true` | Same-origin policy disabled, CORS bypassed |
| `enableRemoteModule` | `false` | Renderer can directly call main process APIs |

---

## IPC Security Architecture

### Secure Pattern (Whitelist)

```typescript
// preload.ts — expose specific functions only
contextBridge.exposeInMainWorld('api', {
  getVersion: () => ipcRenderer.invoke('app:get-version'),
  startDownload: (url: string) => ipcRenderer.invoke('update:download', url),
  openExternal: (url: string) => {
    if (url.startsWith('https://')) {
      return ipcRenderer.invoke('shell:open-external', url)
    }
    throw new Error('Invalid URL scheme')
  },
})
```

### Insecure Pattern (Avoid)

```typescript
// ❌ NEVER expose raw ipcRenderer
contextBridge.exposeInMainWorld('electron', {
  send: ipcRenderer.send,        // any channel callable
  invoke: ipcRenderer.invoke,     // any handler callable
})
```

### IPC Handler Input Validation

```typescript
// main.ts — always validate
ipcMain.handle('store:get', (_event, key: string) => {
  const ALLOWED_KEYS = ['theme', 'language', 'windowBounds']
  if (!ALLOWED_KEYS.includes(key)) {
    throw new Error(`Access denied for key: ${key}`)
  }
  return store.get(key)
})
```

---

## Protocol Handler Security

### Deep Link Token Validation

```typescript
// ❌ No validation
const token = url.match(/accessToken=([^&]+)/)?.[1]
store.set('accessToken', token)

// ✅ With validation
const token = url.match(/accessToken=([^&]+)/)?.[1]
if (token && /^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$/.test(token)) {
  store.set('accessToken', token)
} else {
  log.warn('Invalid token format in deep link')
}
```

### URL Scheme Restrictions

```typescript
// Validate shell.openExternal
function safeOpenExternal(url: string) {
  const parsed = new URL(url)
  if (!['https:', 'http:', 'mailto:'].includes(parsed.protocol)) {
    throw new Error(`Blocked URL scheme: ${parsed.protocol}`)
  }
  shell.openExternal(url)
}
```

---

## Local Data Security

| Method | Security | Use For |
|--------|----------|---------|
| `electron-store` (JSON) | ❌ Plaintext | Non-sensitive preferences only |
| `safeStorage.encryptString()` | ✅ OS-level encryption | Tokens, credentials |
| macOS Keychain (via `keytar`) | ✅ System keychain | Long-lived secrets |
| SQLite + sqlcipher | ✅ Encrypted DB | Large structured sensitive data |

```typescript
// ✅ Encrypt tokens with safeStorage
import { safeStorage } from 'electron'

function storeToken(token: string) {
  const encrypted = safeStorage.encryptString(token)
  store.set('encryptedToken', encrypted.toString('base64'))
}

function getToken(): string {
  const encrypted = Buffer.from(store.get('encryptedToken'), 'base64')
  return safeStorage.decryptString(encrypted)
}
```

---

## Auto-Update Security

- [ ] Update server uses HTTPS only
- [ ] Update packages are code-signed
- [ ] Signature verification enabled in electron-updater
- [ ] `autoDownload: false` (user confirms before download)
- [ ] Rollback mechanism on update failure
- [ ] Update channel separation (stable vs beta)

---

## Code Signing & Notarization

### macOS

```bash
# Verify app is properly signed
codesign --verify --deep --strict /path/to/App.app

# Check notarization
spctl --assess --type execute /path/to/App.app

# List entitlements
codesign -d --entitlements :- /path/to/App.app
```

### Entitlements Audit

| Entitlement | Risk | When Needed |
|------------|------|-------------|
| `com.apple.security.cs.allow-jit` | Medium | JIT compilation needed |
| `com.apple.security.cs.allow-unsigned-executable-memory` | High | Only if absolutely required |
| `com.apple.security.cs.allow-dyld-environment-variables` | High | Debugging only |
| `com.apple.security.cs.disable-library-validation` | High | Third-party dylibs |
| `com.apple.security.device.camera` | Low | Camera features |
| `com.apple.security.device.audio-input` | Low | Microphone features |

---

## Navigation Security

```typescript
// Block external navigation in BrowserWindow
window.webContents.on('will-navigate', (event, url) => {
  const allowed = ['https://your-app.com', 'file://']
  if (!allowed.some(origin => url.startsWith(origin))) {
    event.preventDefault()
  }
})

// Handle new window requests
window.webContents.setWindowOpenHandler(({ url }) => {
  if (url.startsWith('https://')) {
    shell.openExternal(url)
  }
  return { action: 'deny' }
})

// Enforce WebView security
window.webContents.on('will-attach-webview', (_event, webPreferences) => {
  webPreferences.nodeIntegration = false
  webPreferences.contextIsolation = true
  webPreferences.sandbox = true
  webPreferences.webSecurity = true
})
```

---

## Development Mode Security

```typescript
const isDev = !app.isPackaged // reliable check

if (isDev) {
  // Only these bypasses in dev
  process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = 'true'
  // NEVER: app.commandLine.appendSwitch('disable-web-security')
  // NEVER: app.commandLine.appendSwitch('ignore-certificate-errors')
}
```

**Key rule**: `app.isPackaged` is the only reliable way to check production mode. Never use `process.env.NODE_ENV` alone (can be overridden).

---

## References

- [Electron Security Tutorial](https://www.electronjs.org/docs/latest/tutorial/security)
- [Electron Security Checklist](https://www.electronjs.org/docs/latest/tutorial/security#checklist-security-recommendations)
- [OWASP Electron Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Electron_Security_Cheat_Sheet.html)
