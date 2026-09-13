# feature(generator): 完成 WAV を ffmpeg で mp3 化し Drive へ書く

## 1. Summary

このIssueでは `EncodeWAVToMP3` を ffmpeg subprocess で本実装し、`ProduceEpisode` が ConcatWAV 直後に encode してから `WriteEpisode` へ mp3 bytes を渡す状態にする。完了後、新作 episode の Drive 上音声は非空 `{episodeId}.mp3`（`audio/mpeg`）になる。

## 2. Context

1. 事実: A が `EncodeWAVToMP3` stub・drive-layout `.mp3`・gdrive `mp3Ext`/`mp3MIME` を固定済み。stub は ProduceEpisode に未結線。
2. 事実: TTS Port はセグメント WAV のまま（Decision `2026-09-13T13-40-29`）。
3. 仮定: CI / 開発機に `ffmpeg` が PATH にある（無ければ Verification で明示 fail）。
4. 運用: 1 Issue = 1 PR。本番 deploy は lane の release 単位に従う（C2 と同着）。

## 3. Canonical Sources

1. 契約: `contracts/drive-layout.md` / `apps/generator/internal/application/build/encode_mp3.go`
2. 判断: `docs/decisions/2026-09-13T13-40-29-feature-playback-now-playing-audio-listenability.md`
3. 着手順: `docs/decisions/2026-09-13T13-41-00-feature-playback-now-playing-audio-listenability.md`
4. 結合・尺: `apps/generator/internal/application/build/wav_*.go` / Decision `2026-08-25T22-37-31`
5. test 方針: testing-strategy（再掲しない）

## 4. Scope

### In Scope

1. `EncodeWAVToMP3` の ffmpeg 本実装（A 足場 test を behavior test へ置換）
2. `ProduceEpisode.Run` で ConcatWAV → EncodeWAVToMP3 → WriteEpisode の結線
3. encoder 失敗時の error 伝播（既存 error 層慣習）
4. generator SU / Narrow で書込名 `.mp3`・MIME mpeg・body 非空を観測

### Out of Scope

1. playback 読取（`playback-audio-mp3-read`）
2. 既存 Drive 上 wav の一括変換（`audio-mp3-batch-migration`）
3. R2 / cache（Decision のみ。本 Issue にしない）

## 5. Contract

1. `EncodeWAVToMP3(wav []byte) ([]byte, error)`: 成功時非空 mp3。`ffmpeg` 不在・非 0 exit は error。
2. `WriteEpisode` へ渡す `SpeechAudio.Content` は mp3 bytes。
3. Drive put の name / MIME は A 定数（`.mp3` / `audio/mpeg`）。

## 6. Constraints

1. TTS Port / Gemini Adapter に encode を入れない。
2. CGO + LAME を入れない。
3. playback Worker 上で encode しない。

## 7. Acceptance Criteria

1. [ ] `EncodeWAVToMP3` の stub 足場 test が behavior test に置換されている
2. [ ] `ProduceEpisode` 成功経路で書込へ渡る音声が mp3 bytes である
3. [ ] `ffmpeg` が無いとき黙って WAV を書かない（明示 fail）
4. [ ] generator unit coverage gate が緑

## 8. Verification

1. `scripts/generator/check-static.sh`
2. `scripts/generator/test-unit.sh`
3. 必要なら Narrow / Broad の音声書込観測

## 9. Dependencies

1. mp3 A/B 済み
2. 本番 deploy は `playback-audio-mp3-read` と同 release（lane `R-mp3-cutover`）

## 10. Notes

本 PR 単独で本番 produce すると、読取側が未追随の期間に再生が割れる。
