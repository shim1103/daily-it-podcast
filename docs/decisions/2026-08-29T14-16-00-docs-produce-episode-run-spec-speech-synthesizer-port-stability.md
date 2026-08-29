---
name: SpeechSynthesizer Port は Synthesize(text) を固定し vendor option 追随で拡張しない
date: 2026-08-29T14:16:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. `port.SpeechSynthesizer` は **`Synthesize(ctx, text string)` のみ**を公開する。voice / 口調 / speed / pause / language 等の **Port 引数は追加しない**。
2. vendor が voice 等の option を持つ場合、**Gemini Adapter 定数**で選び、Application からは見えない。option が無い・未知の場合は Adapter 内で握りつぶすか既定値を使う。
3. vendor 追加や option 追加を理由に **Port を頻繁に拡張しない**。Application が必要とする能力が変わる rare な場合のみ、Port 変更を **独立した B** で検討する。

## 2. Reason

1. Port の目的は Application を vendor から隔離することである。vendor 追随のたび Port を触ると、隔離が形骸化し、全 Adapter と全 caller が連鎖変更になる（Orthogonality 崩れ）。
2. 今の caller は口調を選ばない。option を Port に載せるのは YAGNI。Infra 定数へ閉じれば Gemini 変更は Adapter 単位で済む（`2026-08-17T17-41-59`）。
3. pause 要件は Port 拡張ではなく segment 分割 + `concatWAV` で満たす（別 Decision）。Port 拡張と mechanism 拡張を混同しない。

## 3. Rejected

1. vendor が option を持ったら実測後に Port を拡張する案 — 拡張が vendor 追随の既定手段になる。Port の意味が「安定境界」ではなく「vendor 設定バッグ」になる。
2. Application から voice を選べるように Port を厚くする案 — 生成方針が Application に残るべき領域と、vendor mechanism が混ざる。
3. TTS ごとに別 Port（`SlowSpeechSynthesizer` 等）を増やす案 — 対称化のための分割。現要件なし（YAGNI）。
