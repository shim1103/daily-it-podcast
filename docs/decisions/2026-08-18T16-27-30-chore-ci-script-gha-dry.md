---
name: coverage 除外は file 単位。error.go 名では落とさない
date: 2026-08-18T16:27:30
branch: chore/ci-script-gha-dry
---

## 1. Decision

1. generator Unit coverage の除外は **Composition Root**（`internal/composition/**`）だけにする
2. `error.go` / `names.go` / `constants.go` を file 名で一括除外しない。各 file を見て、計測対象に残すか test で覆うかを決める
3. 今回の判定結果:
   1. `getxapi/error.go` — Error / Unwrap は既存 sociable で 100%。除外しない
   2. `twitterapiio/error.go` / `gemini/error.go` — 同型の薄い method。除外せず、失敗系 sociable で Error / Unwrap を呼ぶ
   3. `secretnames/names.go` / `entities/constants/*` / `gemini/constants.go` — 定数・名前だけ。Go の coverprofile に statement として出ない。除外行を足さない
4. `test/test-gate-scripts-narrow-integration.shell` は root 入口を再実行するだけなので削除する
5. 本 decision は `2026-08-17T14-45-00-chore-test-and-ci-coverage-layer` の「TwitterAPI.io の薄い Error method を除外する」を上書きする

## 2. Reason

1. 名前一致の除外は「この名前の file は常に薄い」という暗黙契約になり、中身が厚くなっても gate が気づかない
2. coverage skill の薄い wrapper は除外ではなく **100%** 対象。覆えるなら計測に残す
3. 定数の値写し test は minimization が禁ずる。coverprofile に出ないなら除外リストにも載せない
4. composer 契約は `scripts/test-gate-composer-sociable-unit.shell` が担う。`test/` からの二重実行は不要

## 3. Rejected

1. `error.go` / `names.go` / `constants.go` を glob でまとめて除外する案（暗黙依存）
2. twitterapiio / gemini の Error method を除外したままにする案（getxapi だけ計測する非対称）
3. 定数 file 用の存在確認 Unit を足して cover を稼ぐ案（値の二重管理）
