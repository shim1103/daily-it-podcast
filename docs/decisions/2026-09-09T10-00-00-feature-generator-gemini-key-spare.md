---
name: 原稿 fallback（Gemini generateContent）の API key を TTS と別の SPARE_GEMINI_API_KEY にする
date: 2026-09-09T10:00:00
branch: feature/generator-gemini-key-spare
---

## 1. Decision

1. `GeminiConfig`（`internal/config/config.go`）へ `SpareAPIKey Secret` を足し、config 契約の必須 env に `SPARE_GEMINI_API_KEY` を加える。`APIKey`（`GEMINI_API_KEY`）は TTS が、`SpareAPIKey`（`SPARE_GEMINI_API_KEY`）は原稿 fallback が使う。env 名・field の正本は `internal/config/names.go` / `config.go` の source。
2. `newGeminiTextWriter`（Composition）だけを `cfg.SpareAPIKey` へ向ける。`newGeminiSpeechSynthesizer`（TTS）は `cfg.APIKey` のまま。UseCase・Adapter の signature は変えない。
3. Decision `2026-09-07T19-06-00` §1-3「TTS で使う `GEMINI_API_KEY` を流用し、新しい env / Secret は足さない」を本 Decision が supersede する。fallback が本番で実際に使われる secondary であり、流用のままでは TTS と枠を食い合うことが実測で判明したため。transport を独立させる判断（同 Decision §1-6）は維持する。
4. GHA 登録は本番 `SPARE_GEMINI_API_KEY`、System test `TEST_SPARE_GEMINI_API_KEY`。`generator-produce-episode.yml` / `generator-system.yml` の env 注入と、System test の credential 欠落 Skip 判定 key に加える。key 未登録なら System は Fatal ではなく Skip で止まる。
5. 「なぜ分けるか」の理由は `GeminiConfig` の invariant コメントを正とし、`DEPLOY.md` は「どの key がどの区分か」「System test が両方要る」だけを持つ。

non-scope: fallback provider を Gemini 以外へ替えること。TTS 側の key 分離。`SPARE_GEMINI_API_KEY` に有料 tier を割り当てるか否か。fallback 発火の能動通知。

## 2. Reason

### なぜ流用（Decision 2026-09-07T19-06-00）を撤回して key を分けるか

流用時の想定は「日次 1 回 produce なら fallback が発火しても無料枠（RPD≈1,000 / TPM 250,000）に収まる」だった。実際には本番 produce run 34209712652 が gemini fallback 経路で HTTP 429 に到達して失敗した（`generator-lane.md` D 表）。原因は、同一 `GEMINI_API_KEY` で TTS（1 エピソード分の複数 topic 束を `SynthesizeAll`）と原稿 fallback（generateContent、入力 3〜8 万 tokens）が同じ per-key free-tier 枠を消費し、TTS の消費がある日に fallback が乗ると TPM/RPD が超過するため。key を分ければ、fallback が発火する日でも TTS 側の消費が原稿枠を圧迫しない。

### なぜ config の必須 key として足すか（optional にして未設定なら APIKey へ委譲、ではなく）

optional + 委譲だと「`SPARE_GEMINI_API_KEY` が無ければ従来どおり `GEMINI_API_KEY` を共用」という分岐が 1 つ増え、共用時の 429 という直したい状態を静かに再現できてしまう。config 契約が全 env 必須（`Load` の invariant）で揃っているのに、この 1 key だけ挙動が違うのは Principle of Least Astonishment に反する。必須にすれば「分離されているか」を起動時に `config.Load` が保証し、GHA 側の登録漏れも System test の Skip として即座に表面化する。

### なぜ textwriter だけ SpareAPIKey で TTS は APIKey 据え置きか

分けたい理由は「fallback の突発消費が TTS の定常消費と枠を食い合う」ことなので、消費の性質が違う 2 経路を別 key にすれば足りる。TTS 側まで新 key へ移すと、既存の `GEMINI_API_KEY` 登録・rate 計測 dispatch（`TEST_GEMINI_API_KEY` 直読み）の前提が全部ずれ、変更範囲が要件を超える（YAGNI）。

### なぜ理由を config.go の invariant に置き DEPLOY.md には書かないか

「TTS と fallback で key を分ける」は `GeminiConfig` という境界の不変条件そのもので、その型を変更する者が保存すべき制約（Design by Contract の invariant）。DEPLOY.md にも同じ "なぜ" を書くと二重管理になり、片方だけ古くなる（DRY）。DEPLOY.md は運用 SSOT として「登録する key の一覧と区分」「System test の要求」だけを持てばよい。

## 3. Rejected

1. **`GEMINI_API_KEY` を丸ごと `SPARE_GEMINI_API_KEY` へ改名する案** — env 名は変わるが TTS と fallback が同じ 1 key を共有する構造が残り、429 の原因（枠の食い合い）を解消しない。"spare" の語だけ入って実体が伴わない。
2. **`SpareAPIKey` を optional にし、未設定時は `APIKey` を使う案** — 「分離されていない状態」を静かに許容する分岐が増える。直したい 429 状態をそのまま再現でき、config 契約の「全 env 必須」から 1 key だけ外れて一貫性を欠く。
3. **TTS も含めて Gemini 利用を全部 `SPARE_GEMINI_API_KEY` へ寄せる案** — 既存の `GEMINI_API_KEY` 登録・`TEST_GEMINI_API_KEY` 直読みの rate 計測 dispatch の前提が全てずれる。分けたい理由は 2 経路の枠分離であって、片寄せは変更範囲が要件を超える。
4. **key を分けず、尺モデルを旧値へ戻したまま fallback の出力 token を絞って無料枠に収める案** — 既に旧尺へ revert 済み（`3cf8f23`）だが、それは「尺を伸ばせない」制約を受け入れただけで、TTS 消費が多い日に fallback が乗れば依然 429 になりうる。root cause は per-key 枠の共有なので、そこを断つ。
5. **本番 produce workflow_dispatch で今日中に検証まで通す案** — 本 session 時点で本番 Gemini 枠が 429 中。shim 指示により本日は produce を回さず、`generator-system.yml` を `--ref feature-branch` で dispatch して config 変更 + fallback 経路の疎通だけ実証した（run 34298458327 PASS、`from=cursor to=gemini` の実切替を伴う e2e 完走）。本番 produce の workflow_dispatch は次の枠回復日に回す（`generator-lane.md` へ送り）。
