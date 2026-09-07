---
name: Cursor 利用枠喪失時に原稿取得を Gemini Flash-Lite へ切り替える UseCase を Application に置く
date: 2026-09-07T19:06:00
branch: feature/generator-text-writer-fallback
---

## 1. Decision

1. 原稿取得の切り替えは **Application の UseCase**（`internal/application/manuscript` の `TextWriter`）が担う。この UseCase は `port.TextWriter` を 2 つ（primary / secondary）と切り替え通知関数 `onFallback func()` を受け取り、自身も `port.TextWriter` を実装する。`ProduceEpisode` は従来どおり `port.TextWriter` を 1 つ受け取り、その中身がこの UseCase になる（`ProduceEpisode` の signature・本文は変えない）。契約値（dir・constructor・番兵 error・定数）の正本は `apps/generator/internal/application/{manuscript,port}` と `apps/generator/internal/infrastructure/manuscript/geminiapi` の source（A）とする。
2. 2 実装の結線は **Composition Root の `newProduceEpisode`** が行う。primary は `newCursorTextWriter`、secondary は `newGeminiTextWriter`、通知は Composition の `logManuscriptSourceSwitched`。`composition/cursorapi.go` は変えない。
3. secondary の model は **Gemini `gemini-2.5-flash-lite`**。TTS で使う `GEMINI_API_KEY`（`config.GeminiConfig`）を流用し、新しい env / Secret は足さない。
4. 切り替えの発動条件は **cursorapi の create が HTTP 401 もしくは 403**。cursorapi は 401/403 のとき、自分の infra error を **vendor 非依存の番兵 `port.ErrSourceExhausted`** で wrap して返す。UseCase は `errors.Is(err, port.ErrSourceExhausted)` だけを見て切り替え、cursorapi の error 型・Op・HTTP status を知らない。429 / 5xx / SSE 途中断 / 到達不可（`Op=="do"`）は番兵で wrap せず、現行の Cursor 内 retry のままにする。
5. 切り替えは **高々 1 回**。secondary の戻り（成功・失敗）をそのまま `ProduceEpisode` へ返す。secondary の error を別型でラップしない。secondary が再び `port.ErrSourceExhausted` を返しても 3 つ目の取得元は無い。
6. `geminiapi.TextWriter` の transport は `speech/gemini`（TTS）を再利用せず独立させる。endpoint は `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent`、API key は `x-goog-api-key` header。model は Adapter 定数固定で、runtime の model 存在確認はしない。
7. `geminiapi` の retry は `speech/gemini` の重装 retry（call gap・synthesize budget・分単位 backoff）を流用しない。`client.Do` error と 5xx を 1 回、429 を有限回（`MaxAttempts`）+ backoff（`Retry-After` delta-seconds 尊重、上限 `MaxRetryAfter`）で再試行する。401/403・その他 4xx・`finishReason` が STOP 以外・空 text・parse 失敗は再試行しない。
8. `geminiapi.Error` は `cursorapi.Error` と同型（`Op` / `Err` / `Unwrap` / package prefix）。secret 実値を `Err` に入れない。`cursorapi.Error` の型は変えない（HTTP status を外へ出す field は足さない。切り替え可否は番兵で表現する）。
9. 切り替えが起きたことは、当面 Composition が渡す `logManuscriptSourceSwitched` が標準 `log`（stderr）へ 1 行出すだけとする。構造化 log 基盤の導入と、能動的な通知（毎日 Gemini に落ちている状態の検知）は本 Decision の scope 外。

先行 Decision `2026-09-03T17-03-33-feature-generator-cursor-cli-to-http-api.md` の §Decision 2「実装が SSE で行き詰まった時の fallback は将来の別判断とし、初回の正は SSE とする」を、本 Decision が具体化する。先行 Decision は supersede しない（transport を Cloud Agents REST + SSE に置く判断は維持し、本 Decision は「その primary が利用枠喪失で使えないときの退避先」だけを足す）。先行 Decision の file 本文は書き換えない。

non-scope: `gemini-2.5-flash-lite` が返す原稿の品質・token 消費・尺下限割れの実測。3 実装目の追加。Gemini 応答 status の詳細な分類。切り替え発火の能動通知。構造化 log 基盤の導入。

## 2. Reason

### なぜ切り替えを Application の UseCase に置くか（Infrastructure の decorator ではなく）

「原稿を 1 本確保する。primary が枯れたら退避先へ回す」は、`ProduceEpisode` が持つ他の停止条件（Fetch 0 件で `no_source_items`、draft 検証を最大 `TextWriterMaxAttempts` 回）と同じ「原稿確保の方針」であり、Application の関心事である。切り替えの停止条件（どの error で退避するか・何回まで・通知するか）を Application の sociable unit で、`ProduceEpisode` の停止条件と同じ層・同じ test 様式で固定できる。

