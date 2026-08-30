---
name: Playback runtimeごとにsecret設定のSSOTを閉じる
date: 2026-08-19T17:37:00
branch: feature/playback-worker-http-refactor
---

## 1. Decision

1. secret name と設定存在の契約は runtime ごとに所有する。
2. Playback Worker は Cloudflare Workers の `env` と、Worker内の `PlaybackEnv` を所有する。
3. Generator は GHA / AgentSecrets と Go 側の設定定義を所有する。
4. Playback Web は Drive secret を持たず、Worker endpoint の `baseUrl` だけを受け取る。
5. `README.md` は全runtimeのsecret名を一覧する運用文書であり、各runtimeの実行時SSOTではない。

## 2. Reason

runtimeごとに注入経路・必要なsecret・型が異なるため、repo全体の共通secret schemaを作ると不要な結合が生じる。
必要最小限のruntimeが自身のconfigを検証する方が、DRYとKISSを保ち、secretの過剰な共有も防げる。

## 3. Rejected

1. Generator、Playback Worker、Playback Webで同じsecret定義moduleを共有する案。
2. Playback WebへDrive secretを注入し、Web ClientからDriveへ直接接続する案。
3. `README.md` の一覧だけを実行時の型・存在検証として扱う案。

