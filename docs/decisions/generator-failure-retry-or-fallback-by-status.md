---
name: generatorの外部API失敗を、再試行・枯渇（fallback）・即errorのどれにするか、どう分けるか
description: 4xx（429除く）はbugとして即error。429は回復の明示がある時だけ再試行。5xxは1回再試行して枯渇。network断は1回再試行して通常error。401/403と公式の枯渇codeだけ4xxでも枯渇。分類はstatusを再試行定数へ写す純粋関数に置く。
---

## 1. Decision

1. 外部APIの失敗は、response statusと標準header・公式の機械可読codeだけで、次の4方針のどれかへ分ける。bodyの文言は読まない。
2. 4xx（429を除く）は呼び出し側のbugとして、再試行も枯渇扱いもせずそのままerrorにする。例外は、API keyやsubscriptionの失効（401・403）と、公式が利用枠の喪失を示すcode（Gemini TTSの`quota_exceeded`、Cursorの400 + `usage_limit_exceeded`）で、枯渇として扱う。
3. 429は回復の明示がある時だけ再試行し、無ければ待たずに枯渇とする（詳細は`2026-09-17T15-06-00-fix-gemini-429-fallback`）。
4. 5xxは、idempotentな呼び出しなら1回だけ再試行し、使い切りを枯渇とする。非idempotentなCursorのcreateは、再試行せず枯渇とする。
5. network断・body読み取り途中断は、1回だけ再試行し、使い切りは枯渇にせず通常のerrorとする。
6. 分類は、statusを再試行方針の定数へ写す純粋関数に置き、retry loopは定数だけを見る。`retryRateLimited`のような再試行を表す定数は、実際に再試行する時にしか返さない。

## 2. Reason

1. 4xxは要求側の誤りを表す。別sourceへ渡すと誤りを隠したまま次sourceで再発し、原因が見えなくなる。赤で止める方が直しやすい。
2. 5xxと429はprovider側の状態を表し、別sourceなら成功しうる。使い切った後に次sourceへ渡す意味がある。
3. network断は、別sourceも同じnetworkの先にあるため、切り替えの意味が薄い（`2026-09-07T19-06-00-feature-generator-text-writer-fallback`）。
4. 呼び出し元のUseCaseは番兵`port.ErrSourceExhausted`の有無だけを見て切り替える。何を枯渇とするかはAdapterが決め、呼び出し元はstatusを知らない。
5. status分岐をretry loopやfetch関数へ散らすと、どのstatusがどの方針か読めない。純粋関数に抜けば、table testでstatusごとの方針を固定でき、loopは「定数に応じて待つか返すか」だけになる。
6. 401・403を例外にするのは、原稿のfallbackがCursorのsubscription失効（401・403）を契機に導入されたためで、外すと元の目的を失う。

## 3. Rejected

1. 4xx全てをbugとして即errorにする案（401・403も含む） — 原稿fallbackの導入理由（Cursor subscriptionの失効）を失う。例外として枯渇に残した。
2. 5xxを再試行せず即枯渇にする案 — idempotentな呼び出しは1回の再試行で一過性の5xxを拾える。今は代わりに使い切りを枯渇へ変えた。
3. network断も枯渇にする案 — 別sourceも同じnetworkの先にあり、切り替えても同じ失敗を繰り返しやすい。
4. 旧実装のstatus分岐を各fetch関数の中へ直書きする形 — 方針が分散して読めず、`retryRateLimited`が再試行しないケースにも使われていた。
