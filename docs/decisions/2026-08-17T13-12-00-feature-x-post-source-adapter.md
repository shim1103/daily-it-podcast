---
name: TwitterAPI.io は X-API-Key を proxy custom header で注入する
date: 2026-08-17T13:12:00
branch: feature/x-post-source-adapter
---

## 1. Decision

TwitterAPI.io 向け Adapter の認証は AgentSecrets proxy の **custom header 注入**だけを使う。

- upstream header: `X-API-Key`
- proxy 語形: `X-AS-Inject-Header-X-API-Key: TWITTER_IO_API_KEY`（値はキー名のみ）
- 秘密キー名の正は AgentSecrets 登録名 `TWITTER_IO_API_KEY`（README / `secretnames`）

## 2. Reason

TwitterAPI.io 公式 auth は `X-API-Key`（OpenAPI も同名）。`Inject.Bearer` は proxy が `Authorization: Bearer` を付ける経路であり、この vendor と合わない。AgentSecrets 公式は `X-AS-Inject-Header-<HeaderName>` を提供する。

## 3. Rejected

- `Inject.Bearer` で TwitterAPI.io を叩く案（upstream auth が一致しない）
- env に API key 値を載せる案（既存 decision: 名前参照のみ）
- Adapter 側で秘密値を保持する案
