# 資安審核技能（Security Audit Skill）

> **版本**：3.2.0 | **Language**: [English](README.md)

一套完整的 AI Agent 技能，用於對程式碼庫進行多平台資安審核。專為 Cursor IDE 設計——當使用者要求資安檢查時，AI Agent 會自動啟用此技能。

## 功能概述

執行結構化、可重複的資安審核，涵蓋 **6 種平台類型**，產出可執行的弱點報告，包含嚴重等級、修復建議與優先行動計畫。

| 平台 | 範例 | 重點檢查領域 |
|------|------|-------------|
| **前端** | React, Vue, Angular | XSS、Token 儲存、CSP、HTTP 安全標頭 |
| **後端** | Python (FastAPI), Node.js, Go, Java | 注入攻擊、認證授權、序列化安全、資料庫/快取安全 |
| **桌面應用** | Electron, Native 應用 | nodeIntegration、IPC 安全、Preload、自動更新、程式碼簽章 |
| **容器** | Docker, Kubernetes | Dockerfile、映像檔供應鏈、執行期強化、機密管理 |
| **AI/LLM** | ChatGPT 整合、LangChain | Prompt Injection、輸出安全、模型存取控制、資料隱私 |
| **跨層** | 全端組合分析 | Token 流轉、CORS 一致性、錯誤處理鏈、攻擊路徑分析 |

## 運作方式

技能遵循 **六階段審核流程**：

```
階段 1          階段 2          階段 3          階段 4          階段 5          階段 6
情境建立   ──▶  威脅建模   ──▶  系統性掃描 ──▶  深度分析   ──▶  報告撰寫   ──▶  回饋迴圈
```

1. **情境建立** — 了解專案的技術棧、架構、資料流與現有安全措施
2. **威脅建模** — 針對各平台套用 STRIDE 框架，辨識可能的攻擊向量
3. **系統性掃描** — 執行 ripgrep 模式與審核工具，依風險由高到低排序
4. **深度分析** — 以影響程度 × 利用難度矩陣評估每項發現，過濾誤報，執行跨層分析
5. **報告撰寫** — 產出結構化報告，使用 `SEC-XX-YYYY-NNN` 編號格式，附帶證據、修復程式碼與行動計畫
6. **回饋迴圈** — 記錄經驗、更新掃描模式、執行自我更新以取得最新情報

## 目錄結構

```
security-audit/
├── SKILL.md                             # AI Agent 核心指令（< 500 行）
├── README.md                            # 英文說明
├── README-TW.md                         # 本檔案（繁體中文）
├── references/
│   ├── checklist.md                     # 完整逐階段檢查清單（Phase 0–12）
│   ├── scan-commands.md                 # 各平台 ripgrep 掃描模式
│   ├── report-template.md              # 報告結構模板
│   ├── ai-security.md                  # AI/LLM 安全深度指南
│   ├── electron-security.md            # Electron 桌面應用安全指南
│   ├── container-security.md           # 容器/Docker/K8s 安全指南
│   ├── self-update-guide.md            # 自我更新機制說明
│   └── changelog.md                    # 版本變更紀錄
└── scripts/
    └── self-update.py                  # 自動抓取安全情報並產生更新提案
```

## 快速開始

### 觸發技能

在 Cursor 中向 AI Agent 提出要求即可：

```
「請對這個專案做資安檢查」
「幫我做弱點掃描」
「檢查一下程式碼有沒有安全問題」
"Please perform a security audit on this project"
```

Agent 會自動依循六階段流程並產出報告。

### 自我更新

保持技能知識庫與最新安全情報同步：

```bash
# 列出所有情報來源（共 11 個：OWASP, MITRE, Anthropic, Docker, K8s 等）
python3 .cursor/skills/security-audit/scripts/self-update.py --sources

# 預覽會抓取哪些來源（不發出網路請求）
python3 .cursor/skills/security-audit/scripts/self-update.py --dry-run

# 執行完整更新：抓取最新情報並產生更新提案
python3 .cursor/skills/security-audit/scripts/self-update.py

# 將提案儲存至檔案以供審閱
python3 .cursor/skills/security-audit/scripts/self-update.py --output-dir docs/security-updates/
```

