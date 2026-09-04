package gemini

import "time"

const (
	// why: 公式 Interactions TTS preview。変更は Adapter 定数だけで閉じる。
	ModelID     = "gemini-3.1-flash-tts-preview"
	VoiceName   = "Charon"
	EndpointURL = "https://generativelanguage.googleapis.com/v1beta/interactions"
	// why: 公式 Limitation。演出文を読み上げないよう preamble と Transcript ラベルを分ける。
	EnvelopePreamble = "Synthesize the following speech. Read only the transcript below.\n\n"
	TranscriptLabel  = "#### TRANSCRIPT\n"
)

// MaxAttempts は 1 セグメントが連続で消費してよい Gemini 呼び出しの上限。無限 retry を防ぐ。
// why: 同種 error 2 連続で打ち切るので、実効上限は「異なる Op が交互 = 3 回」程度。
//
//	1 セグメントの暴走で SynthesizeBudget 全部を食わせない二段構えの内側（Decision 2026-09-02T13-56-00）。
const MaxAttempts = 3

// SynthesizeBudget は 1 度の SynthesizeAll 呼び出し全体で許す Gemini 呼び出しの合計上限。
// why: 無料枠 RPD=15 を 1 episode で焼き切らないため、セグメント単位の MaxAttempts ではなく
//
//	呼び出し群の合計で絞る。SynthesizeAll は残予算をセグメント横断で消費し、
//	各セグメントは min(MaxAttempts, 残予算) 回まで。合計が SynthesizeBudget へ達したら以降のセグメントは即 error。
const SynthesizeBudget = 15

// 429 / 503 / 5xx 再試行の待機の既定値。
// why: 20s 起点でも System で 429 が尽きる（run 33314746860, ~476s）。60s 起点・上限 3m へ。
// why: rate 計測は SpeechSynthesizer の field へ注入して差し替える（Decision 2026-09-03T14-46-00）。
//
//	既定 constructor（NewSpeechSynthesizer）はこの const 値を使うので挙動は不変。
const (
	defaultRetryBackoffBase = 60 * time.Second
	defaultRetryBackoffMax  = 3 * time.Minute
)

// defaultCallGap は client.Do どうしの最小間隔（成功・失敗を問わない）の既定値。
// why: 無料枠 3 RPM = 20s 間隔に合わせ、連続 segment の 429 を防ぐ（Decision 2026-09-02T13-56-00）。
const defaultCallGap = 20 * time.Second

// httpCallTimeout は Gemini TTS 1 呼び出しの Client 全体 timeout である。
// why: 120s でも長文朗読で awaiting headers が切れた（run 33310692613）。
// Composition から渡る *http.Client は全体 timeout を持たないので、この値は Adapter が付け直す。
const httpCallTimeout = 5 * time.Minute

// Gemini TTS が返す raw PCM の形式（公式 L16: 24 kHz / 16-bit / mono）。
// WAV wrap と decode の前提。HTTP 定数（ModelID 等）とは別責務。
const (
	pcmSampleRate = 24000
	pcmChannels   = 1
	pcmBitDepth   = 16
)

// minSpeechDurationSec は「実質無音でない」とみなす raw PCM の最小尺（秒）。
const minSpeechDurationSec = 0.5

// minPCMBytes は minSpeechDurationSec 相当の raw PCM バイト数。これ未満の PCM は
// 非空でも「実質無音の極小応答」として retryable な decode 失敗に落とす。
// why: Gemini が HTTP 200 で len(pcm)==2（1 サンプル ≒ 1/24000 秒）のような極小 PCM を
//
//	返すことがある。decode_pcm の audio 欠落 500 相当と同じ一過性劣化なので、
//	Adapter のループ内で retry し、非空・最小尺の WAV を contract として保証する。
const minPCMBytes = int(pcmSampleRate * pcmChannels * (pcmBitDepth / 8) * minSpeechDurationSec)
