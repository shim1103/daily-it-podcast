## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。

- [ ] error 3層表現統一 — `docs/tasks/todo/generator-error-taxonomy-unify.md`
- [x] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）
- [ ] Broad Integration — `docs/tasks/todo/generator-broad-integration-produce-episode.md`
- [ ] GHA 本番 produce — workflow 済（`generator-produce-episode.yml`）。Run 実装済につき緑化確認へ。Secret/Variable 登録は人手
- [ ] System — workflow 済（`generator-system.yml`）。suite 実装と TEST_* 登録は後続（`generator-system-e2e-produce-episode`）

### D（未決・未実測・文案）

| topic | 概要 |
|---|---|
| Prompt / limits 文案・数値 | 尺モデルは Decision `2026-08-30T03-06-53`。残は実運用後の微調整 |
| 挨拶文案 | Opening/Closing 定数は date placeholder 入り template で確定。実運用での文言微調整のみ残 |
| composite 高度化 | dedup / sort（2 情報源後） |
| 第 2 情報源 Adapter | 別 Issue 化待ち |
| GHA production workflow | YAML・inventory 名は済。Run 未完のため定時は赤になりうる。repo へ本番 Secret/Variable を登録する人手作業が残る |
| `generator-system-e2e-produce-episode` | workflow 済。suite 本体・assert・TEST_* 値の登録が未 |

### Integration test 方針

```text
gate = secret なし Narrow + Broad（Decision 2026-08-30T11-56-00）
System = gate 外・週次 + dispatch（Decision 2026-08-30T12-49-01）
本番 produce = 毎日 07:00 JST + dispatch（同 Decision）
```
