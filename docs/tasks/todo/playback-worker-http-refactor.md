## Playback worker: HTTP 責務分離リファクタ（fetch.ts 複雑性削減）

Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

`apps/playback/worker/src/routes/fetch.ts` の責務が多く、読み手が「Route 抽象化」「外向き status+code 写像」「ログ payload」「音声 byte Response 作法」を同一ファイルで追う必要がある。

この Issue では **実装はしない**。責務分離の方針だけ確定し、次 executor が最小差分で refactor できる形に scope を切る。

## 2. Context

1. `fetch.ts` の目的は「Route Handler（Inbound）」として wiring と外向き Response を返すこと。
2. `/http-boundary` と `/architecture` が要求する「誰が何を知り、何を知らないか」に合わせる。
3. 特に以下が `fetch.ts` に混ざっている:
   1. method/path match（routing）
   2. External Error → `{ status, code }` 写像
   3. Route 境界ログ payload の組み立て（`console.error`）
   4. 音声 byte の `Response` body 作法（`ArrayBuffer` / `SharedArrayBuffer`）

## 3. Canonical Sources

1. `apps/playback/contracts/http.ts`（path と schema / content-type）
2. `apps/playback/worker/src/routes/fetch.ts`（現状の混雑箇所）
3. `/http-boundary`（Route Handler / status 割り当て / throw 禁止 / logging 境界）
4. `/architecture` backend route-handler / controller / composition-root
5. `docs/tasks/todo/playback-lane.md`（既完了の前提）

## 4. Scope

### In Scope

1. `fetch.ts` が混ぜている責務の切り分け方針を決める
2. refactor 後に残すべき「Route Handler としての知識」を確定する
3. `fetch.ts` が呼び出すべき helper / module の粒度だけを決める（ただしコード生成は executor）
4. DRY（宣言的 mapping の一箇所化）を優先する

### Out of Scope

1. 実装（コード編集、テスト追加、commit/push）
2. 契約 schema 変更
3. `apps/playback/web` 側の変更
4. Drive adapter の実装（本 Issue は HTTP 境界の refactor のみ）

## 5. Deliverables（ドキュメントとして確定させること）

1. `fetch.ts` の残すべき責務（抽象化のゴール）
2. routing / error mapping / logging payload / audio response body の各責務の置き場所
3. 依存方向（誰が contracts を知り、誰が知らないか）
4. refactor の最小手順（TDD の Red→Green→Refactor に沿う）

## 6. Acceptance Criteria（後続 executor 向け）

1. `fetch.ts` は method/path match と wiring、Controller 呼び出し、最終 Response 生成だけが見える状態になる
2. External Error → `{ status, code }` は 1 箇所の宣言表に集約される
3. `console.error` の payload 生成は Route 境界に限定され、重複しない
4. 音声 byte の Response body 作法は `fetch.ts` の中で肥大せず、別 helper が吸収する

## 7. Verification

後続 executor が `cd apps/playback && npm run test:unit` を通す。

