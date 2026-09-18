---
name: catalog由来のHttpErrorはPlaybackApiErrorCode 7種を性質別に分類し、UIへの見せ方は「全体を覆うか」と「retry導線を出すか」の2軸だけをcatalogに許す
date: 2026-09-18T08:47:51
branch: refactor/playback-frontend-state-logic
---

## 1. Decision

1. `CatalogStatus` の `error` 枝は `PlaybackApiErrorCode`（`episode_not_found` / `validation_error` / `configuration_error` / `unavailable` / `client_error` / `network_error` / `invalid_response` の7種）をそのまま保持する。`useEpisodeCatalog` は `ApiResult` の失敗側から受けた code を握り潰さず、値のまま state へ転記するだけに留める（分類・文言化はしない）。
2. UI向けの文言・retry可否は `playback-state.ts` の宣言的 mapping（`catalogErrorPresentation: Record<PlaybackApiErrorCode, {message, retryable}>`）が一元的に持つ。`Record` の網羅性により、`PlaybackApiErrorCode` へ値が増えた時はこの表の更新を tsc が compile error として強制する。
3. catalog由来の7種は、性質分類（`error-handling/defensive-design.md` §8の4分類）上は「外部依存・恒久的失敗」「外部依存・一時的失敗」「内部bug・想定外」の3群に分かれるが、**表示形態（全体を覆うか）の軸では全て同じ結論に収束する**。catalogはpageの唯一のprimary contentであり、これが失敗すれば他に見せられるcontentが無いため、7種いずれも `PageStatus.kind = "unavailable"`（全体を覆う画面）で表す。分岐が生じるのは retryable（外部依存・一時的失敗のみ true）の1点のみ。
4. `PageStatus.unavailable` は `{ message: string; retryable: boolean }` を持つ。`retryable=true` の時だけ、page はmessage直下に小さいretry buttonを出す（`useEpisodeListPage.retry()` = `catalog.load` を呼ぶ配線）。retryableでない時はbuttonを出さない。
5. 「内部bug・想定外」群（`validation_error` / `configuration_error` / `invalid_response` / `client_error`）は、userへ理由を出し分けず同一の汎用文言（「一覧を表示できません」）へ統一する。userが取れるactionが無い失敗の理由を細分化して見せても意味がない。
6. `client_error` は `listEpisodes` が入力を取らない固定 `GET` のため、worker実装のcode上は契約が定義する `400/404/500/503` 以外を返す経路が無い。中間層（proxy等）由来の想定外4xxへのfallbackとして残す。この根拠は本Decisionのみに置き、`catalogErrorPresentation` 側（`playback-state.ts`）はcode commentへ全文転記せず本Decisionを参照するだけに留める（`coding-style/comments.md` の decision 重複禁止規約）。

## 2. Reason

1. 先行Decision（`2026-09-03T13-40-00-feature-playback-web-view-models.md` §1-1）は「catalog取得失敗はblocking、全画面を占める」と定めたが、失敗理由（error code）を保持せず、`PageStatus` は固定文言1種類のみだった。userはretryすれば直る失敗（network）と、直っても意味が無い失敗（validation不整合）を同じ文言・同じ無力感で受け取っていた。
2. defensive-design.md §8の性質分類（外部依存・恒久的／外部依存・一時的／内部bug・想定外）をそのままUIの表示形態分岐へ適用すると、本来「入力起因・回復可能→inline」「外部依存・一時的→通知」等の使い分けが生まれるはずだが、catalog固有の事情（唯一のprimary content・userが選べる代替表示面が無い）により、この分岐は表示形態には効かない。効くのはretry可否のみ。これはcatalog以外の失敗（playback/selectionのerror。非scope）には一般化できない、catalog固有の結論である。
3. `catalogErrorPresentation` を宣言的mappingとして1箇所に持つのは、`error-handling/defensive-design.md` §5「変換規則をinstanceof分岐として散在させず、宣言的データとして1箇所に置く」の直接適用。`Record<PlaybackApiErrorCode, ...>` にすることで、7種の網羅性がtsc compile時に保証される（同§6の静的網羅性）。
4. 内部bug系4種を同一汎用文言へ統一するのは、defensive-design.md §7「利用者に取れる操作がないなら...自動的に正常状態へ倒す」の派生。理由を出し分けてもuser側のactionは変わらないため、文言の粒度を実装都合（error code）ではなくuser視点（打つ手があるか）で揃える。

## 3. Rejected

1. `CatalogStatus.error` を `boolean` のまま（error codeを保持しない）— 先行の実装状態。失敗理由に応じたretry可否・文言の出し分けができず、userは一律「使えません」しか受け取れない。
2. 7種それぞれに専用の表示形態（全体覆う／通知／inline等）を割り当てる — catalogが唯一のprimary contentである以上、「一部だけ隠して他を見せる」という選択肢が存在しない。defensive-design.md §8の4分類をそのまま表示形態へ機械的に適用すると、実際には表現不可能な分岐を型に持たせることになる。
3. `client_error` をcatalog error表から除外する（到達しないため） — 中間層由来の想定外4xxという実在するfailure modeへのfallbackとして必要。「serverのcode上は起こらない」と「protocol上は起こらない」は別の主張であり、後者は成立しない。
4. retry可否を独立した `useState` flagとして持つ — `error` codeから一意に導出できる値を独立stateへ昇格すると、`error` が変わっても`retryable`だけ古い値が残るillegal stateを表現可能にしてしまう（defensive-design.md §10-1）。`catalogErrorPresentation` からのderived valueに留める。
