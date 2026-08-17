## Gemini TTS Adapter — Issue draft

GitHub Issue 未作成。作成時はこの本文を使う。`shim gh` はこの file 作成だけでは実行しない。

**Title（案）:** `feat(generator): Gemini TTS で SpeechSynthesizer を実装する`  
**type / priority（案）:** `feat` / （`workflow/constants.toml` 確定後に合わせる）

```markdown
## 1. Summary

このIssueでは、既存の `SpeechSynthesizer` を Gemini TTS Adapter で満たし、Composition から結線できるようにする。
完了後、本文 1 件を渡すと非空の mp3 bytes が返り、vendor 型は Application へ出ない。

## 2. Context

事実:

- Port・戻り型・秘密名は境界 stub 済み。`Synthesize` 本体・HTTP・PCM→mp3・Error 型・Composition 結線は未実装。
- Adapter 定数 `ModelID` / `VoiceName` / `EnvelopePreamble` / `TranscriptLabel` / `EndpointURL` は空文字。このままでは HTTP できない。値は同じ定数へ埋める（decision）。
- `MaxAttempts` は既に `4`。無限 retry 防止の正はここ。減らして 0 や負にしない。
- 戻り形式の正は Drive 配置の mp3。Gemini Developer API の TTS 戻りは raw PCM（公式例は 24 kHz / 16-bit / mono）。mp3 直出しではない。
- 認証は AgentSecrets proxy の custom header 注入。upstream は `x-goog-api-key`。キー名は `GEMINI_API_KEY`（`secretnames.GeminiAPIKeyName`）。値は code に書かない。
- 既存の外向き HTTP 先例は `twitterapiio`（SDK なし、proxy 経由、Infrastructure Error）。
- 公式 TTS は `input` に演出文と本文が同居する。vague だと指示文を読み上げる、または `PROHIBITED_CONTENT` になる（公式 Limitation）。対策は Adapter が本文へ envelope を被せる。
- 公式 TTS Limitation: 稀に音声の代わりに text token が返り 500 になる。automated retry を置け、と書いてある。
- 公式 Troubleshooting: 429 / 503 / 5xx は exponential backoff。400 / 403 は retry するな。
- archive `packages/tts/src/gemini-tts.ts` は SDK + `generateContent` + Kore で、API bytes をそのまま返していた。本流の契約は mp3 なので、archive をコピーしない。
- 呼び出し入口（CLI / UseCase / 原稿 flatten / chunk 分割）は未決のまま。この Issue は Driven Adapter だけ。

仮定（作業は止めない。実測で直す）:

- REST は公式 speech-generation の Interactions 例（`POST https://generativelanguage.googleapis.com/v1beta/interactions`）を正とする。preview の path / model 名は変わりうる。
- model / voice は公式の TTS 対応一覧から Adapter 定数へ 1 つずつ選ぶ。Port 引数にしない。archive 実装は `Kore`。日次 IT news なら Informative の `Charon` が候補。どちらでも Port は変わらない。選んだ値を `VoiceName` へ書く。
- PCM→mp3 は Adapter の private helper。第二 Port にしない。
- 空本文（trim 後空）は Domain Error 型を新設せず、Infrastructure Error で落とす（UseCase がまだ無い）。

## 3. Canonical Sources

- 設計判断 — `docs/decisions/2026-08-17T17-41-59-feature-tts-speech-synthesizer.md`
- Port 契約 — `apps/generator/internal/application/port/speech_synthesizer.go`
- 戻り型 — `apps/generator/internal/entities/models/speech_audio.go`
- Adapter 定数 — `apps/generator/internal/infrastructure/speech/gemini/constants.go`
- 秘密名 — `apps/generator/internal/infrastructure/secretnames/names.go`、`README.md` の `GEMINI_API_KEY`
- Drive 形式 — `contracts/drive-layout.md`（`{episodeId}.mp3`）
- HTTP / Error / Composition 先例 — `apps/generator/internal/infrastructure/x/twitterapiio/`、`apps/generator/internal/composition/twitterapiio.go`、`apps/generator/internal/infrastructure/agentsecrets/`
- Gemini TTS 公式 — https://ai.google.dev/gemini-api/docs/speech-generation （Limitation・voice・REST 例）
- Gemini retry 公式 — https://ai.google.dev/gemini-api/docs/troubleshooting
- 層・依存 — `DESIGN.md`
- 設計哲学（再定義禁止） — `/Users/shim0729/.claude/skills/philosophy/design-philosophy.md`
- 書き方・公開契約 documentation（再定義禁止） — `/Users/shim0729/.claude/skills/coding-style/SKILL.md`
- test方針（再定義禁止） — `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md`

## 4. Scope

### In Scope

