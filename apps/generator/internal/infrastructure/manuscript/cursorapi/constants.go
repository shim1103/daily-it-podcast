package cursorapi

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

// StreamBufferBytes は SSE 1 行（終端 result event の JSON）を読む scanner buffer 上限。
// why: result event は 1 行 JSON で、長文原稿と envelope を足すと bufio 既定の上限を超えうる。
const StreamBufferBytes = 1 << 20

const (
	// what: Cursor が利用枠の喪失を 400 の body で示す文字列。
	usageLimitExceededCode = "usage_limit_exceeded"
	scanBufferInitialBytes = 64 * 1024
)
