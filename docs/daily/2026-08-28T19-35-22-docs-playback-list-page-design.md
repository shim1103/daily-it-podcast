---
name: playback list の視覚と concept Decision、topics 契約 task
date: 2026-08-28T19:35:22
session_id: 51129181-5b1f-4897-8238-8b39adb9beff
branch: docs/playback-list-page-design
prev: なし
---

## 1. Summary

Playback list-page の見た目（夜・紫黒・Apple list 行）と表示整形 utils、fake 複数 episode、concept / 視覚言語 Decision、listEpisode topics title の達成契約 task を固定した。decision 2本の DRY 重複を責務分離で直し、README / DESIGN には動線だけを足した。pr-completion で PR を作成する。

## 2. Changes

1. date / duration の表示整形を web utils へ切り出し、list item は整形結果だけを描画する。
2. global / list / item の CSS を分け、Pico classless の card・main 幅が生む黒 gap を潰し、左 edge・play pill・余白を載せた。
3. worker fake を JSON 複数件へ移し、dev 一覧の見た目確認を可能にした。
4. Decision `2026-08-28T19-20-00`（concept / setting / motif）と `2026-08-28T19-20-01`（視覚言語）を分離し、motif / 物語の再掲を落として DRY を直した。
5. `docs/tasks/todo/playback-list-episodes-topics-titles.md` に listEpisode `topics: { title }[]` の server+API 契約を残した（web 表示は Out of Scope）。
6. README / DESIGN へ decision path の動線だけを追加。concept 本文は写さない。
7. 途中で `create-issue` の template 参照を GitHub Issue 作成と誤読し Project #9 を作ったが、shim 指摘後に削除。local `project.toml` は未作成のまま。

### Commits

- `59b9ca7`
- `f663239`
- `0ec9bff`
- `27865e8`
