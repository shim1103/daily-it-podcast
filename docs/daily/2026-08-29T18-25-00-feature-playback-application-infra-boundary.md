---
name: playback の EpisodeRepository を生 payload 返却へ変え原稿検証を use-case へ寄せる
date: 2026-08-29T18:25:00
session_id: none
branch: feature/playback-application-infra-boundary
prev: なし
---

## 1. Summary

playback worker の Application/Infra 境界で、原稿 JSON の schema/stem 検証を Driven Adapter から取り除いた。`EpisodeRepository` Port は Google Drive / メモリから取得したままの生 payload を返し、`get-episode` / `list-episodes` / `get-episode-audio` use-case が `application/manuscript` の純関数で検証する。generator の write 方向（`port.EpisodeWriter` は生 bytes を受け `WriteEpisode.Run` が `schema.Validate` する）の read 鏡像に揃えた。`EpisodeNotFoundError` は external `NotFoundError` と紛らわしいため `EpisodeContentError` へ改名。external 契約（HTTP 応答形・error code・status）は不変。

当初 `/issue-manager` で `docs/tasks/todo/playback-application-infra-boundary.md` を実装し完了・削除まで進めたが、その実装は旧 Decision（`2026-08-29T13-43-53`）に沿って「Port は検証済み型を返す・Infra 内で検証層を呼ぶ」形だった。shim の追加指示で「Infra はそのまま返す・use-case が検証する」へ方針転換し、旧 Decision の §1-2 / §3-1 を supersede する新 Decision を起こした。

## 2. Changes

1. `EpisodeRepository` の method を `listManuscripts()` → `RawManuscriptEntry[]` / `getManuscript(id)` → `RawManuscriptDetail | undefined` / `getEpisodeAudio(id)` → `Uint8Array | undefined` へ変更。生 payload（`{ json: unknown }` を含む型）を返し、schema/stem/wav 欠落の判定はしない。
2. 検証は use-case へ移動。`get-episode` が `getManuscript` の戻り値に `verifyManuscript` を適用（`undefined` → `EpisodeContentError('JSON エントリが無い')`、`hasAudio === false` → `EpisodeContentError('wav が無い')`、その後 schema → stem）。`list-episodes` が各 entry に `selectValidListItem` を適用し `undefined` を除外。`get-episode-audio` が `undefined` → `EpisodeContentError`。
3. `manuscript-schema.ts` を `infrastructure/drive/` → `application/manuscript/` へ移動。schema/stem 判定本体を `application/manuscript/verify-manuscript.ts` の純関数（`verifyManuscript` / `selectValidListItem`）へ抽出。
4. 両 Adapter（`GoogleDriveEpisodeRepository` / `InMemoryEpisodeRepository`）から検証を完全除去。Google Drive は `downloadJson`（bytes → best-effort decode）までで止め生値を返す。取得対象の不在は throw せず `undefined` / 空配列で表現（generator `port.ItemSource` の「該当なしは空 slice」に倣う）。`DriveError` は「Drive HTTP 自体の失敗（token・network・非 2xx・応答形式不正）」だけに限定。
5. `EpisodeNotFoundError` → `EpisodeContentError` へ改名。throw 元が Infra から use-case へ移った。`map-internal-error.ts` の kind 名を `domain_not_found` → `domain_content` に、参照を新 Error 名へ。写像規則（`EpisodeContentError` → external `NotFoundError` / 404、`DriveError` → `UnavailableError` / 503）と分岐数は不変。
6. 検証網羅（schema 不適合除外・4 失敗ケース・stem 不一致・複合失敗 precedence）を use-case の sociable unit test へ移した。`verify-manuscript.sociable_unit.test.ts` は純関数の入出力契約、use-case test は「Port 生 payload → 検証 → 結果」の結線、Infra test は「生 payload をそのまま返す / 不在は undefined / HTTP 失敗は DriveError」に責務分離。
7. 複合失敗の precedence は use-case 層（json 在否 → hasAudio → schema → stem）と純関数層（schema → stem）に分かれ、それぞれの層の test で固定。Decision に precedence 記述が無いため実装者裁量で決めて doc + test へ固定。
8. `Composition Root` は `new GoogleDriveEpisodeRepository(...)` / `new InMemoryEpisodeRepository()` の直接生成と use-case 結線を維持。前回セッションで足した「override 無し in-memory mode で repository → use-case → 検証を通る」sociable unit test を新 Port でも維持（空 repo → `listManuscripts()` = `[]` / `getManuscript()` = `undefined` → use-case が `EpisodeContentError` → external `NotFoundError` の結線を通す）。
9. `.dependency-cruiser.mjs` は最終的に HEAD と差分ゼロ。方針転換前の一時期に `worker-infra-ports-only` へ `application/manuscript/` 許可を足したが、Infra が `application/` を一切 import しなくなり不要になったため撤去。
10. 新 Decision `docs/decisions/2026-08-29T18-20-21-playback-episode-repository-returns-raw-payload.md` を起こし、旧 Decision `2026-08-29T13-43-53` に `superseded-by` と維持範囲（§1-1・§1-3〜§1-5 は維持、§1-2・§3-1 が撤回）の注記を付けた。完了した Issue の todo file を削除。
11. flow 中に `/issue-manager` の manager → executor → reviewer で内側 Port `RawManuscriptSource` を新設した中間版を一度作り、`ports-adapters §7`（同一 signature の interface を複数並べない）違反として撤回。純関数抽出へ是正した後、shim 指示でさらに Port 生 payload 化へ移った。
12. 全 commit で pre-commit hook（generator static/unit 91%、playback format/lint/typecheck/lint:layers/unit 241 件）と pre-push hook（generator/playback integration）が pass。
13. `feature/playback-application-infra-boundary` を新規 branch として `origin` へ push。

### Commits

- `ed703f1`
- `d3ebf15`
