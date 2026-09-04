---
name: ManuscriptDraft wire 前処理は先頭 prose 除去を含む
date: 2026-08-30T17:50:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `ManuscriptDraftFromWriterOutput` の wire 前処理は、code fence strip に加えて、先頭の非 JSON 散文を落とし最初の `{` から対応する `}` までを切る。
2. `{` が無い・brace が閉じない場合は従来どおり Unmarshal 失敗を `invalid_manuscript_draft` にする。

## 2. Reason

1. System 実測（run 33302350069）で Cursor envelope decode は成功したが、`result` 先頭が日本語散文のため `invalid character 'ã' looking for beginning of value` となった。
2. Prompt は「JSON オブジェクト 1 つのみ」と書いていても、モデルが前置きを付けることがある。前処理で落とす方が System の再現性を上げる。

## 3. Rejected

1. Prompt 強化だけで済ます案 — 既に「JSON のみ」と書いて失敗した。前処理の方が決定的。
2. vendor envelope を Application が知る案 — Infrastructure 境界を壊す。
