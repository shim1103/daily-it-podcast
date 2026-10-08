package gemini

import "time"

const (
	// why: 公式 Interactions TTS 3.8 Flash。変更は Adapter 定数だけで閉じる。
	ModelID     = "gemini-3.8-flash-tts"
	VoiceName   = "Charon"
	EndpointURL = "https://generativelanguage.googleapis.com/v1beta/interactions"
)

const (
	geminiAPIKeyHeader      = "x-goog-api-key"
	prohibitedContentMarker = "PROHIBITED_CONTENT"
)

// what: 公式 API errors ページが定める error.code（snake_case の機械可読 code）。
const (
	errorCodeQuotaExceeded     = "quota_exceeded"
	errorCodeRateLimitExceeded = "rate_limit_exceeded"
)

// MaxAttempts は 1 セグメントが連続で消費してよい Gemini 呼び出しの上限。無限 retry を防ぐ。
// why: 1 セグメントの暴走で SynthesizeBudget 全部を食わせない二段構えの内側（Decision 2026-09-02T13-56-00）。
// 同種失敗の連続は maxConsecutiveSameOp で先に打ち切るので、この上限まで回るのは Op が入れ替わる失敗だけ。
// RPM とは独立した閾値のため、callGap 短縮（Decision 2026-09-16T11-41-59）でも値は据え置く。
const MaxAttempts = 3

// why: 再試行できる失敗が同種のまま続くのは、その本文に対して決定論的に失敗しているとみなすため（Decision 2026-09-02T13-56-00）。
// 「同種」の判定は sameGeminiOp が持つ。
const maxConsecutiveSameOp = 2

// SynthesizeBudget は TierFree での 1 度の SynthesizeAll 呼び出し全体で許す Gemini 呼び出しの合計上限。
// why: AI Studio 実測 RPD=10（gemini-3.8-flash-tts）を 1 episode で焼き切らないため、
// セグメント単位の MaxAttempts ではなく呼び出し群の合計で絞る（Decision 2026-09-16T11-41-59）。
const SynthesizeBudget = 10

// SynthesizeBudgetPaid は TierPaid での SynthesizeBudget。
// why: paid tier の正確な RPM/RPD は Google 公式ドキュメント上に静的な数値が存在せず未実測。
// 保守的な暫定値として free の 2 倍に留める。実測後に見直すこと。
const SynthesizeBudgetPaid = SynthesizeBudget * 2

// 再試行（回復の明示がある 429・5xx・通信断・音声欠落）の待機の既定値。
// why: 20s 起点でも System で 429 の再試行を使い切った（run 33314746860, ~476s）。60s 起点・上限 3m へ。
// why: rate 計測は SpeechSynthesizer の field へ注入して差し替える（Decision 2026-09-03T14-46-00）。
// 既定 constructor（NewSpeechSynthesizer）はこの const 値を使うので挙動は不変。
const (
	defaultRetryBackoffBase = 60 * time.Second
	defaultRetryBackoffMax  = 3 * time.Minute
)

// defaultCallGap は client.Do どうしの最小間隔（成功・失敗を問わない）の既定値。
// why: AI Studio 実測 RPM=10 を根拠に callGap = 60/RPM = 6s へ
// 改める（Decision 2026-09-16T11-41-59）。旧 20s は無料枠 3 RPM 前提（Decision 2026-09-02T13-56-00）
// だったが、実測値と乖離していたため式ごと差し替える。RPM が変われば 60/RPM を計算し直すこと。
const defaultCallGap = 6 * time.Second

// httpCallTimeout は Gemini TTS 1 呼び出しの Client 全体 timeout である。
// why: 120s でも長文朗読で awaiting headers が切れた（run 33310692613）。
// Composition から渡る *http.Client は全体 timeout を持たないので、この値は Adapter が付け直す。
const httpCallTimeout = 5 * time.Minute

// minSpeechDurationSec は「実質無音でない」とみなす朗読音声の最小尺（秒）。
const minSpeechDurationSec = 0.5
