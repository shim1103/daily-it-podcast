package cursorapi

import "time"

const (
	// APIBaseURL は Cursor Cloud Agents API の base URL。
	APIBaseURL = "https://api.cursor.com"
	// AgentsPath は agent を create する API path。
	AgentsPath = "/v1/agents"
	// StreamPathTemplate は run の SSE stream path。%s は APIBaseURL+AgentsPath / agentId / runId。
	StreamPathTemplate = "%s/%s/runs/%s/stream"
	// RunsPathTemplate は既存 agent へ follow-up run を送る API path。%s は APIBaseURL+AgentsPath / agentId。
	RunsPathTemplate = "%s/%s/runs"

	// ModelID は Cursor Cloud Agents へ指定する model。
	ModelID = "composer-2.5"

	// AuthorizationHeader は API key を渡す HTTP header 名。
	AuthorizationHeader = "Authorization"
	// BearerTokenPrefix は Authorization header の Bearer scheme。
	BearerTokenPrefix = "Bearer "
)

// MaxAttempts は Cursor の 429 応答に対する最大試行数。無限 retry を防ぐ。
const MaxAttempts = 4

// TextWriterMaxAttempts は ManuscriptDraft 検証失敗（invalid-draft）時の Write 内部 retry 上限。
// geminiapi.TextWriterMaxAttempts と同値（Decision 2026-09-16T11-41-26 §1-6：値は据え置く）。
const TextWriterMaxAttempts = 5

// MaxRetryAfter は Retry-After header 由来の待ち時間の上限。
// why: 異常値・DoS 回避。run 全体上限は GHA job / process cancel に委ねる（Decision §6）。
const MaxRetryAfter = 30 * time.Second

// StreamBufferBytes は SSE 1 行（終端 result event の JSON）を読む scanner buffer 上限。
// why: result event は 1 行 JSON。想定最大原稿（約 5,000 字 × UTF-8 3 byte ≒ 15 KiB）に
//
//	runId / status / durationMs / git 等の envelope 余裕を足し、bufio 既定 64 KiB では足りない
//	長文原稿でも切れないよう 1 MiB を確保する。
const StreamBufferBytes = 1 << 20
