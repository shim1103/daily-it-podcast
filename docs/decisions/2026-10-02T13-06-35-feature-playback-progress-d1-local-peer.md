---
name: 進捗のmodeがin-memoryの時は動くin-memory実装を使い、何も永続しないno-opのStubを廃止する
date: 2026-10-02T13:06:35
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. Root は、進捗の mode が `in-memory` の時（episode の mode とは独立した明示設定。`2026-10-02T16-13-40-feature-playback-progress-d1-local-peer.md`）、Map で動く `ProgressRepository`（episode の `InMemoryEpisodeRepository` と並ぶ in-memory 実装）を選ぶ。何も永続しない no-op の `StubProgressRepository` は廃止する。
2. UseCase test 用の Fake（`application/ports/progress-repository.fake.ts` の `createFakeProgressRepository`）と、in-memory mode の実装は、実体を同じにする。重複して持たない。
3. epic の `2026-10-01T18-54-14-feature-playback-progress-application.md` §1.4 の「`StubProgressRepository` は残してよい」は、Stub の廃止で使われなくなる。epic の file は編集せず、ここから置換を明記する。

## 2. Reason

1. episode と progress の両方を in-memory にした構成では、episode は動く in-memory 実装、progress は no-op の Stub という非対称がある。同じ in-memory の中で、片方は状態を持ち、片方は持たない。
2. no-op の Stub は、永続しない成功応答を返すだけである。`create → pull` のように、状態が次の呼び出しに残る流れを、実 UseCase 経由で通す検証には使えない。動く実装なら、local や直接組み立てる test で、その流れを通せる。
3. 本番経路（`playback-controllers-middleware.ts`）は episode `r2` ＋ 進捗 `d1` の明示で固定され、進捗の in-memory は `createPlaybackControllers` を直接呼ぶ test と local に限られる。Controller は request ごとに組み立てられるので、in-memory の状態は同一の組み立ての中だけで保たれる。本番の挙動を変えず、影響は test と local に閉じる。
4. UseCase test 用の Fake と同じ実体にするのは、同じ「Map で動く ProgressRepository」を 2 箇所に実装しないためである（`design-philosophy.md` §2-2 DRY）。片方だけ修正して食い違う余地を残さない。
5. epic の §1.4 は Stub を「残してよい」と許すだけで、残す必然を示していない。必然が無く、検証にも使えない Stub は、廃止の方が単純である（`design-philosophy.md` §2-3）。

## 3. Rejected

1. **Stub を維持する案** — epic の §1.4 は「残してよい」と許すだけで、残す必然が無い。永続しない応答を返すだけで、`create → pull` の検証に使えない。
2. **Fake を test support のままにして、本番 tree の in-memory mode と二重に持つ案** — 同じ Map 実装が 2 箇所になる（`design-philosophy.md` §2-2 DRY）。
