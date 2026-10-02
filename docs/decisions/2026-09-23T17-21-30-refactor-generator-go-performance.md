---
name: segment の尺は port.SpeechSynthesizer の契約に含め、adapter が自分の PCM 長から算出する
date: 2026-09-23T17:21:30
branch: refactor/generator-go-performance
---

## 1. Decision

1. segment（`SpeechAudio` 1 件）の再生尺 `DurationSec` は、`port.SpeechSynthesizer.SynthesizeAll` の戻り値契約に含める。infrastructure の実装（gemini 等）が埋め、呼び出し側は再計算しない。
2. gemini adapter は WAV を自分で組むので、尺は `len(pcm) ÷ pcmByteRate` で求める。`pcmByteRate` は WAV header の byteRate と同じ定数で、尺の式の正本はこの 1 つ。header 由来の尺と一致を保つための等価性 test は置かない。
3. `ProduceEpisode.Run` は `DurationSec` を `build.Timeline` へ渡し、WAV から尺を再計算しない。本番の呼び出し元が無くなった `build.WavDurationSec` は削除する。WAV の解析は `ConcatWAV` が使う非公開の `parseWAV` だけが持つ。
4. `ConcatWAV` と `build.Timeline` の責務と入出力は変えない。episode の尺方針（segment 間の無音、topic・ending の開始秒、全体 duration）は Application に残す。

## 2. Reason

1. 変更前は、`Run` が各 segment を `WavDurationSec` で解析し、直後に `ConcatWAV` が同じ WAV を再び解析していた。尺は WAV を組む側（adapter）が組む時点で既に知っている値で、読み戻して解析する必要がない。
2. 尺を契約に含める理由（4a）: WAV を返すことは `SpeechSynthesizer` の既存契約で、尺も同じ末端が一緒に返すのが自然。`application/speech` に尺専用の wrapper を足すと（4b）、`speech` に入口が 2 つできて責務が分散する。
3. adapter が `application/build` を呼ばない理由: `infrastructure.md` §6（import ルール 1）が infrastructure に許すのは Entities・Application IF（Port）・外部 SDK だけで、`build` は Application の実装 package。
4. 尺の式を 1 つにする理由: header を読む実装と PCM 長で割る実装を併存させると、一致を保つ等価性 test が要り、test と実装が同じ式を持つ二重管理になる。PCM 長から求める 1 実装にすれば、header の byteRate と尺の分母が同じ定数なので、一致は構造で保たれる。
5. segment 尺と episode 尺方針は別物である。segment 尺は adapter が自分の PCM 長から求まる値で、episode の構成を知らない。無音の挿入・`Timeline` の開始秒算出・結合は episode の構成（topic と ending の並び）を知る処理なので Application に残る。[[2026-08-25T22-37-31-feature-generator-cmd-usecase-boundary]] の Rejected 2 が却下したのは後者を adapter に閉じることで、本 Decision はそれと両立する。
6. `ConcatWAV` を変えない理由: 責務は sampleRate・channels・bitsPerSample が同じ WAV の結合だけで、尺を副産物として返すと責務が混ざる。`Timeline` を変えない理由: segment の秒数列を topic・ending・全体の値へ変換する別の抽象度の処理で、尺の出所が変わっても存在理由は変わらない。

## 3. Rejected

1. `application/speech` に尺専用の wrapper を新設する案（4b） — 入口が 2 つになり、WAV を返す契約と尺を返す契約を分ける理由がない。
2. `ProduceEpisode.Run` に尺計算専用の中間 step を新設する案 — `Run` に謎の step が増え、「音声 → 結合」の流れが分かりにくくなる。
3. `ConcatWAV` が尺を副産物として返す案 — 結合に尺という別の関心が混ざる。
4. adapter から `application/build`（`WavDurationSec`）を呼ぶ案 — `infrastructure.md` §6 の層境界違反。
5. 解析関数を Entities 等の共有位置へ移して adapter が使う案 — [[2026-08-25T22-37-31-feature-generator-cmd-usecase-boundary]] が RIFF の読み書きを Entities 公開にしないと決めており、変更範囲も広がる。
6. adapter 内で WAV header を再解析する案 — 自分が書いた値を読み戻すだけで、解析ロジックの二重管理になる。
7. 尺の式を header 由来と PCM 長由来の 2 実装で持ち、等価性 test で結ぶ案 — 正本が 2 つになり、test が実装の式を写すだけになる。共有定数による 1 実装にする。
8. episode の尺方針（無音・`Timeline`・結合）も adapter へ寄せる案 — adapter が episode の構成を知ることになる。
