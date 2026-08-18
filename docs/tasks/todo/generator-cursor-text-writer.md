## 1. Summary

このIssueでは、generator の **Cursor** 呼び出しを driving する `TextWriter` Driven Adapter を実装する。完了後、`TextWriter` Port の Unit（exec Stub）で Cursor envelope 解釈が通り、成功時は非空の text 断片だけを返す。

## 2. Context

- 原稿は `contracts/manuscript.schema.json` の完成系ではなく、Cursor 呼び出しの **断片**として構築する（その中間表現の責務は Adapter に閉じる）。
- Port/定数の正は `apps/generator/internal/application/port/text_writer.go` と `apps/generator/internal/infrastructure/manuscript/cursorcli/constants.go`、および `docs/decisions/2026-08-18T16-30-00-feature-cursor-text-writer.md`。
- Cursor 呼び出しは non-interactive（`agent -p`）で行う。`manuscript.schema.json` への適合は UseCase 側で行う（このIssueは Cursor 断片の生成だけ）。
- HTTP/Drive/TTS は対象外。

## 3. Canonical Sources

- `apps/generator/internal/application/port/text_writer.go` — `TextWriter.Write` の境界
- `apps/generator/internal/infrastructure/manuscript/cursorcli/constants.go` — Cursor CLI argv 決定値
- `docs/decisions/2026-08-18T16-30-00-feature-cursor-text-writer.md` — Cursor envelope 解釈と呼び出し規則
- `DESIGN.md` — 原稿と Cursor/Drive の責務境界
- `testing-strategy` skill — sociable unit の分類と Fault Isolation

## 4. Scope

### In Scope

- `TextWriter` Port を満たす Cursor CLI Driven Adapter（`Infrastructure`）
- `TextWriter` Adapter の exec 実行は stub 化可能な形（同一プロセスの Unit で検証できること）
- Cursor 成功時の stdout JSON envelope から `result`（assistant 最終テキスト）を抽出して、非空 text 断片として返す
- Cursor 失敗時（非0 exit / stdout 不正 / schema 不一致）は Infrastructure Error として返す
- Adapter 定数の利用（argv / flags を `constants.go` へ閉じる）
- Composition Root が Adapter を Port 実装として結線できること

### Out of Scope

- `TextWriter` を N 回呼ぶ UseCase（呼び出し回数/brief設計/断片の連結/決定稿混在）
- `contracts/manuscript.schema.json` への適合検証
- Gemini TTS / WAV / Drive 書込
- cmd / GHA / 実 `agent` Integration（credential・非決定のため）

## 5. Contract

### Port（`TextWriter`）

| 操作 | 入力 | 成功 | 失敗 |
|---|---|---|---|
| Write | brief（trim後非空） | 非空の text 断片 | Cursor envelope 解釈失敗 / 非0 exit → Infrastructure Error |

### Cursor envelope（adapter 内のみ）

- argv は決定的に固定される
- adapter は stdout の JSON envelope を解釈し、`result` を返す

## 6. Constraints

- vendor 依存（Cursor envelope/argv）を Port へ漏らさない
- secrets の値を code に載せない（`CURSOR_API_KEY` は実行環境から supply。Go は env の存在を期待する）
- `--force` / `auto` を使わない

## 7. Acceptance Criteria

- [ ] AC-1: exec Stub が valid stdout JSON を返すとき `Write` は非空 text 断片を返す
- [ ] AC-2: stdout が JSON でない / envelope の `result` が欠落したとき `Write` は失敗する（Infrastructure Error）
- [ ] AC-3: exec Stub が非0 exit を返すとき `Write` は失敗する（Infrastructure Error）
- [ ] AC-4: argv が `constants.go` に一致し、`--force` / `auto` / fast model を含まない
- [ ] AC-5: Composition Root が Port 結線を構築できる

## 8. Verification

- `go test ./internal/application/... ./internal/infrastructure/... ./internal/composition/...`（generator）

## 9. Dependencies

- 先行: `TextWriter` Port、Cursor CLI constants、Cursor text-writer decision（この Issue で参照）
- 後続: 断片を N 回構築する UseCase、TTS、Drive 書込

## 10. Risks

- Cursor envelope の field が変わると adapter が壊れる（対策: envelope decode を失敗時に Infrastructure Error へ畳む）

## 11. Notes

- 「断片」と「`manuscript.schema.json` 完成系」は一致しない。UseCase と adapter の責務境界を混ぜない。

