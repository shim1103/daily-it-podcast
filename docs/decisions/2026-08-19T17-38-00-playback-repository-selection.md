---
name: Playback Workerのrepository選択は本番設定を暗黙にFakeへfallbackさせない
date: 2026-08-19T17:38:00
branch: feature/playback-worker-http-refactor
---

## 1. Decision

1. Playback Workerの本番経路では、Drive接続に必要な4 keyが揃わない場合を `misconfigured` として扱う。
2. `InMemoryEpisodeRepository` はlocal development / unit test専用とし、本番Workerのsecret未注入時に自動選択しない。
3. repository選択のconfig completenessはComposition Rootが検証する。
4. Google OAuthの有効性とDrive API接続可否は、Google Drive Adapter / Integration境界の責務とする。
5. Route HandlerとControllerはDrive secretやrepository選択規則を知らない。

## 2. Reason

secret未注入の本番WorkerがInMemoryへ黙ってfallbackすると、空データを正常応答として返し、設定障害を隠す。
Composition Rootで設定の完全性を判定し、remote認証の検証をAdapter側へ分けると、各責務の変更理由が一つになる。

## 3. Rejected

1. 全key未設定を常にInMemoryへfallbackさせる案。
2. Route HandlerがDrive secretの存在を検証する案。
3. ControllerがOAuthやDrive APIの接続確認を行う案。
4. Composition RootがOAuth tokenの有効性まで保証する案。

