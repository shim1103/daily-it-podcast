---
name: Hono RPC 切替後も path の正本は contracts、RPC は request、API Client は warranty に分ける
date: 2026-08-26T19:27:00
branch: feature/playback-web-view-model-react-hooks
---

## 1. Decision

1. web↔worker の **path 正本**は `apps/playback/contracts` に一本化する。Hono route はそこから template 定数を import して登録する。web 側で path 文字列を再発明しない
2. Hono RPC（`hc<AppType>()`）が担うのは **request 組み立て**（URL / method / path param）だけとする
3. web API Client が担うのは **warranty**（status・payload parse・schema・network 吸収 → `ApiResult`）だけとする。path 文字列も `encodeURIComponent` も持たない
4. HTTP 境界の保証を Hono / RPC に丸投げしない。Outbound の Result 化は `http-boundary` / frontend `api-client` のまま Client に残す
5. repo 根 `contracts/`（Drive・generator 共有）と `apps/playback/contracts/`（playback HTTP）は混ぜない

## 2. Reason

1. 採用決定（`2026-08-26T00-00-00-architecture-reconsider-react-hono.md`）の Hono RPC merit は「手動の path/method 型同期を Hono 機構へ寄せる」ことであり、「runtime の response 保証まで framework に移す」ことではない。保証を RPC に寄せると `http-boundary` の Outbound 3義務（request / response 保証 / network 吸収）と衝突する
2. path を AppType だけに置くと、worker test・`audioRef`・encode 済み concrete URL の oracle が別表現になり、同じ wire 規則が二重になる。contracts に template（`:episodeId`）と concrete helper（encode 済み）を同居させ、Hono が template を、RPC が encode 付き param を、tests が concrete helper を共有すると一本になる
3. Hono RPC は path param を encode しない。encode を Client（warranty）に置くと request 知識が保証層へ漏れる。encode を RPC request 層に閉じ、concrete helper との一致を RPC test で固定する
4. `PlaybackApiClient` の public interface を不変に保つと ViewModel 以降が RPC 実装詳細を知らずに済み、RPC 差し替えと hook 化を並行しやすくなる
5. Drive 原稿の正本を HTTP 契約へ混ぜると generator が playback HTTP を知る経路になる（`2026-08-17T17-40-00` Rejected と同型）。今回の寄せは `apps/playback/contracts` 内に限る

## 3. Rejected

1. HTTP 境界の保証を全部 Hono / RPC に任せる — `http-boundary` の Outbound 責務と矛盾し、壊れた JSON の runtime 落ちを型だけでは代替できない
2. `apps/playback/contracts` の schema / error を捨てて Hono 出力型だけを信じる — runtime 保証と web 語彙の API-error 写像が消える。endpoint 少数でも YAGNI で OpenAPI を再導入するほどではない
3. path 正本を web の RPC 呼び出し形だけにする — worker route・`audioRef`・integration request が別 SSOT になる
4. encode を API Client に残す / Client と RPC の両方が encode する — request と warranty の境界が再び曖昧になる
5. repo 根 `contracts/` と playback HTTP 契約を一本化する — generator と playback HTTP の結合を再導入する
