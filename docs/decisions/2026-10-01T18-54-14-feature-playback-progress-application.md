---
name: ProgressRepository Portは永続能力のみ。merge・冪等・完走判定はApplication。completeのdurationSecは原稿取得
date: 2026-10-01T18:54:14
branch: feature/playback-progress-application
---

## 1. Decision

1. `ProgressRepository` Port は永続能力のみとする。`writeProgress(command)`・`completeProgress(command)` は置かない。競合の再試行（`2026-10-01T18-54-16-feature-playback-progress-d1-local-peer.md`）と cursor の pull（`2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md`）のために、版つき取得・条件付き書込・cursor 差分の面を、旧面（`upsertProgress`・`listUpdatedSince`）と並べて足す。旧面は切替まで動き続け、切替の時に消す（`2026-10-02T17-18-26-feature-playback-progress-d1-local-peer.md`）。操作名の正本は Port の定義で、ここへ写さない。
2. merge（`clientAt` 先勝ち／後勝ち）・冪等 create・行なし update の 404・完走ゾーン判定は Application UseCase の責務とする。
3. complete の `durationSec` は `EpisodeRepository` から原稿 1 件取得（`getManuscript`）して得る。HTTP body / 進捗 store へ `durationSec` を載せない。
4. no-op の `StubProgressRepository` は廃止する。進捗の mode が `in-memory` の時は Map で動く in-memory 実装（`InMemoryProgressRepository`）を選び、UseCase test 用 Fake と実体を同じにする。behavior 検証はこの実装が持つ（`2026-10-02T13-06-35-feature-playback-progress-d1-local-peer.md`）。A の契約面へ Fake を先置きしない点は変わらない。
5. pull の差分印は、時刻（`lastPlayedAt > since`）ではなく DB 採番の `seq` とし、契約は 10 進文字列の `cursor` を使う（`2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md` が正）。切替の実施は後続 Issue #220 で、現行の route・UseCase は `since` のままである。

## 2. Reason

1. architecture/backend/application §5 は Port を capability（永続能力）として置き、HTTP 操作差を Port へ漏らさない。`writeProgress` / `completeProgress` を Port に残すと create・update・complete の HTTP 差が永続面に染み、Application の merge・冪等・404・完走判定の置き場が消える。
2. Decision `2026-09-26T03-48-06` は Application が Write/pull merge と list embed を持つ境界とし、Fake は A に置かない。merge・冪等・行なし 404・完走ゾーンを Application に置くと、その境界と Decision `2026-09-22T19-13-35` / `2026-09-22T18-58-38` の意味が同じ層で閉じる。adapter / Fake へ merge を寄せると architecture と Application Issue 境界が矛盾する。
3. Decision `2026-09-19T16-25-18` は `durationSec` を原稿側正本のまま進捗 store へ複製しないと決めた。complete が要る全長は `EpisodeRepository.getManuscript` の 1 件取得で足り、HTTP に `durationSec` を足すと公開契約の新設と store 複製禁止の両方に触れる。`listManuscripts` 全走査は 1 episode の complete に対して過剰である。
4. A の契約面へ Fake を先置きしない（Decision `2026-09-26T03-48-06` と同軸）。no-op の Stub を廃止する理由と、in-memory 実装を Fake と同じ実体にする理由は `2026-10-02T13-06-35-feature-playback-progress-d1-local-peer.md` が持つ。
5. pull の差分印を `seq` と `cursor` にする理由は `2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md` が持つ。ここへ写さない。

## 3. Rejected

1. **merge を D1 adapter / Fake に置く案** — architecture と Application Issue 境界（Decision `2026-09-26T03-48-06`）と矛盾し、永続実装ごとに merge 意味が分岐する。
2. **Port に `writeProgress` / `completeProgress` の command 委譲を残す案** — HTTP 操作差が Port に漏れ、capability としての Port（§5）が壊れる。
3. **HTTP に `durationSec` を足す案** — 公開契約の新設禁止と、進捗 store へ `durationSec` を載せない Decision（`2026-09-19T16-25-18`）に反する。
4. **complete 毎に `listManuscripts` 全走査する案** — 1 件の全長取得に対して非効率。
5. **server でゾーン判定しない案** — 完走意味の AC と完走ゾーン定数が空になり、Decision `2026-09-22T19-13-35` の complete 責務が消える。
6. **no-op の `StubProgressRepository` を Composition 用 zero として残す案**（旧答え） — 永続しない成功応答を返すだけで、`create → pull` の検証に使えない。残す必然も示されていなかった。in-memory 実装へ置き換えた（`2026-10-02T13-06-35-feature-playback-progress-d1-local-peer.md`）。
7. **pull の差分鍵を `lastPlayedAt > since` にする案**（旧答え） — 遅れて届く書込と、`first_completed_at` だけが動く変更を、他端末が取りこぼす。ISO 文字列比較で足りるという前提は、取りこぼしの前では根拠にならなかった（`2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md`）。