腳本從 11 個來源抓取資料，產生結構化 AI 提示，由 Agent 建議具體的參考檔案變更。所有更新都需人工審核——不會自動套用。

### 自我更新流程圖

```
                        ┌──────────────┐
                        │  執行腳本     │
                        │ self-update  │
                        └──────┬───────┘
                               │
                    ┌──────────▼──────────┐
                    │ 抓取 11 個安全情報來源 │
                    │ OWASP, MITRE, NIST  │
                    │ Docker, K8s, etc.   │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │ 產生結構化 AI 提示    │
                    │ 比對現有技能內容      │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │ AI Agent 分析差異    │
                    │ 建議具體更新內容      │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │ 人工審核             │
                    │ 核准 → 套用變更      │
                    │ 拒絕 → 保持不變      │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │ 更新 changelog.md   │
                    │ 版本號遞增          │
                    └─────────────────────┘
```

## 嚴重等級定義

| 等級 | 定義 | 修復時限 |
|------|------|---------|
| **CRITICAL** | 可直接導致資料洩漏、帳號劫持或遠端程式碼執行 | 24 小時內 |
| **HIGH** | 在特定條件下可被利用 | 1 週內 |
| **MEDIUM** | 增加攻擊面或降低防護效果 | 1 個 Sprint 內 |
| **LOW** | 未遵循安全最佳實踐，但無直接威脅 | 下次維護週期 |
| **INFO** | 建議性改善 | 視資源排定 |

## 發現編號格式

| 前綴 | 平台 |
|------|------|
| `SEC-FE-YYYY-NNN` | 前端 |
| `SEC-BE-YYYY-NNN` | 後端 |
| `SEC-DT-YYYY-NNN` | 桌面應用（Electron / Native）|
| `SEC-CT-YYYY-NNN` | 容器（Docker / K8s）|
| `SEC-AI-YYYY-NNN` | AI / LLM |
| `SEC-FS-YYYY-NNN` | 跨層（全端）|

## 工具生態系

| 類別 | 工具 |
|------|------|
| 機密掃描 | `gitleaks`, `trufflehog` |
| 依賴項審核 | `npm audit`, `pip-audit`, `safety`, `snyk` |
| 靜態分析 | `semgrep`, `bandit`, `eslint-plugin-security` |
| 容器安全 | `trivy`, `grype`, `hadolint`, `dockle`, `Falco` |
| K8s 安全 | `kubesec`, `kube-bench`, `polaris` |
| Electron 審核 | `electronegativity`, ASAR 解包檢查 |
| AI/LLM 測試 | `Petri`, `garak`, `promptfoo` |
| 動態掃描 | `OWASP ZAP`, `Burp Suite` |

## 參考標準

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP LLM Top 10](https://owasp.org/www-project-top-10-for-large-language-model-applications/)
- [OWASP API Security Top 10](https://owasp.org/API-Security/)
- [Electron 安全檢查清單](https://www.electronjs.org/docs/latest/tutorial/security)
- [CIS Docker 基準](https://www.cisecurity.org/benchmark/docker)
- [CIS Kubernetes 基準](https://www.cisecurity.org/benchmark/kubernetes)
- [NIST SP 800-190：容器安全指南](https://csrc.nist.gov/publications/detail/sp/800-190/final)
- [MITRE ATLAS](https://atlas.mitre.org/)
- [STRIDE 威脅模型](https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats)

## 版本歷史

詳見 [references/changelog.md](references/changelog.md)。

| 版本 | 日期 | 重點 |
|------|------|------|
| v3.2.0 | 2026-02-25 | 容器/Docker/K8s 安全、腳本使用說明 |
| v3.1.0 | 2026-02-25 | AI/LLM 安全、自我更新機制 |
| v3.0.0 | 2026-02-25 | 桌面應用/Electron 安全 |
| v2.0.0 | 2026-02-25 | 後端 API 安全 |
| v1.0.0 | 2026-02-25 | 首次發布（前端）|

## 授權

專案內部工具，隨程式碼庫一同分發。
