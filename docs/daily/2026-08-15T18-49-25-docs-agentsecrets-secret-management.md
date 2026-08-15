---
name: AgentSecrets 採用と project 単位の secret deny
date: 2026-08-15T18:49:25
session_id: none
branch: docs/agentsecrets-secret-management
prev: なし
---

## 1. Summary

local secret 管理に AgentSecrets を採用し、`.env` / `secrets/` / `~/.ssh` の agent 読み取り deny を repo 配下へ置いた。project 作成と dummy 含む secret 登録まで通しで確認した。

## 2. Changes

- AgentSecrets 採用の decision を追加した
- Claude / Cursor / Codex の deny を project 配下へ追加した（global 設定は撤回）
- `.agentsecrets/project.json`・workflow・`.gitignore`・`.env.example` を追加した
- `daily-it-podcast` project を作成し、DUMMY 以外に GETX / TWITTER_IO の 2 key が remote にあることを確認した

### Commits

- `7d2244d` — docs(decisions): local secret 管理に AgentSecrets を採用する
- `a446d34` — chore: .env / secrets / ~/.ssh の agent 読み取りを project で deny する
- `df7f962` — chore: AgentSecrets の project 紐付けと .env 系 ignore を追加する
