## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor Cloud Agents REST 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。

### 済み（要約）

1. `ProduceEpisode.Run` / Broad Integration / error 3 層 / 本番 produce workflow
2. 情報源3 Adapter（HackerNews PR #112 / Lobsters PR #113 / ITmedia PR #114）を composite `ItemSource` へ結線。Broad Integration が3源 double で緑
3. Cursor CLI / `commandlaunch` 廃止の A stub と B Decision（`2026-09-03T17-03-33`）。`manuscript/cursorapi` 結線・probe / CLI install 削除
4. [x] `generator-cursor-http-text-writer` — Cloud Agents REST TextWriter 本実装（SU / Narrow）

### 未完了

1. [ ] error 3層表現統一 — `docs/tasks/todo/generator-error-taxonomy-unify.md`
2. [ ] GHA 本番 produce — workflow 済（`generator-produce-episode.yml`）。Run / Broad 実装済。Secret/Variable 登録は人手
3. [ ] System — `generator-system.yml`。再編中（Decision `2026-09-03T14-45-00` / `14-46-00` / `14-47-00` / `16-30-00`）: 全 credential を `TEST_*` 化、cron 週次 + dispatch は **system test を 1 回ずつ通すだけ**、rate 計測は dispatch 専用 test（`system && ratemeasure`）+ 専用 workflow に分離、Cursor install と集計 shell を script 化。台帳 `docs/tasks/todo/generator-system-pass-rate.md`、引き継ぎ `docs/tasks/todo/generator-system-e2e-produce-episode.md`

### D（未決・未実測・文案）

| topic | 概要 |
|---|---|
| Prompt / limits 文案・数値 | 尺モデルは Decision `2026-08-30T03-06-53`。残は実運用後の微調整 |
| 挨拶文案 | Opening/Closing 定数は date placeholder 入り template で確定。実運用での文言微調整のみ残 |
| composite の source またぎ sort | 3 情報源化（Decision `2026-09-02T14-41-00`）で `OccurredAt` 順の混在が起きる。dedup は `SourceID` が全源で異なるため不要。時系列 sort を Application/Composition のどちらで持つかは別判断（事実: 現状は登録順 concat のみ） |
| 別媒体の報道源追加 | Publickey / InfoQ / はてブ IT 等は各々専用 Adapter を新設（`infrastructure/<媒体>/`。RSS 汎用 Adapter は作らない — Decision `2026-09-02T14-41-01`）。RSS 2.0 parse の重複が三度現れたら共通化を検討（未実測） |
| 議論 comment のスレッド深掘り | HN / Lobsters は 1 階層のみ取得（上限は Adapter stub 定数）。ネストした議論を辿るかは未決 |
| TextWriter の web_fetch 実測 | `links:` の URL を Cloud Agents 経路が実際に fetch できるか未実測。できない場合の補完は TextWriter 経路の内側（Application には置かない）。Decision `2026-09-02T14-41-02` / `2026-09-03T17-03-33` 参照 |
| no-repo 原稿品質・token・job timeout | Cloud Agents no-repo が ask 相当の断片になるか、Pro 日次消費、SSE 待ちが GHA job に収まるかは未実測 |
| 撤去済み X 関連 Decision | `2026-08-15T16-39-20` 他の X vendor 記述は指示対象を失った。supersede 注記の要否は log-session / migrate-lessons 側の判断 |
| GHA production workflow | YAML・inventory 名は済。Run 実装済。repo へ本番 Secret/Variable を登録する人手作業が残る。定時緑化は Secret 登録後 |
| `generator-system-e2e-produce-episode` | 既定 System gate 緑化済み（parse バグ修正が主因、Decision `2026-09-02T18-01-00`）。再編中: TEST key 固定・パターン撤去・rate 計測 workflow 分離（Decision `2026-09-03T14-45-00` 系）。詳細は同名 todo |
| Drive ペア書込の補償・staging | 公開型で残骸許容（Decision `2026-08-30T23-32-00`）。補償 delete / staging→rename の再検討は後回し |

### Integration test 方針

```text
gate = secret なし Narrow + Broad（Decision 2026-08-30T11-56-00）
System = gate 外・週次 + dispatch（Decision 2026-08-30T12-49-01）
本番 produce = 毎日 07:00 JST + dispatch（同 Decision）
```
