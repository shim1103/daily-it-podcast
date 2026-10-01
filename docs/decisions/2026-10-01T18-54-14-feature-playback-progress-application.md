---
name: ProgressRepository Portは永続能力のみ。merge・冪等・完走判定はApplication。completeのdurationSecは原稿取得
date: 2026-10-01T18:54:14
branch: feature/playback-progress-application
---

## 1. Decision

1. `ProgressRepository` Port は永続能力のみとする。操作は `getByEpisodeIds` / `upsertProgress(episodeId` + `EpisodeProgress)` / `listUpdatedSince`。`writeProgress(command)`・`completeProgress(command)` は置かない。
2. merge（`clientAt` 先勝ち／後勝ち）・冪等 create・行なし update の 404・完走ゾーン判定は Application UseCase の責務とする。
3. complete の `durationSec` は `EpisodeRepository` から原稿 1 件取得（`getManuscript`）して得る。HTTP body / 進捗 store へ `durationSec` を載せない。
4. `StubProgressRepository` は Composition 用 zero のまま残してよい。behavior 検証は Fake（C の test support）が持つ。
5. `listUpdatedSince` の差分鍵は `lastPlayedAt > since`（ISO 比較で足りる範囲）とする。

## 2. Reason

1. architecture/backend/application §5 は Port を capability（永続能力）として置き、HTTP 操作差を Port へ漏らさない。`writeProgress` / `completeProgress` を Port に残すと create・update・complete の HTTP 差が永続面に染み、Application の merge・冪等・404・完走判定の置き場が消える。
2. Decision `2026-09-26T03-48-06` は Application が Write/pull merge と list embed を持つ境界とし、Fake は A に置かない。merge・冪等・行なし 404・完走ゾーンを Application に置くと、その境界と Decision `2026-09-22T19-13-35` / `2026-09-22T18-58-38` の意味が同じ層で閉じる。adapter / Fake へ merge を寄せると architecture と Application Issue 境界が矛盾する。
3. Decision `2026-09-19T16-25-18` は `durationSec` を原稿側正本のまま進捗 store へ複製しないと決めた。complete が要る全長は `EpisodeRepository.getManuscript` の 1 件取得で足り、HTTP に `durationSec` を足すと公開契約の新設と store 複製禁止の両方に触れる。`listManuscripts` 全走査は 1 episode の complete に対して過剰である。
4. Stub は Composition の配線を通す zero。behavior の検証対象は Fake（C）に置き、A 契約面へ Fake を先置きしない（Decision `2026-09-26T03-48-06` と同軸）。
5. pull 差分は「いつ最後に進捗が更新されたか」で足りる。`lastPlayedAt > since` の ISO 文字列比較は、契約が ISO 時刻を扱う現状の範囲で十分である。

## 3. Rejected

1. **merge を D1 adapter / Fake に置く案** — architecture と Application Issue 境界（Decision `2026-09-26T03-48-06`）と矛盾し、永続実装ごとに merge 意味が分岐する。
2. **Port に `writeProgress` / `completeProgress` の command 委譲を残す案** — HTTP 操作差が Port に漏れ、capability としての Port（§5）が壊れる。
3. **HTTP に `durationSec` を足す案** — 公開契約の新設禁止と、進捗 store へ `durationSec` を載せない Decision（`2026-09-19T16-25-18`）に反する。
4. **complete 毎に `listManuscripts` 全走査する案** — 1 件の全長取得に対して非効率。
5. **server でゾーン判定しない案** — 完走意味の AC と完走ゾーン定数が空になり、Decision `2026-09-22T19-13-35` の complete 責務が消える。
