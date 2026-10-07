---
name: Geminiの429使い切りはTextWriter・TTSともTier非依存でsource枯渇にする
description: 429の枯渇分類はbodyの文言を読まず、status・Retry-After・公式の機械可読codeだけで行う。TextWriterはRetry-Afterが無い429を待たずに枯渇とし、TTSはerror.codeがquota_exceeded、または回復の明示が無い429なら待たずに枯渇とする。
---

## 1. Decision

1. 429の枯渇分類は、statusと標準header`Retry-After`と、providerが公式に定めた機械可読codeだけで行う。response bodyのmessage・文言は読まない。
2. TextWriter（Gemini・Cursor）は、解釈できる`Retry-After`が付いた429だけ`MaxAttempts`まで待って再試行する。解釈できる`Retry-After`が無い429は、待たずに`port.ErrSourceExhausted`をwrapする。再試行を使い切った時も同じ番兵をwrapする。Cursorのcreate（POST）の429も、再createせず同じ番兵をwrapする。
3. Gemini TTSは、公式`error.code`が`quota_exceeded`の429（400も同様）を待たずに枯渇とする。`error.code`が`rate_limit_exceeded`、または解釈できる`Retry-After`が付いた429だけbackoffで再試行し、使い切ったら枯渇とする。どちらも無い429は待たずに枯渇とする。
4. この分類は`TierFree`と`TierPaid`で変えない。再試行の途中で成功した場合は成功を返す。

## 2. Reason

1. `generator-system` run 35188139273で、primary Geminiが429を使い切った後にCursorへ切り替わらず、Infrastructure Errorで停止した。429は課金区分ではなく、そのcredential sourceを現在利用できない状態を表す。source配列内の位置もAdapterの責務ではない。
2. final sourceが同じ番兵を返しても、合成layerは次sourceが無ければ最後のerrorとして返すため、終端の診断は失われない。
3. run 37396892485で、Geminiがquota超過の429を返し、自前のbackoff（1s・2s・4s）で3回待ってから約7秒後にfallbackした。待っても戻らない429を待っていた。`Retry-After`の無い429は、回復の約束が無い。
4. bodyのmessageや`RESOURCE_EXHAUSTED`の文字列はproviderが変えうる。変わるとretryとfallbackの方針が黙って壊れ、provider仕様への依存がcodeへ染み込む。statusと標準headerは、providerを問わず使える共通の契約である。
5. TextWriterの2 providerは、429を一過性とquota超過に分ける公式の機械可読codeを確認できなかった。Geminiの429 bodyは`status: RESOURCE_EXHAUSTED`の1種で、Cursorは公式docsにerror bodyのcodeが定義されていない。共通項は`Retry-After`だけである。
6. TTSは公式API errorsが429を`rate_limit_exceeded`（分・秒単位）と`quota_exceeded`（日次）に分けるcodeを定める。1 episodeで複数回呼び、RPMに当たって待てば戻るため、一過性側のbackoff回復は残す必要がある。その回復の明示は、公式codeの`rate_limit_exceeded`か`Retry-After`に限る。
7. `Retry-After`が実際に付くかは、GeminiもCursorも未実測。付かなければTextWriterの429は実質すぐfallbackになる。1日1回のproduce運用では一過性のrate limitに当たる頻度が低く、fallbackが受けると推論している。

## 3. Rejected

1. `TierFree`だけ枯渇扱いにする案 — Paidも429なら同じく利用不能であり、課金区分でerrorの意味を変える根拠がない。
2. 429を常に即時fallbackする案 — 回復を明示した429（`Retry-After`、TTSは`rate_limit_exceeded`も）をbackoffで回復する経路を失うため、全面採用はしない。回復の明示が無い429に限り、TextWriterとTTSの両方で採った。
3. 429を`Retry-After`の有無を問わず`MaxAttempts`までbackoffで再試行し、使い切ってから枯渇とする旧答え — quota超過の429を待ってしまう（run 37396892485）。回復を明示した429だけに絞った。
4. bodyの文言（quotaを示す文）で一過性とquota超過を分ける案 — 文言が少し変わるだけで判別が壊れる。
5. TextWriterでbodyの`details`（`QuotaFailure`の`quotaId`など）を読んで分ける案 — 公式の安定した契約として確認できず、命名はprovider都合で変わりうる。
