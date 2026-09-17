---
name: System test の既定 topic 数を疎通確認用の 1 にする
date: 2026-09-17T14:41:00
branch: chore/system-topic-default-one
---

## 1. Decision

1. `generator-system.yml` と `produce_episode_system_test.go` が使う既定 topic 数を 1 にする。
2. `workflow_dispatch` の `topic_count` 明示指定は維持する。
3. 本番 `DraftTopicCountTarget` は変更しない。

## 2. Reason

1. System test の責務は実情報源から R2 書込までの疎通であり、本番 topic 数ぶんの原稿生成と TTS を毎回実行する必要はない。
2. 1 topic でも TextWriter、TTS、mp3 encode、R2 書込の境界をすべて通る。
3. master push ごとの CD 起点なので、既定費用と所要時間を抑える。

## 3. Rejected

1. 未指定時に本番 `DraftTopicCountTarget` を使う案 — 疎通確認を超える API 消費と実行時間を毎 release に課す。
2. topic 数を 1 に固定して上書きを禁止する案 — 長い経路を人手で再確認する入口を失う。
