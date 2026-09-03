package cursorapi

const (
	// APIBaseURL は Cursor Cloud Agents API の base URL。
	APIBaseURL = "https://api.cursor.com"
	// AgentsPath は agent を create する API path。
	AgentsPath = "/v1/agents"

	// ModelID は Cursor Cloud Agents へ指定する model。
	ModelID = "composer-2.5"

	// AuthorizationHeader は API key を渡す HTTP header 名。
	AuthorizationHeader = "Authorization"
	// BearerTokenPrefix は Authorization header の Bearer scheme。
	BearerTokenPrefix = "Bearer "
)

// MaxAttempts は Cursor の 429 応答に対する最大試行数。無限 retry を防ぐ。
const MaxAttempts = 4
