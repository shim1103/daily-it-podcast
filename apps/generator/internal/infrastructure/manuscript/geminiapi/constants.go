package geminiapi

import "time"

const (
	// ModelID は原稿生成に使う Gemini model の SSOT。Cursor の利用枠喪失時の fallback 先。
	// why: 同社 TTS で使う GEMINI_API_KEY を流用でき、Gemini の中で最も安い現行 Flash-Lite。
	//      版更新はこの 1 行だけで閉じる（runtime の model 存在確認はしない。docs は値を写さず
	//      geminiapi.ModelID を参照する）。旧 gemini-2.5-flash-lite の base alias は v1beta の
	//      generateContent で 404 になったため（smoke run 34129174375）現行 GA 版へ更新した。
	ModelID = "gemini-3.1-flash-lite"
	// EndpointURLTemplate は generateContent の URL template。%s は ModelID。
	EndpointURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
	// APIKeyHeader は API key を渡す HTTP header 名。
	// why: key を URL query に載せず header に載せる。error / log に URL を出しても key が漏れない。
	APIKeyHeader = "x-goog-api-key"
)

// MaxAttempts は 429 応答に対する最大試行数。無限 retry を防ぐ。cursorapi と同値。
// why: TextWriter 用 model の AI Studio 実測 RPD=100〜500（free）は現状の 1 日 1 回 produce 運用
//
//	では枯渇しにくく、gemini（TTS）の RPD=10 のような「1 episode で焼き切る」動機が無い。
//	tier による値の調整は行わず、この値（および TextWriterMaxAttempts）に
//	tier ごとの差分は付けない。
const MaxAttempts = 4

// TextWriterMaxAttempts は ManuscriptDraft 検証失敗（invalid-draft）時の Write 内部 retry 上限。
// LLM 出力の rune 数・topic 数揺れを吸収する。429 応答用の MaxAttempts とは別物。
// why: application/produce_episode.go にあった旧値をそのまま維持する（Decision 2026-09-16T11-41-26 §1-6）。
const TextWriterMaxAttempts = 5

// textWriterHTTPTimeout は generateContent 呼び出し全体（invalid-draft retry を含む Write 1 回）の
// *http.Client timeout である。
// why: TTS 側 httpCallTimeout（5分。実測で 120s でも長文朗読が timeout した経緯: run 33310692613）を
//
//	参考に、TextWriter は将来複数回 fetch を含みうる分だけ長く 10 分を確保する。
const textWriterHTTPTimeout = 10 * time.Minute

// MaxRetryAfter は Retry-After header 由来の待ち時間の上限。異常値・DoS 回避。
const MaxRetryAfter = 30 * time.Second

// ResponseBufferBytes は generateContent 応答（原稿 JSON を含む）を読む上限。
// why: cursorapi.StreamBufferBytes と同根拠。想定最大原稿（約 5,000 字 × UTF-8 3 byte ≒ 15 KiB）に
//
//	Gemini 応答 envelope（candidates / usageMetadata 等）の余裕を足し 1 MiB を確保する。
const ResponseBufferBytes = 1 << 20
