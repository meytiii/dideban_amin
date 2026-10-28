# Container Security Guide

> Read this reference when auditing projects that use Docker, Kubernetes, or other container technologies.

---

## Table of Contents

1. [Dockerfile Security](#dockerfile-security)
2. [Image Supply Chain](#image-supply-chain)
3. [Runtime Security](#runtime-security)
4. [Secrets Management](#secrets-management)
5. [Network & Orchestration](#network--orchestration)
6. [Kubernetes-Specific](#kubernetes-specific)
7. [CI/CD Pipeline Security](#cicd-pipeline-security)
8. [Monitoring & Incident Response](#monitoring--incident-response)
9. [Scan Commands & Tools](#scan-commands--tools)

---

## Dockerfile Security

### Base Image Selection

| Image Type | Security | Size | When to Use |
|-----------|----------|------|-------------|
| `distroless` | ✅ Best (no shell, no pkg mgr) | Smallest | Production services |
| `*-alpine` | ✅ Good (minimal attack surface) | Small | General use |
| `*-slim` | ⚠️ Moderate (reduced packages) | Medium | When alpine has compat issues |
| `*:latest` | ❌ Unpinned, unpredictable | Varies | Never in production |
| Full OS images | ❌ Large attack surface | Large | Only for build stages |

### Secure Dockerfile Template

```dockerfile
# ✅ Secure multi-stage build example
# --- Build stage ---
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY src/ ./src/
RUN pnpm build

# --- Production stage ---
FROM node:20-alpine AS production

# Create non-root user
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -D appuser

WORKDIR /app

# Copy only production artifacts
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY package.json ./

# Switch to non-root
USER appuser

# Declare port (documentation, not security)
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["node", "dist/index.js"]
```

### Common Dockerfile Vulnerabilities

```dockerfile
# ❌ BAD: Running as root (default)
FROM node:20
COPY . .
CMD ["node", "index.js"]

# ❌ BAD: Secrets in build args
ARG DB_PASSWORD
ENV DB_PASSWORD=$DB_PASSWORD

# ❌ BAD: Copying everything (includes .env, .git, secrets)
COPY . .

# ❌ BAD: Using latest tag
FROM python:latest

# ❌ BAD: Installing unnecessary tools in production
RUN apt-get install -y curl vim net-tools ssh

# ❌ BAD: Running package install without lockfile
RUN npm install  # should be: npm ci --only=production
```

### .dockerignore Essentials

```
# Must ignore
.env
.env.*
.git
.gitignore
node_modules
__pycache__
*.pyc
.venv
.teller.yml
*.pem
*.key
*.p12
docker-compose*.yml
Dockerfile*
.dockerignore
*.md
tests/
coverage/
.cursor/
.vscode/
```

---

## Image Supply Chain

### Threat: Malicious or Vulnerable Base Images

```
Registry ──pull──▶ Base Image ──build──▶ App Image ──deploy──▶ Container
   ↑                    ↑                     ↑                   ↑
  Typosquat        Known CVEs          Injected layers       Runtime exploit
```

### Defense Checklist

- [ ] Base images from official/verified publishers only
- [ ] Image tags pinned to SHA256 digest (most secure) or specific version
- [ ] Image scanning in CI before push to registry
- [ ] Private registry with vulnerability scanning (ECR, GCR, ACR)
- [ ] Image signing with `cosign` (Sigstore) or Docker Content Trust
- [ ] Regular rebuild schedule (weekly) to pick up base image patches
- [ ] SBOM (Software Bill of Materials) generated per image

### Pinning Examples

```dockerfile
# ✅ Good: pinned to specific version
FROM python:3.10-slim

# ✅ Best: pinned to SHA256 digest
FROM python:3.10-slim@sha256:abc123def456...

# ❌ Bad: unpinned
FROM python:latest
FROM python
```

---

## Runtime Security

### Docker Run Hardening

```bash
# ✅ Secure runtime invocation
docker run \
  --read-only \                    # read-only root filesystem
  --cap-drop=ALL \                 # drop all capabilities
  --cap-add=NET_BIND_SERVICE \     # add back only what's needed
  --security-opt=no-new-privileges \ # prevent privilege escalation
  --memory=512m \                  # memory limit
  --cpus=1.0 \                     # CPU limit
  --pids-limit=100 \               # prevent fork bomb
  --tmpfs /tmp:rw,noexec,nosuid \  # writable tmp without exec
  --user 1001:1001 \               # non-root user
  --network=app-network \          # isolated network
  my-app:v1.0.0
```

### Dangerous Flags (Audit for These)

| Flag | Risk | Severity |
|------|------|----------|
| `--privileged` | Full host access, container escape trivial | CRITICAL |
| `--cap-add=SYS_ADMIN` | Near-privileged, mount namespace escape | CRITICAL |
| `--network=host` | Container shares host network stack | HIGH |
| `--pid=host` | Can see/signal host processes | HIGH |
| `--ipc=host` | Shared memory with host | HIGH |
| `-v /:/host` | Mounts entire host filesystem | CRITICAL |
| `-v /var/run/docker.sock:/var/run/docker.sock` | Docker-in-Docker, full host control | CRITICAL |
| `--security-opt apparmor=unconfined` | Disables AppArmor | HIGH |
| `--security-opt seccomp=unconfined` | Disables seccomp filtering | HIGH |

### Docker Socket Exposure

Mounting `/var/run/docker.sock` inside a container is equivalent to giving root access to the host.

```yaml
# ❌ CRITICAL: Never expose Docker socket unless absolutely necessary
volumes:
  - /var/run/docker.sock:/var/run/docker.sock

# If required (e.g., CI runner), mitigate:
# - Use socket proxy (tecnativa/docker-socket-proxy) with read-only API subset
# - Limit to specific containers
# - Monitor all Docker API calls
```

---

## Secrets Management

### Bad vs Good Patterns

```dockerfile
# ❌ BAD: Secret in ENV (visible in image inspect, docker history)
ENV DATABASE_URL=postgresql://user:password@db:5432/app

# ❌ BAD: Secret in build arg (visible in image history)
ARG SECRET_KEY
RUN echo $SECRET_KEY > /app/config

# ❌ BAD: Secret COPYed into image layer
COPY credentials.json /app/

# ✅ GOOD: Secret via runtime mount (never in image)
# docker run -v /secrets/db-pass:/run/secrets/db-pass:ro my-app

# ✅ GOOD: Docker build secrets (BuildKit, never in image layer)
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc npm ci
```

### Orchestrator Secrets

| Platform | Secret Mechanism | Encryption at Rest | Rotation |
|----------|-----------------|-------------------|----------|
| Docker Swarm | `docker secret` | Yes (Raft log) | Manual |
| Kubernetes | K8s Secrets | Base64 only (encrypt with KMS) | External tools |
| AWS ECS | Secrets Manager / SSM | Yes (KMS) | Automatic |
| Vault | Dynamic secrets | Yes | Automatic |

### Kubernetes Secrets Hardening

```yaml
# ❌ Default K8s secrets are only base64 encoded, NOT encrypted
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
data:
  password: cGFzc3dvcmQ=  # base64("password") — trivially decodable

# ✅ Enable encryption at rest (EncryptionConfiguration)
# ✅ Use external secrets operators (external-secrets, vault-injector)
# ✅ Use RBAC to restrict who can read secrets
# ✅ Avoid mounting secrets as env vars (prefer volume mounts)
```

---

## Network & Orchestration

### Docker Compose Network Isolation

```yaml
# ✅ Separate networks for frontend and backend tiers
services:
  frontend:
    networks: [frontend]
  api:
    networks: [frontend, backend]
  database:
    networks: [backend]  # not accessible from frontend

networks:
  frontend:
  backend:
    internal: true  # no external access
```

### Kubernetes Network Policies

```yaml
# ✅ Deny all ingress by default, then allow specific
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-ingress
spec:
  podSelector: {}
  policyTypes: [Ingress]
  # no ingress rules = deny all
```

---

## Kubernetes-Specific

### Pod Security Standards

| Level | Description | Use For |
|-------|-------------|---------|
| `Privileged` | No restrictions | System-level workloads only |
| `Baseline` | Prevents known privilege escalations | General workloads |
| `Restricted` | Heavily restricted, best practice | Security-sensitive workloads |

### Secure Pod Template

```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    fsGroup: 1001
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: app
    image: myapp:v1.0.0@sha256:abc123...
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop: [ALL]
    resources:
      limits:
        memory: "512Mi"
        cpu: "500m"
      requests:
        memory: "256Mi"
        cpu: "250m"
    livenessProbe:
      httpGet:
        path: /health
        port: 3000
```

### RBAC Audit Points

- [ ] Service accounts have minimal permissions
- [ ] No use of `cluster-admin` binding for workloads
- [ ] API server access restricted (not exposed publicly)
- [ ] `automountServiceAccountToken: false` where not needed
- [ ] Namespace isolation for different environments/teams

---

## CI/CD Pipeline Security

### Image Build Pipeline

```
Source Code → Lint Dockerfile → Build Image → Scan Image → Sign Image → Push to Registry
                  ↑                              ↑              ↑
               hadolint                   trivy/grype        cosign
```

### CI Integration Examples

```yaml
# GitHub Actions: Scan image before push
- name: Scan image
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: myapp:${{ github.sha }}
    severity: CRITICAL,HIGH
    exit-code: 1  # fail pipeline on findings

# Lint Dockerfile
- name: Lint Dockerfile
  uses: hadolint/hadolint-action@v3
  with:
    dockerfile: Dockerfile
```

### Pipeline Checklist

- [ ] Dockerfile linted (`hadolint`) in CI
- [ ] Image scanned for CVEs before push
- [ ] Critical/High CVEs fail the pipeline
- [ ] Image signed after successful scan
- [ ] Registry requires signed images for deployment
- [ ] Build context does not include secrets
- [ ] CI secrets are not echoed in logs

---

## Monitoring & Incident Response

### Runtime Monitoring Tools

| Tool | Focus | Deployment |
|------|-------|-----------|
| `Falco` | Syscall-based anomaly detection | DaemonSet / sidecar |
| `Sysdig` | Container visibility + security | Agent-based |
| `Aqua Security` | Full lifecycle container security | Commercial |
| `Twistlock/Prisma Cloud` | Runtime defense + compliance | Commercial |

### Alerts to Configure

- Container running as root
- Unexpected process spawned in container
- Network connection to unusual destination
- File modification in read-only filesystem attempt
- Capability escalation attempt
- Container escape indicators (mount namespace changes)

---

## Scan Commands & Tools

### Dockerfile Linting

```bash
# Lint Dockerfile for best practices
hadolint Dockerfile
hadolint --ignore DL3008 --ignore DL3013 Dockerfile  # ignore specific rules

# Alternative: dockle (image-level lint)
dockle myapp:latest
```

### Image Vulnerability Scanning

```bash
# Trivy — comprehensive, free, fast
trivy image myapp:latest
trivy image --severity CRITICAL,HIGH myapp:latest
trivy image --format json --output results.json myapp:latest

# Grype — Anchore's scanner
grype myapp:latest
grype myapp:latest --only-fixed  # show only fixable CVEs

# Snyk
snyk container test myapp:latest
```

### Kubernetes Security

```bash
# kube-bench — CIS benchmark compliance
kube-bench run

# kubesec — manifest security scoring
kubesec scan deployment.yaml

# Polaris — best practice checks
polaris audit --audit-path ./k8s/
```

### Runtime Inspection

```bash
# Check what user a container runs as
docker inspect --format='{{.Config.User}}' <container>

# Check capabilities
docker inspect --format='{{.HostConfig.CapAdd}}' <container>
docker inspect --format='{{.HostConfig.CapDrop}}' <container>

# Check if privileged
docker inspect --format='{{.HostConfig.Privileged}}' <container>

# Check mounts (look for docker.sock, host root)
docker inspect --format='{{json .Mounts}}' <container> | python3 -m json.tool

# Check resource limits
docker stats --no-stream <container>
```

### Source Code Scanning for Container Issues

```bash
# Secrets in Dockerfile
rg "(password|secret|token|api.key)\s*=" Dockerfile* docker-compose*

# Privileged mode
rg "privileged|SYS_ADMIN|cap_add" docker-compose* k8s/ -i

# Docker socket mount
rg "docker\.sock" docker-compose* k8s/

# Host mounts
rg "host.*path|hostPath" k8s/ docker-compose*

# Root user (missing USER directive)
# Check: does Dockerfile have USER instruction?

# Latest tag usage
rg ":latest" Dockerfile* docker-compose* k8s/

# Insecure registries
rg "http://" Dockerfile* docker-compose* k8s/ | rg -v "localhost"
```

---

## References

- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [Kubernetes Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [NIST SP 800-190: Container Security Guide](https://csrc.nist.gov/publications/detail/sp/800-190/final)
- [Snyk Container Security](https://snyk.io/learn/container-security/)
- [OWASP Docker Security Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [OWASP Kubernetes Security Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Kubernetes_Security_Cheat_Sheet.html)
