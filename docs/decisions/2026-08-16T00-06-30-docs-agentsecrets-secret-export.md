---
name: local 秘密は AgentSecrets の名前参照で渡し env export しない
date: 2026-08-16T00:06:30
branch: docs/agentsecrets-secret-export
---

## 1. Decision

local（DEV / STAGING 検証）の秘密は AgentSecrets（OS keychain + zero-knowledge sync）に置く。アプリの認証 HTTP はキー**値**を process.env に載せず、AgentSecrets HTTP proxy / `call` / MCP の**キー名注入**で渡す。shim と agent で渡し方を分けない（local は名前参照で一本化）。

GHA / Workers の native secrets への寄せは現要件外とし、この判断の対象外とする。

併用防御として、repo 配下の agent 設定で `.env*` / `secrets/` / `~/.ssh` に加え、`printenv` / `env` 等の environ dump を deny する。soft instruction（CLAUDE.md 等）は enforcement に数えない。

## 2. Reason

素の `.env` / vault CLI 取得後の env 載せは agent 文脈へ値が流れうる。AgentSecrets の proxy はキー名だけを受け、値は keychain から注入する。公式も Cursor/Claude 向け MCP と HTTP Proxy（`X-AS-*`）を推奨しており、Go など任意 HTTP client は Proxy 経路が正である。

## 3. Rejected

- shell への `export` や恒久 `.env` を正とする案
- shim だけ `agentsecrets env --`、agent だけ proxy、と実行主体で二重経路を持つ案（保証にならず、公式の HTTP client 経路とも二重化）
- STAGING/PROD 実行を今すぐ AgentSecrets 単一化する案（基盤 native が既定で、無人 keychain 解錠の弱点がある）
