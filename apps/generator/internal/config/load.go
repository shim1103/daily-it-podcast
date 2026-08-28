package config

import (
	"strings"
)

// configField は Config の field 順（Source.GetXAPIKey → Cursor.APIKey →
// Gemini.APIKey → Drive.GoogleOAuthClientID → Drive.GoogleOAuthClientSecret →
// Drive.GoogleOAuthRefreshToken → Drive.FolderID）に対応する。
type configField int

const (
	fieldGetXAPIKey configField = iota
	fieldCursorAPIKey
	fieldGeminiAPIKey
	fieldGoogleOAuthClientID
	fieldGoogleOAuthClientSecret
	fieldGoogleOAuthRefreshToken
	fieldDriveFolderID
	configFieldCount
)

var configFieldKeys = [configFieldCount]string{
	fieldGetXAPIKey:              GetXAPIKeyEnv,
	fieldCursorAPIKey:            CursorAPIKeyEnv,
	fieldGeminiAPIKey:            GeminiAPIKeyEnv,
	fieldGoogleOAuthClientID:     GoogleOAuthClientIDEnv,
	fieldGoogleOAuthClientSecret: GoogleOAuthClientSecretEnv,
	fieldGoogleOAuthRefreshToken: GoogleOAuthRefreshTokenEnv,
	fieldDriveFolderID:           DriveFolderIDEnv,
}

// Load はprocess environmentを一度だけ読み、検証済みConfigを構築する。
//
// @require lookupは注入されたenvironment参照だけを入力源にし、dotenv fileをloadしない。
// @ensure 全fieldが有効な時だけ、Config型の全fieldを満たしたConfigを返す。
// @ensure 1 fieldでも違反があれば、Configのfield順で全違反を束ねた *Errors を返し、Configはzero valueを返す。
// @invariant raw runtime値をerrorへ含めない。
func Load(lookup LookupEnv) (Config, error) {
	var values [configFieldCount]string
	var violations []*Error
	for field, key := range configFieldKeys {
		value, kind := validateEnvValue(lookup, key)
		if kind != "" {
			violations = append(violations, configErr(key, kind))
			continue
		}
		values[field] = value
	}

	if len(violations) > 0 {
		return Config{}, &Errors{Violations: violations}
	}

	return Config{
		Source: SourceConfig{
			GetXAPIKey: newSecret(values[fieldGetXAPIKey]),
		},
		Cursor: CursorConfig{
			APIKey: newSecret(values[fieldCursorAPIKey]),
		},
		Gemini: GeminiConfig{
			APIKey: newSecret(values[fieldGeminiAPIKey]),
		},
		Drive: DriveConfig{
			GoogleOAuthClientID:     values[fieldGoogleOAuthClientID],
			GoogleOAuthClientSecret: newSecret(values[fieldGoogleOAuthClientSecret]),
			GoogleOAuthRefreshToken: newSecret(values[fieldGoogleOAuthRefreshToken]),
			FolderID:                values[fieldDriveFolderID],
		},
	}, nil
}

// validateEnvValue は1 key の lookup 結果を検証し、違反があれば Kind 文字列を、
// valid なら (value, "") を返す。値は auto-trim しない。
func validateEnvValue(lookup LookupEnv, key string) (string, string) {
	value, ok := lookup(key)
	if !ok {
		return "", KindMissing
	}
	if value == "" {
		return "", KindEmpty
	}
	if strings.TrimSpace(value) != value {
		return "", KindInvalidFormat
	}
	return value, ""
}
