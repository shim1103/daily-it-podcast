## Playback worker: HTTP 責務分離リファクタ（fetch.ts 複雑性削減）

## 1. Summary

`apps/playback/worker/src/routes/fetch.ts` の責務が多く、Route 抽象化・外向き status+code 写像・Route境界ログ payload・音声 byte Response 作法を同一ファイルで追う必要がある。責務分離を進め、次 executor が最小差分で refactor できる状態にする。

## 2. Context

事実: 現状の `fetch.ts` は method/path match、External Error → `{ status, code }` の宣言表、`console.error` 用 structured payload 組み立て、音声 byte の `Response` 作法を含む。
事実: `playback-worker-http-refactor.md` に、責務分離方針と accept 条件（後続 executor 向け）が整理済み。
不明/仮定: 本 Issue はコード編集そのものを行わない（次 executor が実装する）。

## 3. Canonical Sources

- `apps/playback/worker/src/routes/fetch.ts` — 現状の混雑箇所
- `apps/playback/worker/src/routes/fetch.test.ts` — boundary の期待（status/code/body/log）
- `apps/playback/contracts/http.ts` — path / schema / content-type
- `http-boundary` skill — Route Handler / logging / throw 禁止の境界
- `architecture` backend route-handler / controller / composition-root — 責務分離の前提
- `testing-strategy` skill — 重複最小化と TDD 適用
- `docs/tasks/todo/playback-worker-http-refactor.md` — この Issue の責務分離方針と Deliverables

## 4. Scope

### In Scope

1. `fetch.ts` が混ぜている責務の切り分け方針を、次 executor が実装可能な粒度に確定する
2. refactor 後に `fetch.ts` に残す「Route Handler としての知識」を定義する
3. `error mapping / logging payload / audio response body` を、重複なく移動する置き場所を確定する
4. 依存方向（誰が contracts を知り、誰が知らないか）を維持する

### Out of Scope

1. コード編集、テスト追加、commit/push、todo 削除
2. `apps/playback/web/` の変更
3. 契約 schema の変更
4. Drive adapter 実装（HTTP 境界の refactor のみ）

## 5. Contract

1. `fetch(request: Request): Promise<Response>` の公開 shape は変更しない（return status/body/content-type の観測可能な契約は維持）
2. `External Error → { status, code }` は 1 箇所の宣言表として保持する（2箇所化は禁止）
3. `console.error` の structured payload 形は重複定義せず、Route境界に限定する
4. 音声成功時の Response は `episodeAudioContentType` を `Content-Type` とし、JSON ではなく raw byte を返す

## 6. Constraints

1. `http-boundary` / backend route-handler / controller / composition-root の責務境界を破らない
2. `fetch.ts` から Controller へ渡す入力は `unknown` のまま維持する
3. logging は Route境界に限定し、Controller / UseCase 側でログ出力しない
4. mapping 漏れ時は従来どおり `500` を返し、失敗 JSON に契約外 `code` を出さない

## 7. Acceptance Criteria

- [ ] `fetch.ts` は method/path match と wiring、Controller 呼び出し、最終 Response 生成だけが読み取れる状態になる
- [ ] External Error → `{ status, code }` の宣言表が 1 箇所に集約される
- [ ] `console.error` payload の構築が Route境界に限定され、重複実装が消える
- [ ] 音声 byte の `Response` 作法が `fetch.ts` から肥大せず helper へ退避している
- [ ] `cd apps/playback && npm run test:unit` が pass する

## 8. Verification

1. `cd apps/playback && npm run test:unit`
2. （必要なら）`cd apps/playback && npm run test:integration`

## 9. Dependencies

- `docs/tasks/todo/playback-worker-http-refactor.md`

## 10. Risks

1. `fetch.ts` の責務移動に伴う import 方向の破壊（contracts / error mapping / logging の参照が散る）risk
2. 観測可能な Response の互換性（status/body/content-type）が崩れる risk

## 11. Notes

この Issue は実装をしない。次 executor が最小差分で refactor できるよう、参照元 (`playback-worker-http-refactor.md`) と観測可能契約（`fetch.test.ts`）を優先して書く。

