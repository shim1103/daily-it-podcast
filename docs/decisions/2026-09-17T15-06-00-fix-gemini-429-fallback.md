---
name: Geminiの429使い切りはTextWriter・TTSともTier非依存でsource枯渇にする
date: 2026-09-17T15:06:00
branch: develop
---

## 1. Decision

1. Gemini TextWriterとTTSは、429の有限retryを使い切ったら`port.ErrSourceExhausted`をwrapする。
2. この分類は`TierFree`と`TierPaid`で変えない。
3. retry途中で成功した場合は従来どおり成功を返す。

## 2. Reason

1. `generator-system` run 35188139273で、primary Geminiが429を使い切った後にCursorへ切り替わらず、Infrastructure Errorで停止した。
2. 429は課金区分ではなく、そのcredential sourceを現在利用できない状態を表す。source配列内の位置もAdapterの責務ではない。
3. final sourceが同じ番兵を返しても、合成layerは次sourceが無ければ最後のerrorとして返すため、終端の診断は失われない。

## 3. Rejected

1. `TierFree`だけ枯渇扱いにする案 — Paidも429なら同じく利用不能であり、課金区分でerrorの意味を変える根拠がない。
2. 429を常に即時fallbackする案 — 一過性のrate limitを既存retryで回復する経路を失う。
