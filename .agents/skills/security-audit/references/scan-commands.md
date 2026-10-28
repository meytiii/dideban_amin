# Scan Commands by Platform

## Frontend 🖥️ (JavaScript/TypeScript)

```bash
# Hardcoded secrets
rg -i "(apikey|api_key|secret|password|token)" src/ -g '!*.test.*'

# Dangerous APIs
rg "eval\(|Function\(|dangerouslySetInnerHTML|innerHTML|document\.write" src/

# Token storage
rg "localStorage\.(setItem|getItem)" src/

# Console output (production leak)
rg "console\.(log|warn|error)" src/ --count

# External links missing rel
rg 'target=.*_blank' src/ --type tsx --type jsx

# Environment variable usage
rg "import\.meta\.env\.|process\.env\." src/

# Open redirect risk
rg "window\.location\s*=|window\.location\.href\s*=" src/

# Fetch/axios without error handling
rg "fetch\(|axios\.(get|post|put|delete)" src/ --count
```

## Backend 🖥️ (Python)

```bash
# Hardcoded secrets
rg -i "(password|secret|api_key|token|client_secret)\s*=\s*[\"']" app/ -g '!*test*'

# Dangerous APIs
rg "eval\(|exec\(|pickle\.loads|yaml\.load\b|os\.system\(|__import__" app/

# Subprocess
rg "subprocess\.(run|call|Popen|check_output)" app/

# SQL injection
rg "\.execute\(.*f\"|\.execute\(.*\.format\(|\.execute\(.*%\s" app/
rg "text\(f\"|text\(.*\.format" app/

# JWT configuration
rg "jwt\.(decode|encode)|JWT_SECRET|verify_exp" app/

# Sensitive data in logs
rg "logger\.\w+\(.*token|logger\.\w+\(.*password|logger\.\w+\(.*secret" app/ -i

# Broad exception handlers
rg "except\s*(Exception|:)" app/ --count

# CORS configuration
rg "CORSMiddleware|allow_origins|cors" app/ -i

# Database connections (SSL check)
rg "create.*engine|DATABASE_URL|postgresql" app/ config/

# Redis connections (SSL check)
rg "redis://|rediss://|REDIS" app/ config/

# Insecure HTTP requests
rg "verify\s*=\s*False" app/

# File operations
rg "open\(.*\+|pathlib|os\.path\.join" app/ | rg -v test
```

## Backend (Node.js)

```bash
# Hardcoded secrets
rg -i "(apikey|api_key|secret|password|token)\s*[:=]\s*[\"']" src/ -g '!*.test.*'

# Dangerous APIs
rg "eval\(|Function\(|child_process|require\([\"']child_process" src/

# SQL injection (if raw queries)
rg "\.query\(.*\`|\.query\(.*\+|\.raw\(" src/

# Prototype pollution
rg "Object\.assign|\.\.\.req\.(body|query|params)" src/

# Insecure deserialization
rg "JSON\.parse\(.*req\.|unserialize|deserialize" src/
```

## Desktop 🖳 (Electron)

```bash
# Electron security settings
rg "nodeIntegration|contextIsolation|sandbox|webSecurity|enableRemoteModule" src/main/

# IPC channels (Main process)
rg "ipcMain\.(handle|on)\(" src/main/

# IPC exposure (Preload)
rg "contextBridge|exposeInMainWorld|ipcRenderer\.(send|invoke|on)" src/preload/

# Dangerous injection
rg "executeJavaScript|loadURL\(.*javascript:|eval\(" src/

# Shell access
rg "shell\.openExternal|shell\.openPath" src/

# Local storage
rg "electron-store|safeStorage|store\.(get|set)\(" src/

# Protocol / Deep link
rg "setAsDefaultProtocolClient|protocol\.(register|handle)|registerFileProtocol" src/main/

# Auto-updater
rg "autoUpdater|electron-updater|checkForUpdates" src/

# Dev mode bypasses
rg "ELECTRON_DISABLE_SECURITY_WARNINGS|disable-web-security|ignore-certificate-errors" src/

# Permissions
rg "setPermissionRequestHandler|setPermissionCheckHandler" src/main/

# WebView security
rg "will-attach-webview|webview|BrowserView" src/main/

# Navigation security
rg "will-navigate|setWindowOpenHandler|new-window" src/main/

# Hardcoded secrets
rg -i "(apikey|api_key|secret|password|token)\s*[:=]\s*[\"']" src/ -g '!*.test.*'
```

