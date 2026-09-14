package config

// CursorConfig は原稿生成に必要なconfigである。
type CursorConfig struct {
	APIKey Secret
}

// GeminiConfig はGemini APIを使う機能に必要なconfigである。
//
// @invariant APIKeyは音声生成（TTS）が、SpareAPIKeyは原稿fallback（generateContent）が使う。
// 別keyに分けるのは、両者を同一free-tier枠へ相乗りさせるとfallback発火時に
// HTTP 429（rate limit）へ到達するため（docs/tasks/todo/generator-lane.md D表）。
type GeminiConfig struct {
	APIKey      Secret
	SpareAPIKey Secret
}

// DriveConfig はGoogle Drive保存に必要なconfigである。
type DriveConfig struct {
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret Secret
	GoogleOAuthRefreshToken Secret
	FolderID                string
}

// R2Config は Cloudflare R2（S3 互換）保存に必要なconfigである。
// Load の必須集合には含めない（現行本番正本は Drive。結線切替は列 6）。
type R2Config struct {
	AccessKeyID     Secret
	SecretAccessKey Secret
	AccountID       string
	Bucket          string
}

// Config はGeneratorがstartup時に確定するcapability別runtime configである。
//
// 全fieldが必須であること、およびvalidation violationの分類・集約順の契約はLoadを正とする。
//
// @invariant VariablesとSecretsの保存区分ではなくcapability単位でgroup化する。
type Config struct {
	Cursor CursorConfig
	Gemini GeminiConfig
	Drive  DriveConfig
}
