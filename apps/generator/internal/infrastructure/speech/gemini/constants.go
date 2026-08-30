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

// MaxAttempts は Gemini 呼び出しの最大試行数。無限 retry を防ぐ。
// why: System 実測で decode_pcm（output audio is missing）が MaxAttempts=4 でも尽きることがある（run 33308282246）。
const MaxAttempts = 6

// 429 / 503 / 5xx 再試行の待機。System 実測で 1s 起点だと 429 が尽きる（run 33313682450）。
const (
	retryBackoffBase = 20 * time.Second
	retryBackoffMax  = 2 * time.Minute
)

// Gemini TTS が返す raw PCM の形式（公式 L16: 24 kHz / 16-bit / mono）。
// WAV wrap と decode の前提。HTTP 定数（ModelID 等）とは別責務。
const (
	pcmSampleRate = 24000
	pcmChannels   = 1
	pcmBitDepth   = 16
)
