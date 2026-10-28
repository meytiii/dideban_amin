# Security Audit Skill — Changelog

## v3.2.0 — 2026-02-25
- **Trigger**: 補充 Container/Docker/Kubernetes 安全檢查 + Script 使用說明
- **Changes**:
  - 新增 Phase 10 — Container Security（Dockerfile, Image Supply Chain, Runtime, Secrets, Network, K8s, CI/CD, Monitoring）
  - 新增 `references/container-security.md` 完整容器安全指南
  - STRIDE 威脅模型新增 Container 欄位
  - 報告模板新增 Container 欄位
  - 掃描命令新增 Docker/K8s 模式
  - 檢查清單新增 Phase 9 Container（30+ 檢查項目）
  - 工具清單新增 `hadolint`, `dockle`, `kubesec`, `kube-bench`, `polaris`, `cosign`, `Falco`, `Sysdig`
  - SKILL.md 新增 `scripts/self-update.py` 完整使用說明（Prerequisites, Commands, Workflow, 觸發時機）
  - Finding ID 新增 `SEC-CT-YYYY-NNN` (container)
- **Source**: Docker Security Best Practices, CIS Docker/K8s Benchmarks, NIST SP 800-190, OWASP Container Security

## v3.1.0 — 2026-02-25
- **Trigger**: Anthropic Claude Code Security 發布 + AI 安全市場分析
- **Changes**:
  - 新增 Phase 11 — AI/LLM Security（Prompt Injection, Model Access, Output Safety, Data Privacy, AI Supply Chain, AI-Assisted Security Tools）
  - 新增 `references/ai-security.md` AI 安全深度指南
  - STRIDE 威脅模型新增 AI/LLM 欄位
  - 報告模板新增 AI/LLM 欄位
  - 掃描命令新增 AI/LLM 模式
  - 工具清單新增 `Petri`, `garak`, `promptfoo`
  - 新增 `scripts/self-update.py` 自我更新機制
  - 新增 `references/self-update-guide.md`
- **Source**: Anthropic 安全工具分析報告, Claude Code Security 技術報告, 市場趨勢分析

## v3.0.0 — 2026-02-25
- **Trigger**: 納入桌面應用（Electron Mac App）審核
- **Changes**:
  - 新增 Phase 8 — Desktop/Electron Security
  - 新增 `references/electron-security.md`
  - STRIDE 新增桌面欄位
  - 掃描命令新增 Electron 模式
  - 工具新增 `electronegativity`, ASAR inspection, `codesign`
- **Source**: maine-coon-cat (Electron 28) 審核經驗

## v2.0.0 — 2026-02-25
- **Trigger**: 納入後端專案（Python FastAPI）審核
- **Changes**:
  - 新增 Phase 7 — Backend Security
  - 所有 Phase 新增後端對應檢查
  - 掃描命令新增 Python 模式
  - 工具新增 `bandit`, `pip-audit`, `safety`, `semgrep`
- **Source**: ocelot (FastAPI) 審核經驗

## v1.0.0 — 2026-02-25
- **Trigger**: 首次建立資安檢查方法論
- **Changes**: Phase 0–9 完整流程（前端為主）
- **Source**: project-name (React SPA) 審核經驗
