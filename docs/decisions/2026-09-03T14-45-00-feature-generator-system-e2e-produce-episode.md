---
name: generator-system の secret 付き test を TEST key 固定にし、cron gate から実行対象パターン切替を外す
date: 2026-09-03T14:45:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `generator-system.yml` が secret 付き test へ渡す credential は、Cursor / Gemini を含め **すべて test 専用（`TEST_*`）** にする。先行 Decision（`2026-08-30T17-45-00` / `2026-08-30T19-55-00`）で暫定例外としていた「Cursor / Gemini だけ本番 key」を終了する。
2. `generator-system.yml` の `workflow_dispatch` から実行対象を絞る input（`test_run` / 対応 env）を外す。cron・dispatch とも `-tags=system` の全 System test を実行する。単一 test を選んで回す用途は、後述の rate 計測を別 workflow へ分離することで不要にする。
3. `generator-system.yml` の inline Cursor CLI smoke step を削除する。smoke（binary + key の疎通）は一度通れば回帰頻度が低く、rate 計測 test の precondition 判定に畳める。

## 2. Reason

1. shim の明示指示「cursor, gemini も TEST 使ってください。勝手に本番 key 使わないで」。本番 key を CI の非本番経路へ通すと、本番の quota・課金・rate limit を test 実行が侵食し、本番 Run（`generator-produce-episode.yml`）の失敗要因になる。暫定例外は「TEST key だと API 到達失敗」という観測（run 33307390160 の Gemini 403 等）を根拠にしていたが、TTS を有料 key へ変更済みで前提が変わった。TEST key で到達できる構成へ寄せ、到達しないなら TEST key 側の登録を直す。
2. 実行対象パターン切替（`SYSTEM_TEST_RUN`）は「TTS 系を焼かずに CursorCLI 単体だけ回す」ために入れた（先行 Decision `2026-09-02T18-26-00` 系）。rate 計測を専用 workflow（`generator-tts-rate.yml` / `generator-draft-rate.yml`、`system && ratemeasure` tag、dispatch 専用）へ分ければ、cron gate は「全 System test を毎回」の単純な形に戻せる。gate の分岐が減り、cron が何を保証するかが読みやすくなる。
3. smoke を独立 step で持つと「rate 計測 test が Skip され続けているのに緑」の見逃しが起きる。rate 計測 test の冒頭で key と binary を確認し、cron gate に載る 1 回版は precondition 不足を `t.Fatalf`（Skip ではなく FAIL）にすれば、疎通異常はそこで検出できる。Cursor だけ smoke がある非対称も消える。

## 3. Rejected

1. Cursor / Gemini だけ本番 key を継続する案 — shim が明示的に禁止した。TTS の有料 key 化で「TEST key だと到達しない」前提も崩れている。
2. `SYSTEM_TEST_RUN` を残し、rate 計測も同 workflow の dispatch で兼ねる案 — cron と dispatch で走る test 集合が変わり、cron gate が「全部」を保証しなくなる。rate 計測は繰り返し回数・調整パラメータ・閾値という別 input 群を持つので、workflow を分けた方が責務が明確。
3. smoke を Go test（`TestCursorCLISmoke`）として独立させる案 — 疎通は決定的で一度通れば回帰しにくい。専用 test を増やすより rate 計測 test の precondition に統合する方が保守点が少ない。Drive / GetX / Gemini も専用 smoke を持たない現状と揃う。