- `infrastructure/speech/gemini/` に `SpeechSynthesizer` 実装（HTTP 1 回、応答から PCM 取り出し、mp3 化）
- 空の Adapter 定数へ公式値を埋める（model / voice / envelope / endpoint）
- envelope は定数連結のみ。Port の `text` は Transcript
- retry は Adapter 内部。打ち切りは既存 `MaxAttempts`
- Infrastructure Error 型（twitterapiio と同型でよい）
- Composition から Port へ結線
- Port 契約を検証する sociable unit test（httptest。実 Gemini は不要）

### Out of Scope

- `application/port/speech_synthesizer.go` と `entities/models/speech_audio.go` の契約変更
- UseCase・cmd・GHA・Drive 書込
- 原稿 JSON から本文への flatten、topic 分割、音声結合
- Port 引数への voice / language / speed
- streaming / multi-speaker
- duration / `startSec`
- Free / Paid を code で切る
- `MaxAttempts` を無上限にする変更
- archive TTS 実装の移植

## 5. Contract

- 既存 `SpeechSynthesizer.Synthesize` の `@require` / `@ensure` / `@invariant` を変えない・満たす。
- 成功時の `SpeechAudio.Content` は非空 mp3 bytes。path を返さない。
- 失敗時は Infrastructure Error。retry の途中失敗は外へ出さない。最後の失敗だけが error。
- 秘密はキー名 `GEMINI_API_KEY` のみ。値を保持しない。

## 6. Constraints

- 書いてよい path: `apps/generator/internal/infrastructure/speech/gemini/**`、Composition のこの Adapter 結線のみ、定数の中身。
- Port / `SpeechAudio` の signature・契約tagは触らない。
- SDK を足さない。既存と同様に REST + AgentSecrets proxy。
- 演出・voice・model・endpoint・envelope 文面は Adapter 定数以外へ出さない。
- retry 対象: 一過性（429 / 503 / timeout / 5xx、および公式が書く TTS 固有の audio 欠落 500）。400 / 403 / `PROHIBITED_CONTENT` は即 error。
- philosophy / coding-style / testing-strategy は Canonical Sources の絶対 path を正とし、Issue 本文へ再定義しない。
- 公開境界の契約は Port 宣言箇所のみを SSoT とする。

## 7. Acceptance Criteria

- [ ] AC-1: `SpeechSynthesizer` を実装する型が `infrastructure/speech/gemini/` にあり、Composition から `port.SpeechSynthesizer` として注入できる。
- [ ] AC-2: `Synthesize` が成功時に非空の `SpeechAudio.Content` を返し、中身は mp3 である。
- [ ] AC-3: Application から Gemini の request/response 型・PCM・voice 名・path が見えない。
- [ ] AC-4: model / voice / envelope / endpoint が空文字ではなく Adapter 定数に入っている。Port 引数には出ていない。
- [ ] AC-5: 同一失敗を `MaxAttempts` 回を超えて繰り返さない（上限到達で Infrastructure Error）。
- [ ] AC-6: 秘密は `GEMINI_API_KEY` のキー名参照のみで、リポジトリに key 値がない。
- [ ] AC-7: Port 契約を検証する test が testing-strategy に沿って追加され、generator Unit gate が pass する。

## 8. Verification

- test の Scope×Sociability・配置・credential は `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md` に従う。
- 実 Gemini と実 API key は Unit に使わない。proxy を httptest で stub する（`twitterapiio` の sociable unit と同型）。
- `./scripts/test-unit.sh` の generator 部分（depguard + coverage）が pass する。playback の vitest 未導入は本 Issue の失敗にしない。

## 9. Dependencies

- blocked by: なし（Port / `SpeechAudio` / 空定数 / 秘密名 / `MaxAttempts` は既存）。
- related: 呼び出し UseCase・Drive 書込・cmd は未着手。この Adapter を待たずに Port 契約は読める。

## 10. Risks

- 薄い Error method（`Error` / `Unwrap`）を足すと、coverage 除外は今 `twitterapiio/error.go` だけなので 90% gate が落ちうる → 同型の method は test で cover するか、既存 decision と同じ除外を `gemini/error.go` へ足す。
- preview model / endpoint が公式更新で変わる → 変更は Adapter 定数に閉じ、Port は維持する。
- PCM を mp3 と誤認して Drive へ書く → AC-2 で mp3 であることを検証する。
- envelope 不足で指示文を読み上げる → 公式 Limitation の preamble + Transcript ラベルを定数連結する。
- retry 無しで TTS 固有 500 がそのまま失敗する → Adapter 内部で有限 retry。

## 11. Notes

- 呼び出し入口・chunk 分割・尺計算は follow-up。この Issue の完了条件に入れない。
- GitHub Issue 作成は `shim gh create-issue`（本 draft の Markdown 本文を stdin へ）。この todo 作成だけでは作成しない。
```
