## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。

- [ ] error 3層表現統一 — `docs/tasks/todo/generator-error-taxonomy-unify.md`
- [x] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）
- [x] Broad Integration（`apps/generator/test/*_broad_integration_test.go`。達成契約 file は完了削除）
- [ ] GHA 本番 produce — workflow 済（`generator-produce-episode.yml`）。Run / Broad 実装済。Secret/Variable 登録は人手
- [x] System — 既定 gate（`-tags=system`）が緑（run 33610705667）。真因は Gemini の応答 parse バグ（Decision `2026-09-02T18-01-00`）。TTS 単体 test + 「Gemini 以外 full」test へ分割、full run は `system && full` へ分離（Decision `2026-09-02T13-57-00` / `16-57-00`）。引き継ぎ `docs/tasks/todo/generator-system-e2e-produce-episode.md`

### D（未決・未実測・文案）

| topic | 概要 |
|---|---|
| Prompt / limits 文案・数値 | 尺モデルは Decision `2026-08-30T03-06-53`。残は実運用後の微調整 |
| 挨拶文案 | Opening/Closing 定数は date placeholder 入り template で確定。実運用での文言微調整のみ残 |
| composite 高度化 | dedup / sort（2 情報源後） |
| 第 2 情報源 Adapter | 別 Issue 化待ち |
| GHA production workflow | YAML・inventory 名は済。Run 実装済。repo へ本番 Secret/Variable を登録する人手作業が残る。定時緑化は Secret 登録後 |
| `generator-system-e2e-produce-episode` | 既定 System gate 緑化済み（parse バグ修正が主因、Decision `2026-09-02T18-01-00`）。残は develop PR と full run（`system && full`）の課金枠移行後確認。詳細は同名 todo |
| Drive ペア書込の補償・staging | 公開型で残骸許容（Decision `2026-08-30T23-32-00`）。補償 delete / staging→rename の再検討は後回し |

### Integration test 方針

```text
gate = secret なし Narrow + Broad（Decision 2026-08-30T11-56-00）
System = gate 外・週次 + dispatch（Decision 2026-08-30T12-49-01）
本番 produce = 毎日 07:00 JST + dispatch（同 Decision）
```
