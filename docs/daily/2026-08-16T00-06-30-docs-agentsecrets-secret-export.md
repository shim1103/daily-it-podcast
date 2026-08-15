---
name: AgentSecrets キー名注入 client と deny 強化
date: 2026-08-16T00:06:30
session_id: none
branch: docs/agentsecrets-secret-export
prev: なし
---

## 1. Summary

local 秘密の渡し方を AgentSecrets の名前参照（HTTP proxy）に固定する方針を固め、generator に proxy client とキー名定数を追加した。agent 設定へ env dump deny を足し、Unit/Integration 入口の空 package skip を直して hook 経由の commit/push を通した。shim 手元では proxy 経由の GetX 呼び出しが成功した。続けて `origin/develop` を merge して conflict を解消し、PR #12 を作成。Copilot 指摘（host 検証・キー名定数 rename）を反映した。

## 2. Changes

- `infrastructure/secretnames` に共有キー名定数を追加
- `infrastructure/agentsecrets` に proxy 向け HTTP client（`X-AS-*` キー名注入）と sociable unit test を追加
- Claude / Codex / Cursor 設定へ `printenv` / `env` 系 deny を追加
- `scripts/test-unit.sh` / `test-integration.sh` を復元し、空 package を skip するよう修正

### Commits

- `e1c4b1f` — chore(scripts): Unit 入口を復元し空 package を skip する（agent deny 設定も同 commit に含む）
- `f653a8e` — feat(generator): AgentSecrets proxy 経由のキー名注入 HTTP client を追加する
- `10a2bdb` — chore(scripts): Integration 入口でも空 package を skip する
- `2721739` — docs(log): セッションログ
- `b5ca44d` — merge: origin/develop を取り込み conflict を解消する
- `760cb94` — fix(generator): Copilot 指摘の URL host 検証とキー名定数を直す

追記 Changes:
- PR #12（base: develop）を `gh pr create` で作成し CI SUCCESS
- lessons / scripts の merge conflict を両側保持＋空 package skip 維持で解消
- Copilot: `url.Parse` + Host 非空、`GetXAPIKey` → `GetXAPIKeyName`
