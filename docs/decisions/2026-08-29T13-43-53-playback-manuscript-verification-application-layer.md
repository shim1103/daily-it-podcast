---
name: playback の原稿 schema/domain 検証は Application 層が持ち、EpisodeRepository Port の signature は変えない
date: 2026-08-29T13:43:53+09:00
branch: feature/playback-list-episodes-topics-titles
---

## 1. Decision

1. Google Drive 上の原稿 JSON の schema 適合・stem 一致・不正 JSON の判定（`ManuscriptSchema` 相当）は、Driven Adapter ではなく Application 層（Use Cases）の責務とする。真の外部境界（Google Drive REST の token 取得・files.list・bytes download）だけを Infrastructure に残す。
2. `EpisodeRepository` Port の signature（`listEpisodes` / `getEpisode` / `getEpisodeAudio` の引数・戻り値型）は変えない。検証責務の移動は Port 契約の再設計ではなく、Port を満たす経路の内部で「Infra が取得した結果を Application が検証して返すまで」を1つの充足単位として組み替えることで行う。Port が公開する戻り値は、検証を通過済みの `EpisodeManuscript` / `EpisodeListItem` のままにする。
3. `getEpisode` の失敗を種別ごとの Error クラスへ細分しない。`ManuscriptError`（または同等の粗い Domain Error 語）を1つ持ち、失敗理由（JSON エントリ欠落 / wav 欠落 / schema 不適合 / stem 不一致）は message で分類する。
4. internal Error（現 `EpisodeNotFoundError`）から external Error（`contracts/external-errors.ts` の `NotFoundError` 等）への写像（`map-internal-error.ts`）を、3 の新しい internal Error 語彙へ更新する。`EpisodeNotFoundError` の名は external の `NotFoundError` と紛らわしいため、internal 側を Domain 語彙へ寄せる。
5. `listEpisodes` が不適合 entry を黙って除外し部分一覧を返す挙動は、method の contract documentation（`@ensure`）に明記する。この「除外」は error-taxonomy §2-2「他層 error を握りつぶさない」に反する隠蔽ではなく、`listEpisodes` 自身の仕様であることを doc で区別する。

## 2. Reason

1. `ports-adapters` §4 は「platform 非依存の入力・手順・判断を application core 側へ置く」「OS command・filesystem semantics・runtime API などの具体依存を Adapter へ閉じ込める」と定める。原稿 JSON の schema 判定は Google Drive という具体 platform に依存しない純粋な判断であり、Adapter に置くと「取得」と「検証」という変更理由の異なる 2 責務が Driven Adapter に同居する（SRP 違反）。同じ判定が `InMemoryEpisodeRepository` にも重複しており、検証を Application へ引き上げれば adapter 実装ごとの重複が消える（`ports-adapters` §7）。
2. Port signature を変えないのは KISS / YAGNI。検証の所在を直すのに Port 契約まで作り直すと、`list-episodes` / `get-episode` use-case、両 repository、controller、全 sociable test が同時に動く大きな差分になり、Fault Isolation が崩れる。Port の外形を保ったまま内部結線だけを組み替えれば、observable な契約（controller が受け取る型、HTTP 応答）は不変のまま検証層だけが移る。「Infra が取得して Application が検証して返すまで」を1単位にすることで、Port の戻り値が「検証済み」である invariant は維持される。
3. `EpisodeIncompleteError` / `ManuscriptInvalidError` のような種別クラスを増やすと、`getEpisode` の失敗軸（存在しない / 壊れている / 不整合）の数だけクラスが増え、`map-internal-error.ts` の分岐もその数だけ増える。external 契約（HTTP 404 `episode_not_found` 相当）は失敗理由を区別しないため、internal で細分しても外部には現れない。粗い 1 語 + message 分類なら、運用時の切り分け情報は message に残り、写像は 1 対 1 のまま保てる（`error-taxonomy` §1、KISS）。
4. `EpisodeNotFoundError` は `entities/errors/` 配下の internal Error だが、名前が external の `NotFoundError` と同語幹で、層の役割が名前から読めない（`Principle of Least Astonishment`）。検証が Application へ移ると schema 不適合は Domain Error（Use Cases で発生するビジネスルール違反）に分類が変わるため、この機会に internal 側の語彙を Domain へ寄せ、写像も更新する。
5. `listEpisodes` の除外挙動は現状 `tryDownloadManuscript` の why コメントにしか書かれておらず、method の `@ensure` は Get 経路の Error 変換しか語っていない。test（`schema 不適合 JSON は一覧に出ない` 他）は挙動を固定しているが、contract 記述が test に追いついておらず、次の実装者が「一覧が部分的なのはバグか仕様か」を判断できない（`logging` §0：context を持たない agent が再現できる状態）。

## 3. Rejected

1. `EpisodeRepository` Port の戻り値を「検証前の生 payload（unknown / bytes）」へ変え、検証を完全に use-case 側へ出す案 — Port の戻り値から「検証済み」invariant が外れ、controller まで未検証データが到達しうる。`ports-adapters` §4「呼び出し側が validation を迂回できない戻り値を公開する」に反する。Port の外形は保つ。
2. `getEpisode` の失敗を `EpisodeIncompleteError` / `ManuscriptInvalidError` 等の種別クラスへ分ける案 — 失敗軸の数だけクラスと写像分岐が増える。external 契約が理由を区別しないため割に合わない。message 分類で足りる。
3. 検証を Infrastructure に残したまま `@ensure` の記述だけ厚くする案 — SRP 違反（取得と検証の同居）と adapter 実装間の判定重複がそのまま残る。doc を足しても構造の問題は解けない（`ring-model` §3「誤った層区分を残したまま内部へ subdir を足しても診断の一部しか解けない」）。
4. C（Application/Infra 境界）を C-1/C-2/C-3 の 3 Issue へ分割する案 — 3 つとも同じ層再配置の別面で、同じ file 群を触る。分けると片方だけ merge された中間状態で層が半分ずれた状態が残る。1 Issue の中の順序（検証移動 → listEpisodes contract 化 → Error 語彙更新）として扱う。
