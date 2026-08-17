package gemini

const (
	// why: 仕様決定前は置き場だけ固定する。model / voice / envelope / endpoint は後で埋める。
	ModelID          = ""
	VoiceName        = ""
	EnvelopePreamble = ""
	TranscriptLabel  = ""
	EndpointURL      = ""
)

// MaxAttempts は Gemini 呼び出しの最大試行数。無限 retry を防ぐ。
const MaxAttempts = 4
