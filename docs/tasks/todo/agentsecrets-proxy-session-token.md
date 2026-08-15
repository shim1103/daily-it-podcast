## AgentSecrets proxy Client に session token を載せる

参照: docs/daily/2026-08-16T00-06-30-docs-agentsecrets-secret-export.md、AgentSecrets Architecture（proxy は session token を検証）

生の `localhost:8765/proxy` 直叩きは `X-AS-Session-Token` が必須。現状の Go Client は注入ヘッダのみで、session token 未設定。

- [ ] Client が keychain の `proxy_session_token`（または公式が示す供給源）から `X-AS-Session-Token` を付ける
- [ ] 成功系 unit / 必要なら proxy 前提の検証を更新する
- [ ] allowlist（例: `api.getxapi.com`）と policy approve の運用前提を README または decision 周辺に最小限で残すか判断する
