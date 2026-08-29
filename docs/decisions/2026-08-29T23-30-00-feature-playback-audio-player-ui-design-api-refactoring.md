---
name: playback 一覧 page の list/detail Feature と seek/audio を SRP+KISS で分割する
date: 2026-08-29T23:30:00
branch: feature/playback-audio-player-ui-design-api-refactoring
---

## 1. Decision

1. `episode-list.tsx` は episode 列挙・focus-mode フィルタ・detail slot の配置だけを持つ。detail の loading/error/success 分岐は内部に書かない
2. `episode-detail.tsx` を独立 Feature として切り出す（別 page/route ではない）。`episode-detail-loading.tsx`・`episode-detail-error.tsx`・`episode-detail-success.tsx` に分け、page 層の list loading/error と同型の status 分岐パターンに揃える
3. `seek` を `episode-list-page` に直書きしない。`use-episode-playback.ts` ViewModel hook が `audioElementRef` と `seek` を持つ。`episode-detail-success` は `audioElementRef` から内部で seek し、list 経由の `onSeek` prop relay を置かない
4. 共有 hidden `<audio>`（EpisodePlayer variant=audio）は `EpisodeList` に 1 つだけ置き、`useEpisodePlayback` で 1 回だけ配線する
5. `useEpisodeListViewModel` は list 取得・選択・detail 取得 state・`select`（hash routing 用）だけを持つ。audio/seek は持たない
6. 任意：`episode-selected-group.tsx` で紫枠 layout（item + detail slot）を切り出し、`episode-list` を簡潔に保つ
7. 先行 Decision `2026-08-25T05-10-48-feature-playback-ui-structure.md` の **1 page inline 詳細** は維持する。別 page/route への詳細分離は採らない

## 2. Reason

1. `feature-component.md` は Feature が表示依存を自分で持つことを求める。現状 `episode-list` が detail state 分岐まで知っており SRP を破っている
2. `page-route.md` は page が compose するだけと定める。list page の loading/error 分岐は page 層に既にある。detail 側も同じ status router パターンを Feature へ移すと、page と Feature の責務が揃う（KISS）
3. `view-model.md` は use case ごとに ViewModel を分ける。list 選択と audio seek は別関心事。`onSeek` を page → list → detail → manuscript と relay すると layer が増え、変更理由が横断する（DRY 違反ではないが SRP 違反）
4. hidden audio を page 合成に 1 つ固定すると、list item の play pill と detail の sequence bar が同じ ref を共有する契約が page 1 箇所で読める
5. 紫枠 layout を optional wrapper に切り出すと、`episode-list` の forEach 本体が「並べる／focus する／slot を置く」だけに戻る

## 3. Rejected

1. 別 detail page/route — `2026-08-25T05-10-48-feature-playback-ui-structure.md` の 1 page inline 方針と矛盾する
2. `SelectedEpisodeDetail` を `episode-list.tsx` 内に残す — detail state 分岐が list に残り、今回の SRP 分割目的を満たさない
3. `onSeek` prop relay（page → EpisodeList → detail → manuscript）— layer が多く、seek 変更が list props まで波及する
4. `seek` を `useEpisodeListViewModel` に戻す — list ViewModel と playback 操作が混ざる
