---
name: TextWriter 戻り string の wire 形は JSON とし正本は entities/models.WriterOutput とする
date: 2026-08-29T15:00:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. `port.TextWriter.Write` の成功戻りは **Port 上 string のまま**とする。中身の wire 形は **JSON 1 オブジェクト**とし、brief で Cursor に JSON のみ返させる（markdown code fence で括ってよい。strip は Builder 側）。
2. JSON field の正本は **`entities/models.WriterOutput`** とする。`json.Unmarshal` はこの型へ行い、Domain validation 後に **`ManuscriptDraft` へ写す**。
3. 平文ラベル（`タイトル` / `導入` 等）による parse と、それ用の **`entities/constants` ラベル定数**は採用しない。
4. 文字数 range・句点・topic 数・全体文字数など **Domain Rule** の正本は引き続き **`entities/constants/manuscript_draft_limits.go`** とする（unmarshal 後に検証）。

## 2. Reason

1. Port は vendor 非依存の `string` のままにでき、Cursor CLI の `--output-format json`（transport envelope）と混同しない。中身 JSON は brief 指示で足りる（KISS）。
2. ラベル parse は delimiter 揺れと LLM の表記 drift に弱い。`json.Unmarshal` + schema 型の方が parse 契約が 1 か所（models）に集約される（DRY）。
3. wire 形を `models` に置くと、Domain 型 `ManuscriptDraft` と役割が分かれる。完成 `manuscript.schema.json` とは別物であることが型名で読める。

## 3. Rejected

1. `entities/constants` に field ラベル定数を置き平文 parse する案 — wire 契約が constants と models に分裂する。JSON 指示と二重管理になる。
2. `TextWriter` Port の戻りを `WriterOutput` 型にする案 — Port が business 途中型を知る（`2026-08-25T23-02-35` Rejected と同型）。
3. Cursor Adapter 内で JSON→Draft する案 — parse と Domain validation が Infra に寄る。
