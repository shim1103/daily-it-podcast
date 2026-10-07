---
name: generatorの外部API失敗は、再試行・fallback・即errorのどれにするか
description: 4xx（429除く）はbugとして即error。429は回復の明示がある時だけ再試行し、5xx・通信断は1回再試行する。使い切り、または回復の明示が無い429はfallbackする。例外として401/403と公式の利用枠喪失codeもfallbackする。分類はstatusなどを再試行定数へ写す純粋関数に置く。
---

## 1. Decision

1. 失敗の分類は、response statusと標準header`Retry-After`と、providerが公式に定めた機械可読codeだけで行う。response bodyのmessage・文言は読まない。
2. 分類は、これらを再試行方針の定数へ写す純粋関数に置く。retry loopは定数だけを見る。`retryRateLimited`のような再試行を表す定数は、実際に再試行する時にしか返さない。
3. 4xx（429を除く）は呼び出し側のbugとして、再試行もfallbackもせずそのままerrorにする。例外としてfallbackするのは、API keyやsubscriptionの失効（401・403）と、公式が利用枠の喪失を示すcode（Gemini TTSの`quota_exceeded`、Cursorの400 + `usage_limit_exceeded`）である。
4. 429は、回復の明示（`Retry-After`。TTSは公式`error.code`の`rate_limit_exceeded`も）がある時だけ待って再試行し、使い切りはfallbackする。回復の明示が無い429は、待たずにfallbackする。
5. 5xx・通信断（Do error・body読み取り途中断）・TTSの音声欠落は、再試行し、使い切りはfallbackする。idempotentな呼び出しは1回、TTSは同種2連続まで。非idempotentなCursorのcreateは再試行せずfallbackする。
6. fallbackはAdapterが`port.ErrSourceExhausted`をwrapして示す。呼び出し元のUseCaseは番兵の有無だけを見る。この番兵は「利用枠が尽きた」ではなく「この取得元は当面使えない」を表し、原因と課金Tier（`TierFree`・`TierPaid`）を問わない。再試行の途中で成功した場合は成功を返す。

## 2. Reason

1. 4xxは要求側の誤りを表す。別sourceへ渡すと誤りを隠したまま次sourceで再発し、原因が見えなくなる。赤で止める方が直しやすい。
2. 429・5xx・通信断は、その取得元が今使えない状態を表し、別の取得元なら成功しうる。通信断がclient起因かserver起因かは区別できない。しかしfallbackは1回の呼び出しが増えるだけで、両方失敗しても合成layerが最後のerrorを返すため、診断は失われない。1回の再試行で一過性を拾い、それでも駄目ならfallbackする方が、切り替えを諦めるより損が小さい。
3. run 37396892485で、Geminiがquota超過の429を返し、自前のbackoff（1s・2s・4s）で3回待ってからfallbackした。待っても戻らない429を待っていた。`Retry-After`の無い429は、回復の約束が無い。
4. bodyのmessageや`RESOURCE_EXHAUSTED`の文字列はproviderが変えうる。変わると方針が黙って壊れ、provider仕様への依存がcodeへ染み込む。statusと標準headerは、providerを問わず使える共通の契約である。
5. TextWriterの2 providerは、429を一過性と利用枠喪失に分ける公式の機械可読codeを確認できなかった。Geminiの429 bodyは`status: RESOURCE_EXHAUSTED`の1種で、Cursorは公式docsにerror bodyのcodeが定義されていない。共通項は`Retry-After`だけである。
6. TTSは公式API errorsが429を`rate_limit_exceeded`（分・秒単位）と`quota_exceeded`（日次）に分けるcodeを定める。1 episodeで複数回呼びRPMに当たって待てば戻るため、一過性側のbackoff回復は残す。`quota_exceeded`は400でも返りうるので、statusを問わず見る。
7. 呼び出し元のUseCaseは番兵だけを見て切り替える。何をfallbackとするかはAdapterが決め、呼び出し元はstatusを知らない。
8. status分岐をretry loopやfetch関数へ散らすと、どのstatusがどの方針か読めない。純粋関数に抜けばtable testでstatusごとの方針を固定でき、loopは「定数に応じて待つか返すか」だけになる。
9. 401・403を例外にするのは、原稿のfallbackがCursorのsubscription失効（401・403）を契機に導入されたためで、外すと元の目的を失う。
10. 429の分類は課金区分ではなく、そのcredentialを現在使えるかを表す。run 35188139273で、primary Geminiが429を使い切った後にCursorへ切り替わらず停止した。
11. 未実測: GeminiとCursorの429に`Retry-After`が実際に付くか。付かなければTextWriterの429は実質すぐfallbackになる。1日1回のproduce運用では一過性のrate limitに当たる頻度が低く、fallbackが受けると推論している。

## 3. Rejected

1. `TierFree`だけfallback扱いにする案 — Paidも429なら同じく使えず、課金区分でerrorの意味を変える根拠がない。
2. 429を常に即時fallbackする案 — 回復を明示した429（`Retry-After`、TTSは`rate_limit_exceeded`も）をbackoffで回復する経路を失う。回復の明示が無い429に限って採った。
3. 429を`Retry-After`の有無を問わず`MaxAttempts`までbackoffで再試行する旧答え — quota超過の429を待ってしまう（run 37396892485）。
4. bodyの文言（quotaを示す文）で一過性と利用枠喪失を分ける案 — 文言が少し変わるだけで判別が壊れる。
5. bodyの`details`（`QuotaFailure`の`quotaId`など）を読んで分ける案 — 公式の安定した契約として確認できず、命名はprovider都合で変わりうる。
6. 4xx全てをbugとして即errorにする案（401・403も含む） — 原稿fallbackの導入理由を失う。例外としてfallbackに残した。
7. 通信断をfallbackしない旧答え（「別sourceも同じnetworkの先にある」） — 未検証の推測だった。client起因か区別できず、別sourceは別hostなので成功しうる。1回再試行してからfallbackする形へ変えた。
8. 5xxを再試行せず即fallbackする案 — idempotentな呼び出しは1回の再試行で一過性を拾える。
9. status分岐を各fetch関数やretry loopへ直書きする旧形 — 方針が分散して読めず、`retryRateLimited`が再試行しないケースにも使われていた。
