---
name: playback web の URL hash 同期は routing library を入れず lib/location-hash + useHashSync で行い、外部ストア購読は useSyncExternalStore へ寄せる
date: 2026-08-27T19:20:30
branch: feature/playback-web-page-jsx-mount
---

## 1. Decision

1. URL の hash とアプリの選択状態（`selectedEpisodeId`）の同期に、React Router / TanStack Router を導入しない。`lib/location-hash.ts`（Driven Adapter）と `view-models/use-hash-sync.ts`（custom hook `useHashSync`）で行う
2. `useHashSync` の契約は dyadic `(selectedId, onHashSelect)`。「hash が変わったら何を select するか」の解釈は page 側（`onHashSelect` ハンドラ）に残し、hook は同期の機構だけを持つ
3. `useHashSync` の外部ストア購読部（`window` の hashchange 購読と現在値の読み取り）は、React 標準の `useSyncExternalStore` へ寄せる。現状の `useEffect` + 手書き `addEventListener` + `lastSyncedHashRef` / `onHashSelectRef` による stale closure 対策は暫定であり、follow-up Issue（`docs/tasks/todo/playback-web-use-sync-external-store.md`）で置き換える
4. state→hash の書き込み（`setLocationHash`）と書き戻し無限ループ抑止は `useEffect` に残す。`useSyncExternalStore` が担うのは外部→React の読み取り方向だけである

## 2. Reason

1. `2026-08-26T00-00-00-architecture-reconsider-react-hono.md` が TanStack Router を「画面遷移が一覧⇔詳細の2画面のみ、型安全 path の恩恵が薄い、対応する機能要件が repo に無い」として Rejected している。React Router は TanStack Router より機能が薄く、同じ論理がそのまま適用される。ネスト route・型安全 path・loader・code splitting のいずれも本 web に要件が無い。`window.location.hash` 1個の同期に汎用 routing engine を持ち込むのは Rule of Least Power（`design-philosophy.md §4-2`）が戒める「不安の解消のための強力な実装への乗り換え」に当たる
2. hash 同期を page（統合点）の `useEffect` 3本に直書きすると、`architecture/frontend/page-route.md §3`「状態管理を置かない」に反する。同期は「hash ↔ selectedEpisodeId の双方向同期」という1つの関心事なので、custom hook 1つへ畳むのが SRP（`design-philosophy.md §3-1 S`）と DRY（§2-2）に沿う。custom hook は ViewModel 層の platform 実装（`architecture/frontend/view-model.md §2-1`「副作用（custom hook）」）であり、置き場は `view-models/`。`lib/` には置けない（`architecture/frontend/external-dependencies.md §3-2`「React hooks を置かない」）
3. その custom hook 内で、外部ストア（`window.location.hash`）の購読を `useSyncExternalStore`（React 18 が「外部の変わりうる値を購読して React state として読む」ために用意した公式手段）ではなく `useEffect` + 手書き listener で実装したのは、Least Astonishment（`design-philosophy.md §5-1`、最優先）に照らして次点である。次に読む者が「なぜ手書き購読か」と驚く。標準へ寄せると、hashchange の取りこぼし対策と、listener を貼り替えず最新の `onHashSelect` を呼ぶための `onHashSelectRef`（stale closure 対策）が hook 標準の保証に載り、`lastSyncedHashRef` の一部も `useSyncExternalStore` が返す最新値との直接比較で代替でき、hook の非自明さが下がる（`design-philosophy.md §2-3 KISS`）
4. ただし `useSyncExternalStore` への移行は本 Issue（`episode-list-page.ts` の JSX 化と `main.ts` の `createRoot` 化）の Scope 外である。`2026-08-26T19-29-00` / `2026-08-26T16-48-00` の Rejected「別 Issue の契約を侵食する」と同型のリスクを避け、follow-up Issue へ分離する。本 Issue の完了時点では `useEffect` 手書き版で Verification を満たしており、挙動は hash 同期6シナリオ + page 8ケースの unit test で担保されている
5. state→hash の書き込みを `useEffect` に残すのは、`useSyncExternalStore` が読み取り方向専用だからである。書き込み側まで hook 標準に載せる仕組みは React に無く、`selectedId` の変化を検知して `setLocationHash` する `useEffect` は移行後も残る。書き戻し無限ループ抑止（自分が書いた hash 変更で発火した hashchange を無視する）も、`useSyncExternalStore` が返す最新 hash と `selectedId` の直接比較へ簡略化はできるが、比較ロジック自体は自前で残る

## 3. Rejected

1. React Router / TanStack Router を採用する — 2画面・hash 1個で対応する機能要件が無い。bundle 増（`react-router` 数十 KB）と route tree・`<Outlet>`・loader の概念負荷を負う。`2026-08-26T00-00-00 §3-4` の却下判断がそのまま当てはまり、要件は変わっていない
2. `useHashSync` を本 Issue 内で即 `useSyncExternalStore` ベースで実装する — 本 Issue の AC は「`episode-list-page.ts` を JSX へ書き換え、`main.ts` を `createRoot` へ」であり、hash 同期の実装方式は Out of Scope。同期ロジックの再設計を本 Issue に混ぜると、契約の変更点と実装経路の変更点が1つの PR で混ざり review 単位が崩れる
3. hash 同期を page の `useEffect` に直書きしたまま維持する — `page-route.md §3` 違反。page（統合点）に `lastSyncedHash` 比較・hashchange 購読・初期化ゲート・stale closure 対策という同期の機微が残り、page の変更理由が「配線」から「状態管理」へ広がる
4. `useHashSync` を triadic `(selectedId, onHashSelect, enabled)` にして「load 完了まで同期を止める」フラグを引数で渡す — `coding-style/function-design.md §1` の dyadic 上限を超え、boolean flag + 設計見直し対象。`selectedId` に `undefined`（未初期化）を渡す2状態表現（`null`=選択なし、`undefined`=同期保留）で代替する
5. `useHashSync` の外部購読を子コンポーネント分割（load 完了後に購読を持つ子を mount）で表現する — `onHashSelect` / `selectedId` を prop で降ろす indirection を1層増やす。関心事1つに対し構造過剰（`design-philosophy.md §2-3 KISS` / §2-4 YAGNI）
