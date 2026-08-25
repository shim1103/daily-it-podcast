---
title: playback web の UI は一覧 page 1 つに統合し、component の組み合わせで一覧・詳細・audio を表す
date: 2026-08-25T05:10:48+09:00
branch: feature/playback-ui-structure
---

## 1. Decision

1. UI は 1 page（一覧 page）に統合する。`#/episodes/{episodeId}` への page 遷移、詳細専用 page（`episode-detail-page.ts`）は作らない。episode 選択状態は一覧 page 内の 1 つの ViewModel state（`selectedEpisodeId` 相当）で持つ
2. 各 component は「API response の該当 field をそのまま描画する」契約（Contract Freeze）を持つ、判断・変換を含まない最小単位に分ける。一覧・詳細の見た目は、この最小単位の component を組み合わせて表す。同じ field を複数 component で重ねて描画しない（DRY）
   - `episode-list-item`：`EpisodeListItem`（`episodeId`・`date`・`title`・`durationSec`）をそのまま描画する
   - `episode-topic`：`topics[]` の 1 要素（`title`・`preface`・`detail`）をそのまま描画する。`startSec` はこの contract に含めない
   - `episode-player`：`audioRef` から `<audio controls src>` を組み立てるだけの component。分岐を持たない
   - `title`・`date` は `episode-list-item` が一覧行として既に描画しているため、選択展開した詳細側で重ねて描画する component（`episode-header` 相当）は置かない
3. 一覧 component（`episode-list`）は `episode-list-item` を episodes 配列の順に並べるだけ。選択中の episode がある場合、その位置に詳細 component 群（`episode-manuscript`（`episode-topic` の組み合わせ）+ `episode-player`）を展開する。`title`・`date` は展開しない
4. audio 取得方式・実行契機は `2026-08-25T04-10-02` 相当の決定を維持する：`<audio src>` 直結、`fetchAudio()` は使わない、取得契機は詳細表示（この decision では「選択状態への遷移」）時の DOM 挿入に一本化する
5. seek 機能・見た目の装飾は今回の scope に含めない

## 2. Reason

1. 「一覧 page 内で完結」という2026-08-25の指示は、hash routing による別 page 遷移（既存実装）と、modal による開閉状態の二重管理のどちらとも異なる第3の形。1 つの選択状態だけを持つ inline 展開が、既存の ViewModel パターン（1 state を subscribe で配信する）に最小の変更で乗る（KISS）
2. component を「API response の field をそのまま描画する」契約単位まで分解すると、一覧・詳細という 2 つの見た目は同じ component 群の組み合わせ方の違いに還元できる。契約と組み合わせを分離すると、見た目の変更（今回の 1 page 統合）が個々の component の contract を破らずに済む
3. `episode-list-item`・`episode-topic`・`episode-player` は判断・分岐を持たないため、Contract Freeze（A）としてそのまま固定できる。加工を要する値（`durationSec` の表示形式変換等）はこの contract に含めず、必要になった時点で純粋関数層へ分離する
4. `title`・`date` を選択展開側でも描画する案（`episode-header`）は、一覧行に既に同じ値が見えている状態で同じ値を画面上に重複表示することになり、DRY（同一知識を複数箇所に表現しない）に反する。API response の同じ field を 2 つの component 契約に重ねて持たせない

## 3. Rejected

1. 別 page 遷移を維持したまま一覧へ「戻る」導線を強化する案 — 「一覧 page 内で完結」という明示指示と矛盾する
2. modal で詳細を開閉する案 — 開閉状態と選択状態という 2 つの state を持つことになり、1 状態で足りる inline 展開より複雑
3. `episode-list`・`episode-detail` を今のまま個別 Feature Component として維持し、中身だけ増やす案 — 「現状 component もそれらの組み合わせとして表示させろ」という指示は、既存 2 component をそのまま拡張するのではなく、より小さい contract 単位への再分解を求めている
4. 選択展開した詳細側にも `title`・`date` を描画する `episode-header` component を置く案 — 初期実装で採用したが、一覧行（`episode-list-item`）に同じ値が既に表示されているため、選択展開すると同じ episode の `title`・`date` が画面上に 2 回現れる。DRY 違反として撤回し、`episode-header` は削除した
