package config

// SourceConfig はproduction情報源に必要なconfigである。
type SourceConfig struct {
	GetXAPIKey Secret
}

// CursorConfig は原稿生成に必要なconfigである。
type CursorConfig struct {
	APIKey Secret
}

// GeminiConfig は音声生成に必要なconfigである。
type GeminiConfig struct {
	APIKey Secret
}

// DriveConfig はGoogle Drive保存に必要なconfigである。
type DriveConfig struct {
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret Secret
	GoogleOAuthRefreshToken Secret
	FolderID                string
}

// Config はGeneratorがstartup時に確定するcapability別runtime configである。
//
// @ensure 全fieldは必須であり、先頭または末尾にwhitespaceを含む入力はinvalid_formatとしてrejectする。
// @ensure validation violationはConfigのfield順で全件集約される。
// @invariant VariablesとSecretsの保存区分ではなくcapability単位でgroup化する。
type Config struct {
	Source SourceConfig
	Cursor CursorConfig
	Gemini GeminiConfig
	Drive  DriveConfig
}
