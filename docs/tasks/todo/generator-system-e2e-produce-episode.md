# System e2e 引き継ぎ（途中）

## 日quota memo
2026-08-30: 本番 `GEMINI_API_KEY` で System TTS が連続 **429**（例 run `33314746860`）。backoff 延長だけでは緑にならない。**日次/RPM quota 回復待ち**、別 key、または課金枠が要る。

## 現状
1. Cursor smoke / draft は本番 `CURSOR_API_KEY` で到達済み。`TEST_CURSOR_API_KEY` は値不一致で落ちやすい。
2. Gemini も一時的に本番 `GEMINI_API_KEY` を workflow が使用。`TEST_*` は未整理。
3. 同日完成ペア skip・Drive upsert・公開型 json→wav は Decision + 実装済（`b93cc6c` / `fbfd330`）。
4. Gemini hardening push 済: Retry-After / callGap / 60s base・3m max（`79845dd`）、System timeout 延長（`0a5993c`）。**この版の dispatch は未検証**。
5. Drive 書込までの緑は未達が多い（TTS 前で落ちる）。

## 次
1. quota 回復後に `generator-system.yml` を `workflow_dispatch`。
2. 緑なら feature → develop PR。master への workflow 到達は既存パターンに従う。
3. 本番 secret 依存をやめ、`TEST_*` を実働値へ揃えるか方針を決める。
4. 同日 skip あり: test Drive に今日の完成ペアがあると Fetch 前 skip で exit 0 になりうる（「新規 produce」緑ではない）。
