---
name: Gemini TTS の rate 計測は調整パラメータ注入つき専用 test にし、先頭 1 束を N 回だけ回す
date: 2026-09-03T14:46:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Gemini TTS の PASS 率と所要時間を測る専用 test（`//go:build system && ratemeasure`）と専用 workflow（`generator-tts-rate.yml`、`workflow_dispatch` のみ、cron なし）を新設する。
2. この test は `gemini.SpeechSynthesizer` の `callGap` / `retryBackoffBase` / `retryBackoffMax` を **注入で差し替え可能**にした上で、workflow の input 値を渡して回す。差し替えのために `NewSpeechSynthesizerWithTuning` を追加し、現行の `const` 値は既定値として残す（`NewSpeechSynthesizer` は既定 tuning へ委譲）。`MaxAttempts` は注入対象に含めない。
3. 1 回の計測は、`build.SpeechTexts` が返す束のうち **先頭 1 束だけ**を `Synthesize` し、非空 WAV かつ尺 > 0 を PASS とする。これを N 回（input `runs`、既定 10）直列で繰り返し、`pass/runs` が閾値（input `pass_threshold`、既定 0.8）以上なら緑、下回れば `t.Fatalf`。
4. 各回の所要秒・PASS 率・使用した tuning 値を `t.Logf` へ出し、workflow が `$GITHUB_STEP_SUMMARY` へ pass/total と平均所要を書く。

## 2. Reason

1. 「1 回の Gemini TTS が正しい形式で返り、無料枠内に収まること」は code assert の責務（retry 分岐・budget・応答 parse は sociable unit で固定済み／固定する）。1 回 PASS すれば retry 系の保証は従属する。この workflow に残る役割は「実 API 相手にその 1 回がどれくらいの率と時間で成功するか」を単純計測することだけ。よって全束を回す必要はなく、先頭 1 束で足りる。TEST key の req 消費も最小化できる。
2. `callGap` / backoff は 429 頻度と総所要のトレードオフを決めるパラメータで、最適値は実測でしか分からない。値を変えた run と変えない run を同じ台帳に並べられるよう、注入口を設けて workflow input から渡す。`const` を書き換える運用だと変更のたびに commit が要り、比較しづらい。
3. cron に載せない理由は、繰り返し実行が RPM を圧迫し他の System test と干渉するため。dispatch 専用にして、回したい時に回す。
4. `MaxAttempts` を注入対象にしないのは、retry 上限は「無料枠を焼かない」ための code 側の保証であって、計測で動かす変数ではないから（先行 Decision `2026-09-02T13-56-00`）。

## 3. Rejected

1. `speech_synthesis_system_test.go`（既存の `system` tag）に rate 計測を足す案 — cron gate で毎回 N 回走ることになり RPM を焼く。tag と workflow を分けて dispatch 専用にする。
2. topic+2 束フルを N 回回す案 — 1 episode 相当の連続呼び出しを N セット行い req 消費が `(topics+2)×N` になる。「1 回 PASS すれば retry/quota は code 保証」という前提の下では、連続時の 429 傾向まで見る必要がなく、req を無駄に消費するだけ。
3. `callGap` 等を `const` のまま、値を変えたいときに都度書き換える案 — 変更ごとに commit が要り、同一条件での再現・比較がやりにくい。注入口を一度作れば以後は input だけで動かせる。
4. 閾値未達でも緑にして数値だけ記録する案 — dispatch 専用なので cron を汚さない。閾値で `t.Fatalf` にすれば run 結果の赤/緑で改善要否が一目で分かる。
