---
name: Generator Drive 保存は schema 検証・OAuth・REST を層で分離する
date: 2026-08-19T15:00:00
branch: feature/generator-drive-adapter
---

## 1. Decision

generator の Drive 書込を次の 3 機能に分離する。

1. **Application:** `manuscript.schema.json` による原稿検証（Domain Error）。repo 根 `contracts` を import する読み手は **Application** とする（generator について）。
2. **Infrastructure（OAuth）:** Google OAuth refresh。AgentSecrets 経由。別 task（`docs/tasks/todo/generator-google-oauth-adapter.md`）。
3. **Infrastructure（Drive 保存）:** `DRIVE_FOLDER_ID` folder 直下へ `{episodeId}.json` と `{episodeId}.wav` を put するだけ。schema の field 意味は知らない。

Port は **`EpisodeWriter.Write(ctx, episodeID, manuscript, audio)` を維持**する。`PutFile` 単位の Application Port には下げない。

Drive HTTP client（list / create / upload）は **file name・MIME・byte のみ**を扱い、json/wav の domain 意味は知らない。`EpisodeWriter` の Port 実装が `drive-layout.md` の命名（拡張子・stem）を map する。

保存先は **`DRIVE_FOLDER_ID` 指定 folder の直下フラット**。サブ folder は作らない。

本判断は `2026-08-19T13-00-00` と `2026-08-19T13-31-00` の「generator Infrastructure が schema を import して enforce」を **上書き**する。読み手の表は `DESIGN.md` を正とする。

## 2. Reason

schema 検証は Domain 知識であり Application の責務。OAuth refresh は Drive I/O と vendor 契約が異なる mechanism。保存 Adapter に混在すると層違反と test の肥大化になる。

`drive-layout.md` の配置（拡張子・folder 直下）は wire 配置であり、HTTP put に必要な最小知識として Port 実装が持つ。`manuscript.schema.json` の field 検証と混同しない。

`contracts` package の `go:embed` と snapshot 禁止は維持する。読み手が Infrastructure から Application へ移るだけ。実装トラックは GitHub Issue（`docs/tasks/todo/generator-lane.md` 参照）。

## 3. Rejected

1. generator の Drive 保存 Adapter が引き続き schema を import して enforce する案（`2026-08-19T13-*` の方針。Application へ移す）
2. Application Port を `PutFile(name, mime, bytes)` に下げ、命名を UseCase が毎回組み立てる案（UseCase が HTTP 配置の詳細を持ちすぎる）
3. episode ごと sub folder を `drive-layout` に追加する案（Reader の `*.json` 列挙と playback 契約を広げる）
4. folder 直下方針をやめ、Drive 全体を name だけで upsert する案（layout 契約と乖離。create 時の `parents` は維持）
