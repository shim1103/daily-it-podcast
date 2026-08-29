## 1. Summary

このIssueでは、playback worker の原稿 JSON の schema/domain 検証を Infrastructure（Driven Adapter）から Application 層（Use Cases）へ移す。真の外部境界（Google Drive REST の I/O）だけを Infrastructure に残し、`EpisodeRepository` Port の signature は変えない。あわせて `getEpisode` の失敗 Error 語彙を1つの粗い Domain Error へ集約し、internal→external Error 写像と `listEpisodes` の除外挙動の contract を更新する。

## 2. Context

1. `apps/playback/worker/src/infrastructure/drive/manuscript-schema.ts`（原稿 JSON の schema 検証）が infrastructure 層に置かれている。schema 適合・stem 一致・不正 JSON の判定は Google Drive という具体 platform に依存しない純粋な判断。
2. `google-drive-episode-repository.ts` の `tryDownloadManuscript` が「Drive から bytes 取得」と「schema・stem 判定」を1メソッドで兼務している。同じ判定は `in-memory-episode-repository.ts` にも重複。
3. `getEpisode` の失敗3分岐（JSON エントリ欠落 / wav 欠落 / schema 不適合・stem 不一致）が全て `EpisodeNotFoundError`（`entities/errors/episode-not-found-error.ts`）に畳まれている。「見つからない」と「見つかったが壊れている」が同一 Error。
4. `EpisodeNotFoundError` は internal Error だが、external の `contracts/external-errors.ts` の `NotFoundError` と同語幹で、名前から層の役割が読めない。
5. `listEpisodes` は不適合 entry を黙って除外し部分一覧を返すが、method の contract documentation にその `@ensure` がない。挙動は `tryDownloadManuscript` の why コメントと test（`schema 不適合 JSON は一覧に出ない` 他）にしか固定されていない。
6. 本Issueの設計判断は `docs/decisions/2026-08-29T13-43-53-playback-manuscript-verification-application-layer.md` に固定済み。Issue本文へ Decision 本文を転記しない。

## 3. Canonical Sources

1. `docs/decisions/2026-08-29T13-43-53-playback-manuscript-verification-application-layer.md` — 検証責務の層、Port signature 不変、Error 語彙、listEpisodes contract 化の判断の正。
2. `architecture` の `ports-adapters.md`（§4 platform 依存の分離、§7 Port 粒度）、`ring-model.md`（§3 層分割）、`error-taxonomy.md`（§1 層別分類、§2 変換境界）。
3. `apps/playback/contracts/http.ts` / `contracts/external-errors.ts` — external 契約と external Error の正。
4. `contracts/manuscript.schema.json`（repo 根） — 原稿 JSON schema の正。本Issueで内容は変えない。
5. `contracts/drive-layout.md` — Drive 上の json/wav ペア規約。
6. `testing-strategy` の `levels.md` / `minimization.md` — Scope×Sociability と重複最小化。
7. `error-handling` — Error クラス設計・防御責務・wrapping の正。

## 4. Scope

### In Scope

1. 原稿 JSON の schema 適合・stem 一致・不正 JSON 判定を Application 層（Use Cases 経路）の責務へ移す。`manuscript-schema.ts` を application 層へ再配置する。
2. Infrastructure（`google-drive-episode-repository.ts` / `in-memory-episode-repository.ts`）を、真の外部境界の I/O（token 取得・files.list・bytes download / in-memory の格納取り出し）と、取得結果を検証層へ渡すまでに限定する。
3. `EpisodeRepository` Port の signature（引数・戻り値型）は保ち、Port の戻り値は検証通過済みの `EpisodeManuscript` / `EpisodeListItem` のままにする。検証を「Infra 取得 → Application 検証 → 返す」で1つの充足単位として結線する。
4. `getEpisode` の失敗を粗い1 Domain Error 語（例 `ManuscriptError`）へ集約し、失敗理由は message で分類する。
5. internal Error →  external Error 写像（`controllers/map-internal-error.ts`）を新しい internal 語彙へ更新する。`EpisodeNotFoundError` を Domain 語彙へ改名する。
6. `listEpisodes` の「不適合 entry を除外して部分一覧を返す。entry 単位の失敗で throw しない」を method の contract documentation（`@ensure`）へ明記する。
7. 上記に伴う sociable unit test の再配置・追従。

### Out of Scope

1. `contracts/manuscript.schema.json` の schema 内容の変更（field 追加・削除）。
2. `apps/playback/contracts/http.ts` の HTTP 契約の変更。external から見た応答（200 の JSON 形、404 相当の code）は不変。
3. `EpisodeRepository` Port を「検証前の生 payload を返す」形へ変える再設計（Decision で Rejected）。
4. Generator の原稿生成仕様。
5. deploy・Access・runtime config。
6. `ManuscriptSchema.safeParse` の戻り値 discriminated union の型名付け（`ManuscriptParseResult`）と `success: false` 意図の doc 化。先行セッションで `manuscript-schema.ts` 内に実施済み。本Issュは層移動に伴い当該型を一緒に運ぶだけ。

## 5. Contract

1. `EpisodeRepository` の `listEpisodes` / `getEpisode` / `getEpisodeAudio` の signature（引数・戻り値型）は変更前と同一。
2. `getEpisode` は、対象 episodeId の原稿が存在しない・wav が無い・schema 不適合・stem 不一致のいずれの場合も、単一の Domain Error 型を throw する。失敗理由は Error の message で区別できる。
3. `listEpisodes` は、個々の entry が schema 不適合・stem 不一致・不正 JSON でも throw せず、適合した entry だけの一覧を返す。Drive HTTP 自体の失敗（token・network・非 2xx）はこの限りでなく Infrastructure Error を throw する。
4. `controllers/map-internal-error.ts` は、新しい internal Domain Error を external の `NotFoundError` 相当へ、Infrastructure Error を external の `UnavailableError` 相当へ写像する。写像後の external Error 種別と HTTP status は変更前と同一。
5. schema 検証を行う module は `apps/playback/contracts`（HTTP schema）を import しない。
6. `GET /episodes` / `GET /episodes/:episodeId` の成功時応答形と、失敗時の external error code は変更前と同一（external から観測不能な変更に留める）。

