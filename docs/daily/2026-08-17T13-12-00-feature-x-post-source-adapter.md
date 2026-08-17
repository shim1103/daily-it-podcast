---
name: TwitterAPI.io PostSource Adapter と秘密名整理
date: 2026-08-17T13:12:00
session_id: none
branch: feature/x-post-source-adapter
prev: なし
---

## 1. Summary

TwitterAPI.io 向け `PostSource` Adapter を実装し、Composition から結線できるようにした。AgentSecrets proxy に custom header 注入を足し、vendor 公式の `X-API-Key` をキー名だけで渡す。秘密名の正を README に寄せ、完了した Adapter todo draft と session token todo を削除した。

## 2. Changes

- `agentsecrets.Inject.Headers` で `X-AS-Inject-Header-<name>` を載せる
- `infrastructure/x/twitterapiio` に `ListByUser`・Infrastructure Error・sociable unit test を追加
- `composition.NewTwitterAPIIOPostSource` で Adapter を結線
- `secretnames.TwitterIOAPIKeyName`（`TWITTER_IO_API_KEY`）を追加
- README の秘密名を AgentSecrets 登録名に揃え、DESIGN から秘密名列を外す
- `docs/tasks/todo/x-post-source-adapter.md` / `agentsecrets-proxy-session-token.md` を削除
- generator-lane / Issue2 draft の参照を実装 path に更新

### Commits

- `4994f03` — feat(generator): AgentSecrets proxy に custom header 注入を足す
- `4710de5` — feat(generator): TwitterAPI.io で PostSource Adapter を実装する
- `67bdf86` — docs: 秘密名を README に寄せ Adapter task を片付ける