`ItemSource` の複数源 merge は Composition の composite に置いている（`DESIGN.md` §3）が、そちらは「複数源を**対等に** merge、Application は源個数を知らない」。原稿の切り替えは「primary が主・secondary が退避」という**優先順位のある冗長化**で、"どの失敗で次へ行くか" という判断ロジックを伴う。対等 merge と優先順位付き退避は性質が違うので、置き場が分かれても非対称ではなく、それぞれの性質に合った層に置いている。UseCase は `port.TextWriter` を 2 つ受け取るだけで、Cursor / Gemini という具体は Composition が注入するため、Port の「vendor 非露出」invariant は保たれる。

### なぜ番兵 error（`port.ErrSourceExhausted`）か

UseCase が primary の vendor（cursorapi）の error 型を `errors.As` で覗くと、UseCase が primary の実装を知ることになり、切り替え先を増やす・primary を差し替えるたびに UseCase を触る羽目になる。「この取得元は枯れた、別があれば行ってよい」を vendor 非依存の番兵として `application/port` に置き、発生源（cursorapi）が 401/403 のときにそれで wrap すれば、UseCase は `errors.Is` 一行で済み、cursorapi / geminiapi のどちらも import しない。判定の知識は「枯渇を知っている cursorapi 自身」に閉じ、Composition にも UseCase にも漏れない。

### なぜ Gemini `gemini-2.5-flash-lite` か

要件は「Cursor subscription が切れても、できれば完全無料枠で原稿を出し続ける」。候補は Gemini Flash-Lite / Groq Llama / Anthropic Haiku。

1. **新しい secret がゼロ**: `GEMINI_API_KEY` は TTS 用に本番 GHA Secret 登録済み（`DEPLOY.md`）。原稿へ流用でき、鍵管理の追加がない。Groq / Anthropic は新規 Secret 登録と、その失効運用が増える。
2. **同社の中で最も安い model**: `gemini-2.5-flash-lite` は入力 $0.10 / 出力 $0.40（per 1M tokens）。有料換算しても最安級。
3. **無料枠が日次 1 回の produce を包む**: 無料枠は RPD 約 1,000・TPM 250,000・1M context。1 日 1 回の `ProduceEpisode()` は、原稿 1 回あたり入力 3.3 万〜8.3 万 tokens・出力 4 千〜5 千 tokens（System 実測の原稿長 3,362〜3,723 字から換算）、draft 検証 retry を最大数えても 1 日 2〜3 call。RPD・TPM のどちらにも当たらない。
4. **出力が今の 3 倍になっても無料枠内**: 尺モデルを 3 倍（全体 24〜36 分相当）にしても出力は 1.2 万〜3 万 tokens/日で、入力は `{{SOURCES}}` 支配のためほぼ不変。無料枠で $0、有料換算でも月 $1 未満。Anthropic Haiku は無料枠が無く、同条件で月 $2.7〜$10 かかり「完全無料枠」要件に反する。Groq は無料枠内だが新 Secret と日本語原稿品質の未実測リスクを抱える。

### なぜ 401/403 限定の trigger か

Cursor Cloud Agents REST は Dashboard の `crsr_` key 1 種で、subscription / Pro の失効も key の失効も `create` の HTTP 401/403 として現れる（`2026-09-04T15-05-00` の daily で 401 を実証済み）。これが「Cursor が使えない」と確実に言える唯一の観測点である。429 は rate limit で、Cursor 内の backoff retry で回復しうる。5xx や SSE 途中断は一過性で、切り替えると「一時的な Cursor 障害のたびに Gemini 原稿へ落ちる」ことになり、primary の品質を捨てる頻度が上がる。到達不可（`Op=="do"`）は network 要因で、secondary も同じ network の先にあるため切り替えの意味が薄い。

### なぜ primary を降格せず env 切替もしないか

要件は「subscription が切れた場合の fallback」であり、subscription を保つ運用は続く。Cursor を常に secondary へ降格したり無料枠を primary にすると、正常時に Cursor を使わなくなり、「subscription を維持する」前提と矛盾する。`TEXT_WRITER_PROVIDER` のような runtime 切替は、日次 1 回の bat に分岐を 1 つ増やすだけで、切替を人間が正しく操作し続ける運用コストに見合わない（YAGNI）。毎回 primary を先に試す UseCase なら、subscription 復活後は次の run で自然に Cursor へ戻る。

### なぜ `geminiapi` の transport / retry を TTS と別にするか

