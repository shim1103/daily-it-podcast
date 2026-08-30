---
name: playback の EpisodeRepository Port は取得したままの生 payload を返し、原稿検証は use-case が行う
date: 2026-08-29T18:20:21+09:00
branch: feature/playback-application-infra-boundary
---

## 1. Decision

1. `EpisodeRepository` Port は、取得元（Google Drive / メモリ）から取れたものを decode したまま返す。schema 適合・stem 一致・不正 JSON・wav 欠落の判定は一切しない。Port の戻り値型は生 payload（`{ json: unknown }` を含む型）とする。
2. 原稿の schema 適合・stem 一致・wav 有無の検証は use-case（`application/use-cases/get-episode` / `list-episodes` / `get-episode-audio`）が行う。use-case が `application/manuscript` の純関数（`verifyManuscript` / `selectValidListItem`）を呼び、検証済みだけを controller へ返す。`EpisodeContentError`（Domain Error）は use-case が throw する。
3. Port は「取得対象が存在しないこと」を戻り値で表現する。`getManuscript` / `getEpisodeAudio` は対象が無い時 `undefined` を返し、throw しない。`listManuscripts` は該当ゼロ件で空配列を返す。use-case が `undefined` を `EpisodeContentError` へ変換する。Infrastructure Error（`DriveError`）は「取得元 HTTP 自体の失敗（token・network・非 2xx・応答形式不正）」だけに限る。
4. これは Decision `2026-08-29T13-43-53`（同名: playback の原稿 schema/domain 検証は Application 層が持ち、EpisodeRepository Port の signature は変えない）の §1-2 と §3-1 を supersede する。§1-1（検証は Application の責務）・§1-3〜§1-5（粗い1 Domain Error / 写像更新 / `EpisodeNotFoundError` 改名 / `listEpisodes` 除外の `@ensure` 化）は維持する。

## 2. Reason

1. 旧 Decision §1-2 は「Port の外形（検証済み型を返す）を保ったまま内部結線だけ組み替える」とした。これを満たそうとすると、Infra が検証層を内部で呼ぶ（Infra が判定ロジックへ依存する）か、内側 Port + wrapper を新設する（同一 signature の Port が2枚並ぶ、`ports-adapters` §7 が戒める形）かのどちらかになる。前者は「取得」と「検証」を Driven Adapter が再び兼務し SRP に反する。後者は Port 増殖。旧 Decision の制約下では、検証を Application へ寄せる目的と Port 外形維持が両立しない。
2. generator は同じ境界を write 方向で既に解いている。`port.EpisodeWriter.Write(bytes)` は生 bytes を受け、`application/WriteEpisode.Run` が `schema.Validate` を自分で実行して検証済みだけを Port へ渡し、`infrastructure/drive/gdrive/writer.go` は渡された bytes を書くだけで検証しない。playback の read 方向はこの鏡像にすると、Port は生 payload を返し、use-case が検証し、Infra は I/O だけを持つ。同一 codebase 内で read と write の境界の切り方が揃い、次の実装者が generator を読めば playback の構造を推測できる（Principle of Least Astonishment）。
3. 「Port の戻り値が検証済みである invariant が controller まで未検証データを届けない保証になる」（旧 §3-1 の Rejected 理由）は、use-case が Port と controller の間に必ず入る構造で代替できる。Composition Root は controller へ use-case しか渡さず、use-case は必ず `verifyManuscript` / `selectValidListItem` を通す。検証を通さない経路が存在しないことは Composition Root の sociable unit test（override 無し経路が repository → use-case → 検証を通る）で固定する。invariant の在り処が「Port の戻り値型」から「use-case が検証を挟む結線」へ移るだけで、controller が未検証データを受け取らない保証は保たれる。
4. 「取得対象の不在」を `DriveError` の throw で表現すると、`map-internal-error` で `DriveError` が external `UnavailableError`（HTTP 503）へ写像されるため、token 失敗などの真の I/O 失敗と同じ status になり「不在 = 404」を型で分離できない。generator の `port.ItemSource`（「該当なしは空 slice、throw しない」）に倣い、不在を戻り値（`undefined` / 空配列）で表現すれば、新しい Infrastructure Error クラスを追加せず、use-case が `undefined` → `EpisodeContentError` → external `NotFoundError`（404）へ写像でき、写像規則も status も変えずに済む。

## 3. Rejected

1. 旧 Decision `2026-08-29T13-43-53` の §1-2 を維持し、Infra が内部で検証層を呼ぶ / 内側 Port + wrapper を新設する案 — §2-1 の通り、SRP 違反か Port 増殖のどちらかになる。実装セッションで内側 Port 版を一度作ったが `ports-adapters` §7（同一 signature の interface を複数並べない）に反すると判断し撤回した。
2. 「取得対象の不在」を Infra が `DriveError`（または新設の Infrastructure Error）で throw し、use-case ではなく `map-internal-error` に分岐を足して 404 へ写像する案 — Error クラスが増え、写像分岐も増える。`defensive-design` §1（型が増えるたび構造が増える）を避け、不在は戻り値で表現する。
3. Port の method 名を旧 `listEpisodes` / `getEpisode` のまま戻り値型だけ生 payload へ変える案 — 戻り値が「一覧の完成品」から「検証前の生 entry」へ変わるのに名前が変わらないと、呼び出し側が「`getEpisode` は episode を返す」と誤読する。`getManuscript`（原稿を取得したまま返す）/ `listManuscripts` へ改名し、返すものが生 payload であることを名前に出す。