## AI/LLM 🤖

```bash
# LLM API keys
rg -i "(openai|anthropic|claude|gpt|api.key|api_key|bearer)" src/ -g '!*.test.*'

# Prompt construction (injection risk)
rg "system.*prompt|user.*prompt|\.chat\.completions|messages.*role" src/

# Unsafe output rendering
rg "dangerouslySetInnerHTML|innerHTML|v-html" src/ # check if LLM output flows here

# LLM output used in eval/exec
rg "eval\(.*response|exec\(.*response|Function\(.*response" src/

# Model selection by user input
rg "model.*req\.(body|query|params)|model.*user" src/

# PII in prompts
rg "email|phone|address|ssn|social.security" src/ # cross-reference with LLM API calls

# Cost control
rg "max_tokens|temperature|top_p" src/ # verify limits are set
```

## Container 🐳 (Docker / Kubernetes)

```bash
# --- Dockerfile Analysis ---

# Root user (missing USER directive)
rg "^USER" Dockerfile*  # if no result, container runs as root

# Secrets in Dockerfile
rg -i "(password|secret|token|api.key|credentials)\s*=" Dockerfile* docker-compose*

# Latest tag usage
rg ":latest" Dockerfile* docker-compose* k8s/

# ADD instead of COPY (can auto-extract remote URLs)
rg "^ADD " Dockerfile*

# Missing .dockerignore check
ls .dockerignore 2>/dev/null || echo "WARNING: No .dockerignore found"

# --- Docker Compose Analysis ---

# Privileged mode
rg -i "privileged|SYS_ADMIN|cap_add" docker-compose*

# Docker socket mount (full host control)
rg "docker\.sock" docker-compose* k8s/

# Host network mode
rg "network_mode.*host|hostNetwork.*true" docker-compose* k8s/

# Host path mounts (potential filesystem escape)
rg "host.*path|hostPath|type:\s*bind" docker-compose* k8s/

# Missing resource limits
rg -L "mem_limit|memory|resources" docker-compose*  # files WITHOUT limits

# Insecure registries
rg "http://" Dockerfile* docker-compose* k8s/ | rg -v "localhost|127\.0\.0"

# --- Kubernetes Manifests ---

# Running as root
rg "runAsNonRoot.*false|runAsUser.*0" k8s/

# Privilege escalation allowed
rg "allowPrivilegeEscalation.*true" k8s/

# Missing security context
rg -L "securityContext" k8s/  # manifests WITHOUT securityContext

# Service account token auto-mount
rg "automountServiceAccountToken" k8s/

# Cluster-admin binding (overprivileged)
rg "cluster-admin" k8s/

# --- Image Scanning (run separately) ---

# Trivy scan
# trivy image <image>:<tag>
# trivy image --severity CRITICAL,HIGH <image>:<tag>

# Hadolint (Dockerfile lint)
# hadolint Dockerfile

# Dockle (image lint)
# dockle <image>:<tag>

# Docker history (check for secrets in layers)
# docker history --no-trunc <image>:<tag> | grep -i "secret\|password\|token\|key"
```

## Universal 🌐

```bash
# Git history secrets check
git log --diff-filter=A --name-only | head -100  # check for .env files in history

# Large files that might contain secrets
rg -l "BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY" .

# TODO/FIXME security notes
rg -i "TODO.*secur|FIXME.*secur|HACK.*secur|XXX.*secur" src/

# HTTP (non-HTTPS) URLs in source
rg "http://" src/ -g '!*.test.*' -g '!*.md' | rg -v localhost
```
