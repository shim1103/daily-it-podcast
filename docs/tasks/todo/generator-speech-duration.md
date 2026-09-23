## 1. Summary

このIssueでは WAV segment の尺（duration）計算を `port.SpeechSynthesizer` の契約実装側（infrastructure 層）へ移し、`ProduceEpisode.Run` および `ConcatWAV` 側での重複した尺計算を廃止する。完了後、`models.SpeechAudio.DurationSec` に各 segment の尺が埋まった状態で `SynthesizeAll` から返る。

## 2. Context

1. `ProduceEpisode.Run`（`internal/application/produce_episode.go`）は現行、`audios` の各要素へ `build.WavDurationSec` を呼び尺を算出しているが、同じ `parseWAV` 相当の処理が `build.ConcatWAV` 内部でも実行されており、二重計算になっている。
2. `ConcatWAV` の責務は「WAV 結合のみ」であり、topic・ending の区別を知るべきでない。尺計算を `ConcatWAV` 側の副産物として持たせる案は Decision で Rejected。
3. `port.SpeechSynthesizer` は既に WAV を返す契約を持っており、尺情報も同じ箇所（infrastructure の実装が WAV を組み立てる末端）で一緒に返す方針とした（Decision 4a 採用）。
4. `models.SpeechAudio` へ `DurationSec float64` を追加済み（`internal/entities/models/speech_audio.go`）。`port.SpeechSynthesizer` の contract-docs も改訂済み（`internal/application/port/speech_synthesizer.go`）。
5. `build.Timeline` の責務・存在は変更しない。segment 毎の秒数列を受け取り、topic 開始秒・ending 開始秒・全体 duration へ変換する役割のまま。

## 3. Canonical Sources

1. 尺計算統合方針の Decision: `docs/decisions/2026-09-23T17-21-30-refactor-generator-go-performance.md`
2. 契約定義（固定済み）: `apps/generator/internal/entities/models/speech_audio.go`（`SpeechAudio.DurationSec`）、`apps/generator/internal/application/port/speech_synthesizer.go`（`SpeechSynthesizer` contract-docs）
3. 尺計算の既存実装: `apps/generator/internal/application/build/wav_duration.go`（`WavDurationSec`、`parseWAV`）
4. fallback chain実装: `apps/generator/internal/application/speech/synthesizer.go`
5. test方針は `skills/1:terms/testing-strategy/SKILL.md` を参照する。

## 4. Scope

### In Scope

1. `internal/infrastructure/speech/gemini/synthesizer.go`（および他に `port.SpeechSynthesizer` を実装する vendor adapter があれば同様）で、WAV を組み立てた直後に `build.WavDurationSec` 相当の処理を呼び、`models.SpeechAudio.DurationSec` を埋めて返す。
2. `internal/application/speech/synthesizer.go`（fallback chain）が、source から返る `DurationSec` をそのまま透過して呼び出し元へ返すことを確認する（fallback ロジック自体は変更しない）。
3. `internal/application/produce_episode.go` の `Run` 内、現行 L114-123 の尺計算 loop（`build.WavDurationSec` を独自に呼ぶ処理）を削除し、`audios[i].DurationSec` を直接使うよう変更する。
4. 上記変更に伴う既存 test の改修。

### Out of Scope

1. `build.ConcatWAV` の入出力契約変更（Decision で Rejected 済み、変更しない）。
2. `build.Timeline` の責務変更（変更しない）。
3. `apps/generator` の I/O fan-out 化（別 Issue [[generator-io-fanout]]）。
4. `port.TextWriter` 側（原稿生成）への同様の変更（本 Issue の対象外）。

## 5. Contract

1. `port.SpeechSynthesizer.SynthesizeAll` の signature（引数・戻り値の型）は変更しない。`[]models.SpeechAudio` の各要素に既に追加済みの `DurationSec` へ実際の値を埋める実装を行う。
2. `DurationSec` は該当要素の `Content`（WAV）から算出した再生尺（秒）。`Content` が空でない限り `DurationSec >= 0`（`models.SpeechAudio` contract-docs 参照）。
3. `ProduceEpisode.Run` は `build.WavDurationSec` を直接呼ばない（尺計算は `port.SpeechSynthesizer` 実装側の責務に一本化する）。

## 6. Constraints

1. `build.WavDurationSec` / `parseWAV` 自体の実装（`internal/application/build/wav_duration.go`）は変更しない。呼び出し元を `ProduceEpisode.Run` から `port.SpeechSynthesizer` 実装側へ移すだけであり、尺計算ロジック自体の再実装は行わない。
2. `build.ConcatWAV` は変更しない（Out of Scope §4 参照）。`ConcatWAV` 内部の `parseWAV` 呼び出しはこの Issue の対象外（重複計算は残るが、Decision で `ConcatWAV` 側を変更しない方針を確定済み）。

## 7. Acceptance Criteria

- [ ] `internal/infrastructure/speech/gemini/synthesizer.go`（他に実装があれば同様）が `models.SpeechAudio.DurationSec` に実際の尺を埋めて返す。
- [ ] `internal/application/speech/synthesizer.go`（fallback chain）が `DurationSec` を欠落させずに透過する。
- [ ] `internal/application/produce_episode.go` の `Run` から独自の尺計算 loop（`build.WavDurationSec` 直接呼び出し）が削除されている。
- [ ] `Run` が `audios[i].DurationSec` を使って `segmentDurations` を構築している。
- [ ] 既存 test が変更後も pass する。

## 8. Verification

```
cd apps/generator && go build ./... && go vet ./... && go test ./...
```

1. `SynthesizeAll` の戻り値に対し、`DurationSec` が `WavDurationSec` の算出値と一致することを確認する test がある。
2. `ProduceEpisode.Run` の統合 test で、尺計算 loop 削除後も `build.Timeline` への入力が変わらないことを確認する。

## 9. Dependencies

なし。[[generator-io-fanout]] とは独立に実装できる。

## 10. Risks

なし。既存の尺計算ロジック（`build.WavDurationSec`）自体は変更せず、呼び出し元を移すだけの変更。

## 11. Notes

1. Rejected 案（`application/speech` 層への wrapper 関数新設、`ProduceEpisode.Run` 内への中間 step 新設、`ConcatWAV` が尺情報を副産物として返す案）は `docs/decisions/2026-09-23T17-21-30-refactor-generator-go-performance.md` を参照。
