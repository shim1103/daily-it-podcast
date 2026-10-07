package geminiapi

import "time"

const (
	// ModelID は原稿生成に使う Gemini model の SSoT。
	// why: 同社 TTS で使う GEMINI_API_KEY を流用でき、Gemini の中で最も安い現行 Flash-Lite。
	// 版更新はこの 1 行だけで閉じる（runtime の model 存在確認はしない。docs は値を写さず geminiapi.ModelID を参照する）。
	ModelID = "gemini-3.1-flash-lite"
	// EndpointURLTemplate は generateContent の URL template。%s は ModelID。
	EndpointURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
	// APIKeyHeader は API key を渡す HTTP header 名。
	// why: key を URL query に載せず header に載せる。error / log に URL を出しても key が漏れない。
	APIKeyHeader = "x-goog-api-key"
)

// ResponseBufferBytes は generateContent 応答（原稿 JSON を含む）を読む上限。
// why: 想定最大原稿の UTF-8 byte 数に、応答 envelope（candidates / usageMetadata 等）の余裕を足す。
const ResponseBufferBytes = 1 << 20

// why: 1 Write は invalid-draft retry で複数回 fetch しうるので、TTS 側の httpCallTimeout より長く取る。
const textWriterHTTPTimeout = 10 * time.Minute
