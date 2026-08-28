package config

import (
	"errors"
	"fmt"
	"strings"
)

// configField はConfigの1 fieldに対応する検証対象environment keyの位置である。
//
// 宣言順はConfigのfield順（Source.GetXAPIKey → Cursor.APIKey → Gemini.APIKey
// → Drive.GoogleOAuthClientID → Drive.GoogleOAuthClientSecret
// → Drive.GoogleOAuthRefreshToken → Drive.FolderID）と一致する。
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

// configFieldKeys はconfigFieldの宣言順に並んだenvironment keyの表である。
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
// @ensure 1 fieldでも違反があれば、Configのfield順で全違反をerrors.Joinしたerrorを返し、Configはzero valueを返す。
// @invariant raw runtime値をerrorへ含めない。
func Load(lookup LookupEnv) (Config, error) {
	var values [configFieldCount]string
	var errs []error
	for field, key := range configFieldKeys {
		value, err := validateEnvValue(lookup, key)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
			continue
		}
		values[field] = value
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
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

// validateEnvValue は1 keyのlookup結果を検証する。
//
// keyが未定義ならErrMissing、定義済みだが空文字ならErrEmpty、
// 先頭または末尾にwhitespaceを含むならErrInvalidFormatを返す。
// それ以外は(value, nil)を返し、値をauto-trimしない。
func validateEnvValue(lookup LookupEnv, key string) (string, error) {
	value, ok := lookup(key)
	if !ok {
		return "", ErrMissing
	}
	if value == "" {
		return "", ErrEmpty
	}
	if hasSurroundingWhitespace(value) {
		return "", ErrInvalidFormat
	}
	return value, nil
}

// hasSurroundingWhitespace は先頭または末尾にwhitespaceを持つかを判定する。
//
// trimして長さが変わる入力をwhitespace付きと判定する。判定のみで値は変更しない。
func hasSurroundingWhitespace(value string) bool {
	return strings.TrimSpace(value) != value
}
