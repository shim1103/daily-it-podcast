---
name: Playbackの薄いcacheはWorkers Cacheとbrowser HTTP cacheをresponse headerで制御する
date: 2026-09-16T19:45:00
branch: feature/playback-audio-cache
---

## 1. Decision

1. 先行Decision `2026-09-13T14-23-30`が未決としていたR2移行後の薄いcacheは、Cloudflare **Workers Cache**とbrowser標準HTTP cacheの二段で実現する。具体値の正本は`apps/playback/wrangler.jsonc`とHTTP response組み立てcodeとする。
2. 音声成功responseはbrowserよりedgeを長くfreshにする。音声は原則不変だが同一keyのupsertは可能なため、無期限cache・`immutable`・deployを跨ぐcache共有は採らない。
3. episode一覧は日次更新の反映を長時間止めず、短いbrowser / edge cacheとedgeのbackground revalidationを使う。差分APIは作らない。
4. `ETag`・`Last-Modified`・`Expires`、Generatorからの通知、自動purgeは初手に持たない。cache対象外のAPI errorは保存不可をresponseで明示する。
5. Worker自身のRange処理はcache無効時のHTTP契約として維持し、本番Workers CacheではCloudflareのfull-response cacheとRange slicingを利用する。

## 2. Reason

1. Workers CacheはWorker実行前にlookupし、hit時はWorkerとR2読取をともに省ける。tiered cacheとrequest collapsingを持ち、標準`Cache-Control`で制御できるため、先行Decisionが避けた低水準のCache APIをapplication codeへ持ち込まずに目的を満たす。
2. browser cacheは端末・browser profile内の再取得を省き、Workers Cacheはbrowser cache miss後のR2読取を省くため、責務が重ならない。edge専用headerを使えば両者のfresh期間も独立に決められる。
3. 音声pathはepisode IDで安定し通常は内容が変わらない一方、R2 writerの契約は同一key upsertを許す。有限TTLとversion別cacheを維持すれば、訂正時のstaleを期限付きにしつつ通常再生をcacheできる。
4. episode一覧は小さく、更新頻度も日次である。短いTTLなら連続reloadによるR2の一覧・原稿読取を吸収しながら新着の遅延を小さくできる。差分同期はclient state、追加・変更・削除の契約、不整合回復を新設するため、現状規模に対して過剰である。
5. browserのfresh期間後もedgeがfreshならR2へ到達せず再取得できる。初手でvalidatorをR2 metadataからHTTP境界まで運ぶ変更は得が小さく、絶対時刻の`Expires`も`max-age`と重複する。更新通知やpurgeもGeneratorとPlaybackの運用結合を増やす。
6. Workers CacheはcoldなRange requestから`Range`を除いてWorkerの`200`全文を保存し、clientへ`206`を返す。既存Range処理を削除するとlocal・cache無効時のseek契約を失うため、fallbackとして残す。

## 3. Rejected

1. **browser向けheaderだけを付ける案** — 同じ端末では効くが、browser cache missのたびにWorkerとR2が動き、先行DecisionのCloudflare edge cacheを満たさない。
2. **Worker Cache APIを直接呼ぶ案** — Access配下では利用できず、data center localでtiered cacheとrequest collapsingも持たない。Workers Cacheより低水準な制御をcodeへ増やす理由がない。
3. **episode一覧を音声と同じ長期cacheにする案** — 日次追加が長時間見えず、更新頻度の異なるrepresentationを同じ方針へ畳むとLeast Astonishmentに反する。
4. **一覧の差分更新を作る案** — 現在の件数・利用者・更新頻度では短いTTLの全件responseで足り、同期protocolとclient mergeの追加costに見合わない。
5. **初手からvalidator・Generator通知・自動purgeを揃える案** — 原則不変の音声と短命な一覧は有限TTLで自然更新できる。即時訂正の要求が無い段階でcross-runtimeの失効経路を作るのは過剰である。
