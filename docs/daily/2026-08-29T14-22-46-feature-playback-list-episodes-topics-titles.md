---
name: list episode に topic 題名を載せ、周辺の test fixture と型を整理
date: 2026-08-29T14:22:46
session_id: none
branch: feature/playback-list-episodes-topics-titles
prev: なし
---

## 1. Summary

`GET /episodes` の各 episode へ `topics: { title: string }[]` を追加し、契約と worker の listEpisode 経路を同じ到達で揃えた。web の list-item 表示は別 Issue に切った。あわせて、複数の sociable unit test が手写ししていた WAV サンプル byte を中立の共有 fixture へ一本化し、`ManuscriptSchema.safeParse` の戻り値へ型名を与えた。原稿の schema/domain 検証を Application 層へ移す設計判断を decision に固定し、実装 Issue を todo に起こした。

## 2. Changes

1. `episodeListItemSchema` に `topics: z.array(listTopicSchema)`（0 件以上、`{ title }` の strict object）を追加。get 用の `topicSchema` は流用せず list 用に別定義。
2. `GoogleDriveEpisodeRepository` / `InMemoryEpisodeRepository` の `listEpisodes` が、既読の原稿から `body.topics[].title` を射影して返す。追加 I/O なし。Port signature は不変。
3. 契約 reject test 6 件（topics 欠落 / 空配列許容 / 余剰 field / 空文字 title / 契約外 field / 正常）と、list の題名列が GetEpisode の題名列と順序込み一致することを検証する sociable unit test を GoogleDrive と InMemory の両方に追加。
4. web 側の list fixture 6 file に `topics` を追従（表示 assertion は不変）。
5. WAV magic 12 byte を `apps/playback/worker/src/test/fixtures/audio-bytes.ts` に集約し、`controllers/fake-use-cases.ts` から定義を外した。infra 層・use-case 層 test → controller 層 の逆向き import が消えた。
6. `ManuscriptSchema.safeParse` の無名 inline union に `ManuscriptParseResult` の型名を与え、`success: false` が失敗理由を持たない意図を doc 化。実行時挙動は不変。
7. `docs/decisions/2026-08-29T13-43-53-playback-manuscript-verification-application-layer.md` と `docs/tasks/todo/playback-application-infra-boundary.md` を新設。原稿 schema 検証を Application へ移し Port は保つ、`getEpisode` の失敗は粗い 1 Domain Error + message 分類、`listEpisodes` の除外を `@ensure` 明記、という判断を固定。
8. 役目を終えた `worker/src/application/.gitkeep` / `worker/src/entities/.gitkeep` を削除。
9. 全 commit で pre-commit hook（generator static/unit、playback format/lint/typecheck/lint:layers/unit 228 件）と pre-push hook が pass。
10. `feature/playback-list-episodes-topics-titles` を新規 branch として `origin` へ push。

### Commits

- `4625a08`
- `6e0c43b`
- `4290c70`
- `9f18905`
- `64dc93b`