## 6. Constraints

1. external から観測可能な契約（HTTP 応答 JSON 形、error code、status）を変えない。変更は worker 内部の層配置と internal Error 語彙に限る。
2. Port signature を変えない（Decision §1-2、Rejected §1）。
3. `getEpisode` の失敗を種別ごとの Error クラスへ細分しない（Decision §1-3、Rejected §2）。粗い1語 + message 分類。
4. C を複数Issueへ分割しない（Decision Rejected §4）。本Issュ内の順序（検証移動 → listEpisodes contract 化 → Error 語彙更新）として扱う。
5. 汎用の architecture / error-handling / testing 規則を本Issュへ再定義せず SSOT を参照する。
6. secret / credential を test / log / Error message / docs へ書かない。

## 7. Acceptance Criteria

1. [ ] AC-1: 原稿 JSON の schema 適合・stem 一致判定を行う code が application 層に配置され、`google-drive-episode-repository.ts` / `in-memory-episode-repository.ts` は当該判定ロジックを持たない。
2. [ ] AC-2: schema 検証を行う module が `apps/playback/contracts` を import しない（`lint:layers` / grep で確認できる）。
3. [ ] AC-3: `EpisodeRepository` の3メソッドの signature が変更前と同一（型 diff なし）。
4. [ ] AC-4: `getEpisode` が、存在しない / wav 欠落 / schema 不適合 / stem 不一致 の各ケースで単一の Domain Error 型を throw し、ケースごとに message が異なる。
5. [ ] AC-5: `listEpisodes` が entry 単位の不適合で throw せず部分一覧を返すことが method の `@ensure` に明記され、対応する sociable unit test が存在する。
6. [ ] AC-6: `map-internal-error.ts` 経由の external Error 種別と HTTP status が、全失敗ケースで変更前と同一。
7. [ ] AC-7: `GET /episodes` / `GET /episodes/:episodeId` の成功・失敗応答（JSON 形・error code・status）が変更前と同一。
8. [ ] AC-8: playback の既存 + 新規 unit / sociable 確認が全 pass、`tsc --noEmit` 0 error、`biome lint` clean、`lint:layers` 依存違反なし。

## 8. Verification

```bash
./scripts/playback/test-unit.sh
cd apps/playback && npm run typecheck && npm run lint && npm run lint:layers
```

1. schema 検証 module から `apps/playback/contracts` への import が無いことを grep / `lint:layers` で確認する。
2. `getEpisode` の4失敗ケースの sociable unit test が、同一 Error 型・異なる message を assert することを確認する。
3. `map-internal-error` の sociable unit test で、新 internal Error → external Error の写像が変更前と同じ external 種別・status になることを確認する。
4. route 層の sociable unit test（`app.sociable_unit.test.ts`）で `GET /episodes` 系の応答が不変であることを確認する。
5. `git diff` で `contracts/manuscript.schema.json` と `apps/playback/contracts/http.ts` に差分が無いことを確認する。

## 9. Dependencies

1. なし。`ManuscriptParseResult` 型切り出しは先行セッションで `manuscript-schema.ts` 内に完了済み。本Issュは当該 file の層移動時にその型も一緒に運ぶ。
2. 本Issュ内の実施順: (a) 検証責務を application へ移動 → (b) `listEpisodes` の除外を `@ensure` 化 → (c) `getEpisode` の Error 語彙集約と `map-internal-error` / 改名。(a) が Port 内部結線の土台で (b)(c) はその上。

## 10. Risks

1. Port の外形を保ったまま内部結線だけ組み替える際、検証層を通さない経路が残ると未検証データが controller へ到達しうる — mitigation: Port の戻り値型は検証済み型のまま固定し、Composition Root で結線される経路が必ず検証層を通ることを sociable unit test（repository 経由 use-case）で確認する（Decision §1-2、Rejected §1）。
2. `EpisodeNotFoundError` 改名で import が広範囲に波及する — mitigation: 改名は機械的置換。`entities/errors/` 配下の1クラスと参照箇所（repository・use-case・map-internal-error・各 test）に限定され、external 契約には及ばない。
3. schema 不適合が Domain Error（Use Cases 発生）へ分類変更されることで、Infrastructure Error との境界判断が変わる — mitigation: 「Drive HTTP 自体の失敗＝Infrastructure Error」「取得は成功したが内容が不適合＝Domain Error」の線引きを Decision と `error-taxonomy` §1 に沿って method doc へ明記する。

## 11. Notes

1. Domain Error の型名は `ManuscriptError` を第一候補とする。`DomainError` は粒度が粗すぎるため、原稿という対象を残した語にする。実装者判断。
2. 「Infra が取得して Application が検証して返すまで」を1つの充足単位とする（shim の note）。Port は infra 結果を application で検証して返すところまでを内部で担い、外へは検証済み型だけを出す。
3. 採らない案は Decision の Rejected を参照（Port 戻り値の生 payload 化 / Error 種別クラス細分 / doc だけ厚くする / 3 Issue 分割）。本Issュへ再掲しない。
4. follow-up 候補（scope 外）: `list-episodes` / `get-episode` use-case test の double パターン統一。
