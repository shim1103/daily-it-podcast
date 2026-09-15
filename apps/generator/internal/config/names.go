// Package config はGeneratorのprocess environment configuration boundary契約を提供する。
package config

const (
	CursorAPIKeyEnv            = "CURSOR_API_KEY"
	GeminiAPIKeyEnv            = "GEMINI_API_KEY"
	SpareGeminiAPIKeyEnv       = "SPARE_GEMINI_API_KEY"
	GoogleOAuthClientIDEnv     = "GOOGLE_OAUTH_CLIENT_ID"
	GoogleOAuthClientSecretEnv = "GOOGLE_OAUTH_CLIENT_SECRET"
	GoogleOAuthRefreshTokenEnv = "GOOGLE_OAUTH_REFRESH_TOKEN"
	DriveFolderIDEnv           = "DRIVE_FOLDER_ID"

	// R2（generator S3 互換）。Load 必須。本番 write/lookup 結線の切替は列 6。
	R2AccessKeyIDEnv     = "R2_ACCESS_KEY_ID"
	R2SecretAccessKeyEnv = "R2_SECRET_ACCESS_KEY"
	R2AccountIDEnv       = "R2_ACCOUNT_ID"
	R2BucketEnv          = "R2_BUCKET"
)
