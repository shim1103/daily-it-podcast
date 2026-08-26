---
name: ViewModel hook 化と Hono RPC request/warranty 分離
date: 2026-08-26T19:32:00
session_id: none
branch: feature/playback-web-view-model-react-hooks
prev: なし
---

## 1. Summary

issue-manager で ViewModel を React hooks へ書き換え、API Client を Hono RPC request と warranty に分け、path SSOT と AppType method chain を Decision に固定したうえで pr-completion まで進めた。

## 2. Changes

1. `useEpisodeListViewModel` へ移行し、未 JSX page 向けに `mount-episode-list-view-model`（createRoot 一時橋）を置いた。
2. `createPlaybackRpcClient` が encode 付き request を持ち、`createPlaybackApiClient` は `readJsonResult` + Zod のみにした。
3. contracts に route template を置き、worker `app` を method chain 化して `hc<AppType>` を型 safe にした。
4. Hono 境界・AppType chain・ViewModel 橋の Decision 3本と lane / 依存 Issue を完了状態へ揃えた。
5. gate（typecheck / unit 219 / lint:layers / integration）を pass 確認した。
6. PR #68 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `b0ba6e5`
- `359c3da`
- `f901d0e`
- `b1743b1`
- `6c154ff`
