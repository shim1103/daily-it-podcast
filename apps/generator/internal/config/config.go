package config

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
// 全fieldが必須であること、およびvalidation violationの分類・集約順の契約はLoadを正とする。
//
// @invariant VariablesとSecretsの保存区分ではなくcapability単位でgroup化する。
type Config struct {
	Cursor CursorConfig
	Gemini GeminiConfig
	Drive  DriveConfig
}
