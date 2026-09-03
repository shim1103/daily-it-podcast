## Generator System PASS 率台帳

参照: docs/decisions/2026-09-03T14-45-00 / 14-46-00 / 14-47-00 / **16-30-00**（feature/generator-system-e2e-produce-episode）

cron の 1 回通し（`generator-system.yml`）は「壊れていないか」を測る。壊れてはいないが安定しない挙動を追う**事後調査**が rate 計測 test（dispatch 専用）。定常で率を測らない（Decision `2026-09-03T16-30-00`）。

一次 evidence は GHA run URL（`go test -v` の `t.Logf` と `$GITHUB_STEP_SUMMARY`）。この台帳は run 番号を引いた要約。更新は `/log-session` 実行時に session 内の run を拾って行う。

---

### generator-system（cron 週次 + dispatch・`-tags=system`・1 回通し）

`speech_synthesis` / `cursorcli_draft`(1 回版) をまとめて 1 回ずつ実行。1 回でも FAIL なら run が赤。
赤になったら下の rate 計測 test を手動 dispatch して原因を切り分ける。

| date | run | 結果 | 内訳 | 備考 |
|---|---|---|---|---|
| 2026-09-02 | 33627209650 | （移行前・参考） | — | TEST key 化前。CursorCLI 単体のみ dispatch、3/3 |

累計: —（TEST key 化・full 系削除後の初 run から計上）

---

### TestGeminiTTSRate（`generator-tts-rate.yml`・dispatch のみ・`system && ratemeasure`）

**いつ**: cron の 1 回通しで `speech_synthesis` が落ちた／不安定なとき。
先頭 1 束を `runs` 回 `Synthesize`。非空 WAV かつ尺 > 0 で PASS。env は `TEST_GEMINI_API_KEY` 直読み。

| date | run | pass/runs | 平均所要 | callGap | retryBackoffBase | retryBackoffMax | pass_threshold | 備考 |
|---|---|---|---|---|---|---|---|---|
| | | | | 20s | 60s | 3m | 0.8 | 既定値。初回はここから |

累計: —

観測メモ:
- 429 が出た回の `runs` 内位置・待機の伸び。
- callGap / backoff を上げたときの平均所要の変化。

---

### TestCursorCLIDraftRate（`generator-draft-rate.yml`・dispatch のみ・`system && ratemeasure`）

**いつ**: cron の 1 回通しで `cursorcli_draft` が落ちた／尺下限割れが再発したとき。prompt variant の A/B。
固定擬似ソース → `ComposeBriefWithTemplate(items, variant)` → `Write` → `ManuscriptDraftFromWriterOutput`。
Op=`run` の error はその回を分母から除外（環境経由）。正常応答のうち valid が PASS。env は `TEST_CURSOR_API_KEY` 直読み。

| date | run | pass/(pass+fail) | 環境skip | 各回文字数 | prompt_variant | pass_threshold | 備考 |
|---|---|---|---|---|---|---|---|
| | | | | | default | 0.8 | 現行 constants.TextWriterBriefPrompt |

累計: —

観測メモ:
- FAIL 回の原因内訳（全体尺下限割れ / topic 件数 / 各 field range / JSON 崩れ）。
- variant を変えたときの valid 率・文字数分布の変化。下限マージン（total − `DraftTotalCharsMin`）も記録する。

---

### 運用ルール（案・確定は別 Decision）

1. cron の 1 回通しが 2 週連続で同じ test を落としたら bug 扱いで Issue 化。1 週だけの赤は provider 起因として再 dispatch。
2. rate 計測 workflow は「1 回通しが落ちた後の切り分け」または「改善パラメータ / variant を試すとき」に dispatch し、この台帳へ 1 行追記。
3. PASS 率目標は「無料枠時代」と「課金枠時代」で別。無料枠は TTS 系「dispatch で回せたとき緑」で可、cron 定時緑化は課金枠移行後。
