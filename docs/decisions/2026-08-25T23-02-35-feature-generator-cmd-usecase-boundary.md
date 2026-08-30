---
name: TextWriter は string を返し、ManuscriptDraft 化は ProduceEpisode の Domain 手順とする
date: 2026-08-25T23:02:35
branch: feature/generator-cmd-usecase-boundary
---

## 1. Decision

1. `port.TextWriter.Write` の成功戻りは **非空 string**（text 断片）とする。`ManuscriptDraft` も完成 `manuscript.schema.json` も返さない。
2. string → `ManuscriptDraft` の解釈は **`ProduceEpisode`（Builder）** が行う。失敗は **Domain Error**（`InvalidManuscriptDraft`）。Infra Error にしない。
3. 専用の TextWriter UseCase は新設しない。変換は Builder 方針の一部とする。
4. `SpeechSynthesizer` は 1 text → 1 WAV のまま（変更なし）。
5. 本 Decision は `2026-08-25T22-37-29` のうち「TextWriter が ManuscriptDraft を返す」結論を上書きする。同 file の SpeechSynthesizer 薄さ・定型/TTS 順が ProduceEpisode である結論は維持する。

## 2. Reason

1. Adapter が知るのは vendor から取り出した text である。Intro/Topics/ClosingSummary への分解は episode 生成の形であり、Cursor が変わっても同じ方針になりうる。Port に Draft を置くと Infra が business 途中型を知り、parse 失敗が Infra Error に見えやすい。
2. Domain Error にすると、CLI 障害（Infra）と「文はあるが方針上の Draft にならない」（Domain）を観測上分けられる。
3. 変換だけの UseCase を足すと、今は変わり方が Builder と同じ塊なのに単位だけ増える。独立ゲートの実測が無い（YAGNI）。

## 3. Rejected

1. `TextWriter` が `ManuscriptDraft` を返す案（`2026-08-25T22-37-29` の当該結論）— business 変換が Adapter に寄る。
2. string→Draft を Cursor Adapter 内で行い Infra Error にする案 — 層と Error 分類が崩れる。
3. `WriteEpisode` が string→Draft する案 — Gate が構築方針を持つ。
4. 変換専用 TextWriter UseCase を今新設する案 — 対称化のための分割。
