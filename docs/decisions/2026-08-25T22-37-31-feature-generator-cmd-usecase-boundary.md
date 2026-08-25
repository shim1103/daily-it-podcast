---
name: WAV の尺算出と結合は Application 非公開 helper に置き Entities 公開にしない
date: 2026-08-25T22:37:31
branch: feature/generator-cmd-usecase-boundary
---

## 1. Decision

1. RIFF/WAV からの尺算出と複数 WAV の結合は、`ProduceEpisode`（Builder）が使う **Application 非公開 helper** に置く。
2. これらを Entities の公開 Domain 操作として扱わない。`durationSec` という値の意味は Domain、RIFF の読み書きは mechanism とする。
3. Drive 配置契約（`contracts/drive-layout.md`）は拡張子とペアリングの正であり、RIFF 解析手続きの所有者ではない。
4. 原則の正は architecture `backend/application.md` §4（フォーマットの機械的分解・結合は Builder の非公開 helper でよい）。

## 2. Reason

1. Entities は外部 I/O が変わっても残る意味を置く。RIFF header 操作は保存形式の mechanism であり、Drive 配置の「`.wav` という名前」とも同一ではない。Entities に上げると、内側 ring が伝送フォーマット詳細を抱え、`contracts/` を Entities が import しない規則とも緊張する。
2. 今の呼び出し手は Builder UseCase だけである。公開 Entities API にすると再利用前に抽象が先走る（YAGNI）。複数 UseCase が同じ helper を要する実測が出てから共有を検討する。
3. A で Entities に置いた stub がある場合、本 Decision を正として Application 非公開へ移す（実装は C）。

## 3. Rejected

1. `WAVDurationSec` / `ConcatWAV` を Entities 公開の Domain concept とする案 — mechanism を Domain に昇格させすぎる。
2. 尺・結合を `SpeechSynthesizer` Adapter に閉じる案 — Port/Adapter が episode の尺方針と結合を知る。
3. Drive 配置契約を RIFF 手続きの SSOT とみなす案 — 配置契約は名前とペアであり、byte 解析の所有者ではない。
