---
name: playback list/detail の視覚設計実装と dev fake 再生対応
date: 2026-08-29T19:12:00
session_id: none
branch: docs/playback-list-detail-page-design
prev: なし
---

## 1. Summary

playback list/detail page の視覚設計を実装した。list は番号付き title・topic 題名行、detail は専用 CSS・紫枠グループ・選択時 focus（他 episode 非表示）。topic 見出しの MM:SS ボタンから ViewModel の `seek` で再生位置を移動する。worker 側は fake 原稿の充実と、`durationSec` 一致の再生可能 WAV（8kHz・episode 別 cache）を追加した。`DESIGN.md` へ detail インラインの decision 参照を1行追記。

## 2. Changes

1. pre-commit で branch coverage 100% 未達が出たため、`seek` の audio 未接続分岐と fake WAV cache 命中を test で補った。
2. fake WAV は当初 1kHz でブラウザ再生不能だったため 8kHz へ変更。長尺 episode は初回生成で大きい byte になるが仕様通り。
3. 番号表示は list `n.　title`（全角スペース）、detail topic `n. title` を user 指定として固定。
4. unit test 242 件 green（commit 時 pre-commit 通過）。

### Commits

- `853a0ebedb8a71335c714d52efc5b55fa07443cb`
- `7a84b1fd91aae425b82f0fa2a4e267116ad783be`
