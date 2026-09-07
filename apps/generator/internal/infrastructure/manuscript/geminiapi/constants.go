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
const MaxAttempts = 4

// MaxRetryAfter は Retry-After header 由来の待ち時間の上限。異常値・DoS 回避。
const MaxRetryAfter = 30 * time.Second

// ResponseBufferBytes は generateContent 応答（原稿 JSON を含む）を読む上限。
// why: cursorapi.StreamBufferBytes と同根拠。想定最大原稿（約 5,000 字 × UTF-8 3 byte ≒ 15 KiB）に
//
//	Gemini 応答 envelope（candidates / usageMetadata 等）の余裕を足し 1 MiB を確保する。
const ResponseBufferBytes = 1 << 20
