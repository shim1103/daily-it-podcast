---
name: 情報源 Adapter の失敗は Infrastructure Error を return するだけとし、fallback・transaction を持たず retry は transient HTTP のみ最小に留める
date: 2026-09-02T15:27:00
branch: feature/hackernews-api-adapter
---

## 1. Decision

主 Decision（`2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）で追加する3 Adapter（HackerNews / Lobsters / ITmedia）の失敗時挙動を定める。

1. **Adapter は失敗を `*<pkg>.Error{Op, Err}` として return するだけ。** Application は Infrastructure Error を Domain Error へ変換しない（先行 Decision `2026-08-30T18-30-00-feature-generator-system-e2e-produce-episode.md` §4 と同型）。cmd/ への伝播は `ProduceEpisode.Run` が `fetch.Run` の error をそのまま return し、`compositeItemSource.List` が「いずれか1源が error なら nil, err」を返す既存契約に乗る。
2. **fallback しない。** composite の既存契約どおり、1源でも error なら composite 全体を error にし、その run の episode 生成を中止する。「残った源だけで episode を作る」ことはしない。
3. **retry は transient HTTP 失敗のみ、最小限。**
   1. `client.Do` の error（接続断・タイムアウト）と 5xx は **1 回だけ即再試行**する。2 回目も失敗したら `*<pkg>.Error` を return する。backoff は入れない（1 回だけなので）。
   2. 4xx（429 含む）・応答 body の parse 失敗・想定外 schema は再試行せず即 return する。
   3. N+1 の個別 item / comment 取得（HackerNews の `item/<id>.json`、Lobsters の `/s/<id>.json` の comment）は、**その要素の取得失敗ならその要素を落として `List` を続行**する。一覧 endpoint（`topstories.json` / `hottest.json` / feed 本体）の取得失敗は `List` ごと失敗させる。
4. **Adapter は独自 timeout を持たない。** `http.NewRequestWithContext` で `ctx` を伝播するだけ（先行 Adapter getxapi と同型）。1 リクエストの上限は Composition の共有 `httpTimeout`、run 全体の上限は GHA job timeout に委ねる。並行取得はしない（逐次）。
5. **transaction 機構を足さない。** `ProduceEpisode.Run` の既存契約「途中 error なら WriteEpisode.Run を呼ばない（書込なし）」「WriteEpisode は最後に一度だけ」で充足する。Adapter は Drive に触れない。

## 2. Reason

### 伝播

Generator CLI に client UI は無く、Internal → External の写像は Driving Adapter の Delivery に1箇所だけ置く（先行 Decision `2026-08-30T18-30-00`）。Adapter が `*<pkg>.Error{Op, Err}` の形（A の `error.go` stub が既にこの形）で return し `Unwrap()` を実装していれば、`ProduceEpisode.Run` → cmd/ まで型が保たれ、kind=infrastructure / op=`<Op>` / cause 連鎖が stderr に出る。secret は3源とも無いため、stderr snippet 漏洩の考慮も不要。Adapter 側で新たに決めることは Error の `Op` slug 命名だけで、それは A の stub 実装コメントが持つ。

### fallback をやらない

先行 Decision（`2026-08-15T16-39-20-feature-x-api-adoption.md`）§Rejected が「第三者 API + scraper fallback 多層構成」を「相関 failure する / fallback が成立しない」と退けた。今回の3源は相関 failure しない別 route だが、それでも fallback を入れないのは別理由: 3源は「議論2 + 報道1」で役割が違い、報道枠（ITmedia）が欠けた episode は片肺になる。1源が落ちたら全体を止めて赤にする方が、運用者が「今日の episode は情報が欠けている」に気づける。先行 Decision（`2026-08-30T16-23-00-feature-generator-system-e2e-produce-episode.md`）§3「書込前終了は precondition 不成立として赤にする。skip で緑にしない」と同じ思想。「残った源で作る」を将来入れるなら composite の契約変更であり、別 Decision と lane 項目にする。

### retry を最小にする

Gemini TTS の重装 retry（先行 Decision `2026-09-02T13-56-00-feature-generator-system-e2e-produce-episode.md`: 同種2連続打ち切り / MaxAttempts=3 / callGap=20s / backoff 60s〜3m）は、無料枠 RPD=15 という Gemini 固有の制約への対処である。HackerNews は rate limit 無しを運営元が明記、Lobsters・ITmedia も無料の公式配信で、あの機構の前提が無い。1日1回の cron なので、transient なネットワーク瞬断を 1 回拾えれば十分で、それ以上は決定論的失敗（4xx / parse）か仕様変更の兆候（429 が返る）であり、retry しても同じか、運用者が気づくべき事象。N+1 の個別要素を落として続行するのは、500 件中 1 件の comment 欠落で episode 生成全体を止めるのが過剰なため。一覧 endpoint が死んだら題材ゼロなので、そこは `List` ごと失敗させる。

### timeout と並行化

1 リクエストの timeout を Adapter が独自に持つと、Gemini だけ専用 timeout にした先行 Decision（`2026-08-30T22-10-00-feature-generator-system-e2e-produce-episode.md`）のような個別調整が源ごとに分散する。3源は長文 TTS のような長い headers 待ちが無いので共有 `httpTimeout` で足り、`ctx` 伝播だけ持てば run 全体の deadline（GHA job timeout）が最終防壁になる。HackerNews の逐次 N+1（一覧上位 + 各 story の comment。取得件数の上限は Adapter stub の定数が正本）が GHA job timeout に収まるかを実測してから、収まらなければ並行化を検討する（YAGNI。lane 送り）。

### transaction

`ProduceEpisode.Run` は fetch → brief → draft → TTS → concat のあと最後に一度だけ `WriteEpisode.Run` で Drive に put する。途中で落ちれば Drive に何も書かれず、ロールバック対象が無い。json/wav ペアの原子性は先行 Decision（`2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`）と Drive upsert の範囲で既に扱われており、source Adapter は Drive に触れないため無関係。

## 3. Rejected

1. **1源が落ちても残りの源で episode を作る案** — composite の契約変更が要る。報道枠が欠けた片肺 episode を黙って出すより、全体を止めて運用者に気づかせる方がよい。将来必要なら別 Decision。
2. **Gemini と同じ重装 retry（同種連続打ち切り・callGap・長 backoff）を3源にも入れる案** — あれは無料枠 RPD=15 への対処で、rate limit 非公表の公式3源には前提が無い。transient HTTP の 1 回再試行で足り、それ以上は決定論的失敗か仕様変更の兆候。
3. **429 を retry 対象に含める案** — 3源は rate limit を公表しておらず、429 が返ること自体が仕様変更の兆候。backoff して粘るより、即失敗させて運用者が気づく方がよい。
4. **個別 item / comment の取得失敗でも `List` ごと失敗させる案** — 500 件中 1 件の欠落で episode 生成全体を止めるのは過剰。一覧 endpoint の失敗（題材ゼロ）だけを `List` の失敗にする。
5. **Adapter に個別の timeout を持たせる案** — 源ごとの timeout 調整が分散する。`ctx` 伝播と共有 `httpTimeout`、GHA job timeout の3層で足りる。
6. **item / comment fetch を最初から並行化する案** — 逐次の実測が GHA job timeout に収まるか未確認。収まるなら並行化は不要な複雑さ（YAGNI）。収まらなければ lane から着手する。
7. **source 取得を跨いだ transaction / 中間結果の永続化を足す案** — 「途中 error なら書込なし・最後に一度だけ WriteEpisode」で充足済み。source Adapter は Drive に触れない。
