## 1. Summary

このIssueでは、`GET /episodes`（listEpisode）の各 episode に topic title の配列を載せ、server と HTTP 契約を同じ到達で揃える。完了後は list 応答だけで各 episode の topic 題名を観測できる。web の list-item 表示は扱わない。

## 2. Context

1. 現状の `ListEpisodesResponse` の episode 要素は `episodeId` / `date` / `title` / `durationSec` のみ（`apps/playback/contracts/http.ts`）。
2. topic の全文（`title` / `preface` / `detail` / `startSec`）は `GetEpisodeResponse.body.topics` に既にある。
3. list 画面で topic 題名を使う需要があるが、表示（Feature / CSS）はこのIssueの外。
4. 形の仮定（採用）: `topics` は `{ title: string }[]`。GetEpisode の topic から title だけを射影する。単なる `string[]` や `topic_titles` 改名は採らない（既存 topic object との形の連続を優先）。

## 3. Canonical Sources

1. `apps/playback/contracts/http.ts` — list / get の HTTP schema の正。
2. `apps/playback/worker` の listEpisode 経路（controller / use case / repository Port）— server 到達の正。
3. `testing-strategy` — Scope と契約境界の検証方針。
4. `docs/decisions/2026-08-28T19-20-00`（concept / setting / motif）と `2026-08-28T19-20-01`（list 視覚言語）— UI 表示判断の正。本Issueは契約のみ。

## 4. Scope

### In Scope

1. `ListEpisodesResponse` の各 episode へ `topics: { title: string }[]` を追加する契約変更。
2. worker の listEpisode 経路が、その契約どおり topic title を返す。
3. fake / in-memory 等、list 応答を組み立てる既存供給経路の契約追従。
4. 契約・server 側の自動確認の更新。

### Out of Scope

1. `episode-list-item` および web Feature が topic title をどう表示するか。当該 TSX / CSS は変更しない。
2. GetEpisode の topic 字段追加・削除（既存 `body.topics` 全文契約の変更）。
3. Generator の原稿生成仕様の変更（list が読めるデータが既にある前提。不足が分かったときだけ別Issue）。
4. list 以外の UI・deploy・Access。

## 5. Contract

1. `GET /episodes` 成功時の各 episode は、既存字段に加えて `topics` を持つ。
2. `topics` は 0 件以上の配列。各要素は `title`（空でない string）だけを持つ object。
3. 各 `topics[].title` は、同一 episode の GetEpisode における `body.topics[].title` と同じ題名を、同じ順序で表す。
4. 既存字段（`episodeId` / `date` / `title` / `durationSec`）の意味は変えない。
5. 契約外の字段を list episode 要素へ足さない（strict object を維持）。

## 6. Constraints

1. web の list-item 表示コードをこのIssueの差分に含めない。
2. GetEpisode の topic 全文 schema を list 用に複製して別定義を増やしすぎない。title 射影で足りる。
3. secret・credential を test / log / Error / docs へ書かない。

## 7. Acceptance Criteria

1. [ ] AC-1: `ListEpisodesResponseSchema` が各 episode の `topics: { title: string }[]` を要求し、欠落・不正形を reject する。
2. [ ] AC-2: listEpisode の server 経路が、契約を満たす `topics` 付き JSON を 200 で返す。
3. [ ] AC-3: 同一 episode について、list の `topics[].title` 列が GetEpisode の `body.topics[].title` 列と一致する（順序含む）。
4. [ ] AC-4: fake / in-memory 等の list 供給が新契約で通る。
5. [ ] AC-5: playback の既存 unit / sociable 確認が新契約込みで pass する。
6. [ ] AC-6: `episode-list-item` の表示実装に差分がない。

## 8. Verification

```bash
# contracts / worker の list 関連 test（repo の既存 playback 用 script に従う）
./scripts/playback/test-unit.sh
```

1. schema の欠落・余剰字段ケースが fail することを確認する。
2. list 応答 JSON に `topics[].title` が並ぶことを確認する。
3. `git diff` で web Feature の list-item 表示ファイルに変更がないことを確認する。

## 9. Dependencies



## 10. Risks

1. Drive 上の原稿から list 用に title だけ読む経路が未整備だと、repository 実装が Get 相当の全読込に寄る — mitigation: 既存 Get の topic 配列から title 射影できるならそれを使い、追加 I/O 方針は実装時に最小を選ぶ。
2. web がまだ `topics` を知らない — mitigation: Out of Scope。client 型更新が必要なら表示Issueで扱う。

## 11. Notes

1. 採らない案: `string[]` や `topic_titles: string[]` — 短いが GetEpisode の topic object と形が切れ、後で preface 等を list に足すときに改名が要る。
2. 表示 follow-up は別 task / Issue。本 file は server+API 契約の達成だけを指す。