`speech/gemini` の重装 retry（`SynthesizeBudget`・20s call gap・60s〜3m backoff）は TTS 無料枠の RPM=3 という TTS 固有前提への対処であり、`generateContent` には過剰である。情報源 3 Adapter が既に「transient を 1 回拾うだけ」で足りると決めている（先行 Decision `2026-09-02T15-27-00`）ので、`geminiapi` もそれに倣い、Cursor と同じく 429 だけ有限回粘る。`generateContent` は POST だが **idempotent**（agent のようなリソースを作らず、同 body なら同じ生成試行で副作用が無い）ため、`cursorapi` が create の 5xx を再試行しなかったのと違い、`geminiapi` は 5xx も 1 回再試行してよい。

### なぜ model 存在確認をしないか

先行 Decision `2026-09-03T17-03-33` §Reason「なぜ model list 確認をしないか」と同型。日次 1 回の起動で `GET /models` を列挙しても稀な破壊的変更への過剰防衛にしかならず、失敗したら Infra Error で落ちて定数を直す方が分岐が少ない。

### なぜ log は当面 stderr 1 行か

generator は現状 error 経路以外に stdout/stderr へ出す仕組みを持たない（`cmd/generator` が最終 error を `delivery.Format` で出すのみ）。切り替えが起きたら secondary の成功で痕跡が消えるため、最小限「切り替わった事実」を GHA run log に残す必要がある。標準 `log` で 1 行出せば後から run を見て気づける。構造化 log 基盤を今この Decision で導入すると scope が広がるため、user 合意のうえ暫定策とし、恒久化（sink 注入や provenance）は別判断へ送る。

## 3. Rejected

1. **切り替えを Infrastructure の decorator（`port.TextWriter` を実装し 2 実装を合成する Adapter）に置く案** — `ItemSource` composite と対称に見えるが、`ItemSource` は対等 merge、原稿切り替えは「どの失敗で次へ行くか」の優先順位付き判断を伴う。この判断は `ProduceEpisode` の他の停止条件と同じ Application の関心事で、同じ層で test したい。Infrastructure に置くと「原稿を確保する方針」が Application と Infrastructure に分かれる。
2. **切り替え可否の判定関数（`shouldFallback func(error) bool`）を UseCase へ注入する案** — 注入元（Composition）がその関数の中で cursorapi の error 型を知る必要があり、判定の知識が Composition へ漏れる。cursorapi 自身が番兵で wrap する方が発生源に近く、Composition も UseCase も vendor を知らずに済む。
3. **`cursorapi.Error` に HTTP status の field を足し、外側でそれを見て判定する案** — status を外へ出すと、読み手が cursorapi の error 構造に依存する。「枯渇かどうか」という 1 bit だけを番兵で表せば、error 構造は cursorapi 内に閉じられる。
4. **fallback 先を Groq Llama にする案** — 無料枠は足りるが新 Secret `GROQ_API_KEY` の登録・失効運用が増え、日本語原稿の品質が未実測。既存鍵で済む Gemini より不利。
5. **fallback 先を Anthropic Haiku にする案** — 無料枠が無く、現状でも月 $2.7〜$5.9、出力 3 倍で月 $5〜$10。「完全無料枠で済ませたい」に反する。
6. **Cursor を常に secondary へ降格 / 無料枠を primary にする案** — 正常時に Cursor を使わなくなり「subscription を維持する」前提と矛盾する。要件は「切れたときの退避」であって primary の入れ替えではない。
7. **`TEXT_WRITER_PROVIDER` 環境変数で primary を切り替える案** — 日次 1 回の produce に runtime 分岐を増やし、切替操作を人間が正しく保つ運用コストがかかる。毎回 primary を先に試す UseCase なら復帰も自動で、切替 flag が要らない（YAGNI）。
8. **切り替えを `ProduceEpisode.writeManuscriptDraft` のループへ直接埋める案** — draft 検証 retry（品質不足で書き直し）と取得元切り替え（Cursor が使えない）は軸が別。1 つのループに混ぜると「品質不足で 5 回目に provider が変わる」ような読みにくい挙動になる。独立した UseCase として `ProduceEpisode` の外に切り出し、`ProduceEpisode` はそれを 1 つの `port.TextWriter` として呼ぶ。
9. **`geminiapi` に `speech/gemini` と同じ重装 retry を持たせる案** — TTS 無料枠 RPM=3 という TTS 固有前提への対処であり、`generateContent` には過剰。情報源 Adapter の最小方針に倣う。
10. **切り替え発火時に構造化 log 基盤を導入する案** — generator に log 基盤が無く、導入は本 Decision の scope を大きく超える。暫定の stderr 1 行に留め、恒久化は §1 non-scope として別判断へ送る。
