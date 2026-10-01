---
name: SpeechSynthesizer 実装は DurationSec を自前の PCM 長から算出し、application/build を import しない
date: 2026-10-01T11:50:38
branch: refactor/generator-wav-duration
---

## 1. Decision

1. `port.SpeechSynthesizer` の実装（gemini adapter、infrastructure 層）は、`DurationSec` を `len(pcm) ÷ byteRate` で算出する。`byteRate` は既存の定数 `pcmSampleRate`・`pcmChannels`・`pcmBitDepth` から求める。
2. infrastructure 層は `application/build` を import しない。`build.WavDurationSec`・`parseWAV` を adapter から呼ばない。WAV header の再 parse もしない。
3. `DurationSec` が `build.WavDurationSec(Content)` と一致することは、`apps/generator/test/` の等価性 test で担保する。`apps/generator/test/` は層に属さない外部の test tree なので、この test は `application/build` を import してよい。infrastructure package 内の test は `build` を import しない。
4. adapter が返すのは segment 単位の再生尺だけである。episode 尺方針（segment 間の無音挿入・`build.Timeline`・`ConcatWAV` による結合）は Application に残し、adapter へ移さない。
5. 尺を `SynthesizeAll` の戻り値契約に含め、infrastructure 実装が返すこと自体は [[2026-09-23T17-21-30-refactor-generator-go-performance]] が正（`ConcatWAV` の責務維持も同 Decision 3）。本 Decision はその「infrastructure 実装がどう算出するか」だけを決める。

## 2. Reason

1. `infrastructure.md` §6（import ルール 1）が infrastructure 層に import を許すのは Entities・Application IF（Port）・外部 SDK だけである。`application/build` は Application 層の実装 package であり、Port ではない。adapter から `build.WavDurationSec` を呼ぶと、外側の層が内側の実装 package に依存し、層境界を破る。
2. adapter は `pcmToWAV` で自分自身が WAV を組み立てている。つまり PCM の長さ（`len(pcm)`）と、header に書く `byteRate` を、組み立て時点で既に知っている。出来上がった WAV を `parseWAV` で読み直すのは、自分が書いた値を header から読み戻すだけの二度手間になる。PCM 長から直接割る方が単純である。
3. `byteRate` を求める定数（`pcmSampleRate`・`pcmChannels`・`pcmBitDepth`）は `pcmToWAV` が header に書く値と同じ定数である。定数が 1 箇所なので、header の値と尺計算の分母がずれる余地は小さい。
4. 算出式を adapter 側に持つと、`build.WavDurationSec`（WAV header 由来）と式が二重化する。一致が崩れる危険は、`apps/generator/test/` の等価性 test が `DurationSec` と `build.WavDurationSec(Content)` を比較することで検知する。この test が `build` を import してよいことは確定している。`apps/generator/test/` は層に属さない外部の test tree で、`test/system/tts_rate_system_test.go` が `application/build` を、`test/integration_support_test.go` が `application` を既に import している。`infrastructure.md` §6 の import ルールは本番 code の層への規則であり、この test tree には及ばない。逆に infrastructure package 内の test が `build` を import すると、本番 code が許されない依存を test に持ち込むことになるので避ける。
5. segment 単位の再生尺と episode 尺方針は別物である。segment 尺は adapter が自分で作った PCM の長さに由来する値で、episode の構成を知らなくても求まる。一方、segment 間の無音挿入（`SegmentSilenceSec`）・`build.Timeline` の開始秒算出・`ConcatWAV` による結合は、episode の構成（topic・ending の並び）を知る処理で、Application に残る。この区別により、先行 Decision との関係は次のとおり整理できる。
   1. [[2026-08-25T22-37-31-feature-generator-cmd-usecase-boundary]] の Rejected 2（尺・結合を Adapter に閉じる案）が却下したのは「Port/Adapter が episode の尺方針と結合を知る」ことである。本 Decision の adapter は segment の PCM 長を返すだけで、episode の尺方針も結合も持たないので抵触しない。同 Decision 1 の RIFF 解析（`WavDurationSec`・`ConcatWAV`）は Application 非公開 helper のままで、adapter は RIFF を解析しない。
   2. [[2026-09-13T13-40-29-feature-playback-now-playing-audio-listenability]] の Decision 2 のうち、結合と episode 尺方針を Application 非公開 helper が持つ部分は維持される。segment 単位の尺が `SpeechSynthesizer` の契約へ移ったことは [[2026-09-23T17-21-30-refactor-generator-go-performance]] が決めており、同 Decision 2 の記述はその点だけ直した。
6. 先行 Decision [[2026-09-23T17-21-30-refactor-generator-go-performance]] は「infrastructure 実装が尺情報を返す」とするが、算出手段までは定めていない。したがって答えは変わらず、手段の軸が足りないだけである。独立に答えが変わり得る別軸で、Reason・Rejected も共有しないため、既存 file は更新せず別 file とした（`decision.md` §2「overarchingと派生」・§3）。

## 3. Rejected

1. **infrastructure から `application/build`（`build.WavDurationSec`）を import する案** — `infrastructure.md` §6（import ルール 1）に反する層境界違反。実装は一番短いが、Port ではない実装 package への依存が内向きに増える。
2. **`WavDurationSec`・`parseWAV` を Entities など共有位置へ移す案** — import 境界は守れるが、`ConcatWAV` と `wav_duration.go` の変更が必要になり、変更範囲が広がる。また [[2026-08-25T22-37-31-feature-generator-cmd-usecase-boundary]] は RIFF の読み書きを Entities 公開にしない判断をしており、それとも緊張する。
3. **adapter 内で WAV header を再 parse する処理を複製する案** — `pcmToWAV` が書いた値を読み戻すだけで、parse ロジックの二重管理になる。PCM 長からの除算で足りる。
4. **episode 尺方針（無音挿入・`Timeline`・結合）も adapter 側へ寄せる案** — [[2026-08-25T22-37-31-feature-generator-cmd-usecase-boundary]] の Rejected 2 と同じく、adapter が episode の構成を知ることになる。adapter が持つのは自分の PCM 長由来の segment 尺までとする。
