---
name: AgentSecrets HTTP proxy 正本吸収と command 素材の配置固定
date: 2026-08-26T12:15:59
session_id: none
branch: feature/generator-agentsecrets-http-proxy-absorb
prev: なし
---

## 1. Summary

issue-manager で AgentSecrets HTTP proxy を `secrettransport/agentsecrets` の単一実装へ吸収し、旧 `infrastructure/agentsecrets` HTTP API を削除した。続けて EnvWrapper を `commandlaunch/agentsecrets` へ移し、袋 package を消した。gate 3つ pass。PR 作成まで進めた。

## 2. Changes

1. wrap を解消し PROXY `X-AS-*` 組み立てを `secrettransport/agentsecrets.Client` が直接所有する形へ書き換えた。review 後に DefaultProxyURL test を RoundTripper 証拠へ強化した。
2. EnvWrapper を `commandlaunch/agentsecrets` へ移し、`infrastructure/agentsecrets` を空にして削除した。所有境界の Decision と lane を更新した。
3. static / unit / integration gate を pass 確認した（unit coverage 91.2%）。
4. PR #60 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `1575d53`
- `f54b06e`
- `c7aa125`
- `919239c`
