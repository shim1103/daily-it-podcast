## 1. Summary

`application/build.WavDurationSec` と `build.ConcatWAV` を実装する。segment 間無音は `SegmentSilenceSec` 定数どおり PCM 無音 insert とする。

## 2. Context

1. TTS segment 分割と無音 insert は Builder 責務である（Decision `14-10`、`14-15`）。
2. `SpeechSynthesizer` Port は拡張しない。
3. 両 func は stub のまま。

## 3. Canonical Sources

1. `docs/decisions/2026-08-29T14-10-00-docs-produce-episode-run-spec-tts-segment-split.md`
2. `docs/decisions/2026-08-29T14-15-00-docs-produce-episode-run-spec-wav-concat-segment-silence.md`
3. `apps/generator/internal/application/build/wav_duration.go`
4. `apps/generator/internal/application/build/wav_concat.go`
5. `apps/generator/internal/entities/constants/speech_segment.go`

## 4. Scope

### In Scope

1. 同一 PCM パラメータ WAV から durationSec を算定する。
2. 複数 WAV を結合し、隣接 part 間に無音 PCM を挿入する。
3. corrupt input は `corrupt_speech_audio` で fail する。
4. unit test（fixture WAV、無音長 assert）。

### Out of Scope

1. TTS orchestration、manuscript timing field の最終組立（`ProduceEpisode.Run` = D）。
2. vendor 固有 PCM 変換。
3. `SegmentSilenceSec` 数値チューニング（D）。

## 5. Contract

1. vendor 定数を使わず header から PCM パラメータを読む。
2. 無音 insert 分は duration 算定に含める。
3. 入力形式不一致は error（best-effort 変換しない）。

## 6. Constraints

1. A/B で固定されていない型・Port を新設しない。
2. Infrastructure package を import しない。

## 7. Acceptance Criteria

1. [ ] `WavDurationSec` / `ConcatWAV` が panic stub でない。
2. [ ] 2 本以上の part 結合で無音 insert が観測できる test がある。
3. [ ] corrupt fixture で `corrupt_speech_audio` になる test がある。
4. [ ] `go test ./internal/application/build/...` が pass する。

## 8. Verification

```bash
cd apps/generator
go test ./internal/application/build/... -count=1 -run 'Wav|Concat'
go build ./...
```

## 9. Dependencies

1. A/B artifact（`wav_*.go` stub、`speech_segment.go`）。
2. 他 build Issue と並行可。

## 10. Risks

1. fixture 配置を repo 慣習外にすると gate が複雑化する。既存 test dir 慣習に合わせる。

## 11. Notes

1. GitHub Issue 化は別判断。本 file が達成契約の正。
