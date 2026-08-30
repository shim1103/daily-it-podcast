---
name: local secret 管理に AgentSecrets を採用する
date: 2026-08-15T17:48:16
branch: docs/agentsecrets-secret-management
---

## 1. Decision

local 開発の secret 管理に AgentSecrets（OSS）を採用する。運用は local + OS keychain 単独とし、復元性のため `secrets push`（zero-knowledge cloud sync）を有効化する。

## 2. Reason

無料・agent に消されない・消えても復元可能・漏洩しない、の 4 要件を同時に満たす候補は AgentSecrets のみだった。値が agent memory に載らない zero-knowledge proxy、OS keychain 保護、cloud sync による復元、OSS 無料が揃う。

## 3. Rejected

- 素の `.env`（agent が context / commit / log へ流出させる事例が複数確認された）
- git 暗号化 commit 系（SOPS+age / git-crypt）（file 自体が残り agent 削除可、復元が git 履歴依存で GitHub push を強要し「push 非必須」と不整合）
- Bitwarden CLI / 1Password CLI 等の vault 型（取得後の値が agent memory に載る。1Password は有料で予算外）
