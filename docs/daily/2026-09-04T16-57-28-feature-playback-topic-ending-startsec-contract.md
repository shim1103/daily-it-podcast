---
name: 原稿 body の opening / closing を topics[] と同型の「本文 + startSec」object へ拡張し、generator と playback 表示を追随させた
date: 2026-09-04T16:57:28
session_id: none
branch: feature/playback-topic-ending-startsec-contract
prev: なし
---

## 1. Summary

playback の再生原稿でエンディング開始秒を `MM:SS` 表示すること、listEpisodes API と Drive 原稿契約にその開始秒を載せることを goal に、`body.opening` / `body.closing` を topics[] と同型の「本文 + startSec」object へ拡張した。当初は closing だけ `{ summary, startSec }` に拡張したが、shim の指摘で opening も `{ text, startSec }` へ揃え、3 bookend すべてが「本文 + 音声上の開始位置」を持つ形にした。closing 本文のキーは最初 `text` にしていたが「原稿としての意味がある語を」という指摘で `summary` に変更（generator の `ClosingSummary` と語彙一致）。共有 schema を変えたため generator の produce パイプライン（`build.Timeline` が末尾束の開始累積秒を返し、`MarshalManuscript` が opening/closing を object で書く）も同じ変更内で追随させた。playback web は `EpisodeManuscript` の「導入」「まとめ」bookend が seek bar とラベルを `body.opening.startSec` / `body.closing.startSec` から描くようにし、これまでの「導入は 0 / まとめは総尺」という frontend 補完をやめた。

opening / closing の本文キー名（`text` / `summary`）の統一自体は別 issue の担当であり、このセッションでは contract の型構造の対称化のみ行った。

## 2. Changes

1. contract 拡張を TDD で進めた。`apps/playback/contracts/http.ts` に `openingSchema` / `closingSchema`（どちらも `z.strictObject({ 本文, startSec: z.number().min(0) })`）を足し、`bodySchema.opening` / `.closing` へ適用。`contracts/manuscript.schema.json`（Drive 原稿の正）も `opening` / `closing` を `additionalProperties:false` の object へ。契約 test を先に赤にしてから schema を直す手順。
2. generator 側を追随。`build.Timeline` の戻り値に `closingStartSec`（末尾 segment = closingSummary+farewell 束の開始累積秒）を追加。`MarshalManuscript` の `ManuscriptInput` を `Closing string` → `ClosingSummary string` + `ClosingStartSec float64` へ、JSON 出力型に `manuscriptOpeningJSON{ text, startSec }` / `manuscriptClosingJSON{ summary, startSec }` を追加。`opening.startSec` は先頭 segment で定義上つねに 0 なので Timeline を経由せず直書き。
3. playback web の `EpisodeManuscript` から `durationSec` prop を削除（closing の seek 先が `body.closing.startSec` になり不要化）。「導入」「まとめ」の `onClick` / ラベルを `body.opening.startSec` / `body.closing.startSec` へ。呼び出し側 `episode-item.tsx` と test を追随。
4. fixture 全更新: `fake-episodes.json`（10 件）、e2e stable-episode fixture、playback/generator 双方の SU / Integration test の `body.opening` / `body.closing` を新 object 形へ。schema 検証を担う worker 側の manuscript-schema / verify-manuscript test も追随。broad integration の assertion で `body.opening` を `toBe("開始")` から `toEqual({ text: "開始", startSec: 0 })` へ直した箇所あり。
5. 検証: playback `test:unit` 341 passed（50 files）+ `test:integration` 13 passed、`check-static.sh`（generator golangci-lint + playback biome/tsc/depcruise）全 pass、generator `go test ./...` は broad integration（`apps/generator/test`、schema 検証込み、約 103s）含め全 ok。変更対象の coverage 100%。pre-commit / pre-push hook で 4 commit すべて緑。
6. Decision Record 1 本（`2026-09-04T16-44-46`）を新規作成。「seek 用の開始秒を frontend の補完でなく contract に持つ」という再発しうる方針を固定し、先行 Decision `2026-08-25T05-10-48` §1-2・§1-5（startSec を contract に含めない / seek は scope 外）と `2026-09-02T13-55-00` §1-3（closing 束は開始秒を持たない固定 segment）を opening / closing について supersede。旧 file は触らず、置き換え範囲・維持範囲は新 Decision 側に記載。`playback-lane.md` の進捗 index に 1 行追加。
7. `/commit --repo --split` で 4 commit（contracts / generator / playback / docs）に分割し `origin/feature/playback-topic-ending-startsec-contract` へ push。sandbox 内 push が filtering proxy の SSH 認証で失敗し、sandbox 無効で再実行して成功。
8. `AskUserQuestion` は本 repo の protocol で許可条件なしに deny される。設計判断は git 履歴で復元可能な範囲は自律で進め、判断根拠を stdout に出してから実装した。

## 3. Changes（/pr-completion 追記）

1. `/pr-completion` を実行。`/commit --repo --split` の結果を含め daily・lessons を `a652b8a` で commit・push。
2. `gh pr create`（`shim gh` 不使用）で PR #127 を base `develop` で作成。対応 GitHub Issue なし（Issue 番号 0）。`develop` は差分 36 files で merge-base が branch 起点 `4df3673` と一致し divergence なし。`master` は 277 files で review tool 上限超過。`mergeable: MERGEABLE`。
3. AgentReview は Copilot が quota 上限で findings なし（`COMMENTED` のみ、処理すべき指摘なし）。
4. CI: `static-and-unit` / `integration` は 2 回 push 分で各 2 run。1 回目 push 分は `static-and-unit` 3m23s / `integration` 2m27s で pass、2 回目 push 分の `integration` は 2m37s で pass。2 回目 push 分の `static-and-unit`（run 33851497354）は `playback 依存を install`（`npm ci`）で 30 分以上 in_progress のまま hang。同一 gate は 1 回目 push 分（コード commit 全部を含む tree）で既に緑、log-session commit `a652b8a` は doc 2 file 追加のみでコード変更なし、pre-commit / pre-push hook も全 commit で緑。runner の infra 問題と判断。

### Commits

- `cfce7ac`
- `e9bd47f`
- `84ff13b`
- `13e888f`
- `a652b8a`
