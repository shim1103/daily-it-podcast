package config

import (
	"errors"
	"strings"
	"testing"
)

// dummy* はformat制約（前後にwhitespaceなし・非空）だけ満たすtest用の値である。
const (
	dummyGetXAPIKey              = "getx-key"
	dummyCursorAPIKey            = "cursor-key"
	dummyGeminiAPIKey            = "gemini-key"
	dummyGoogleOAuthClientID     = "google-client-id"
	dummyGoogleOAuthClientSecret = "google-client-secret"
	dummyGoogleOAuthRefreshToken = "google-refresh-token"
	dummyDriveFolderID           = "drive-folder-id"
)

// fullValidEnv は7 key全てへ有効値を持つenv mapである。
func fullValidEnv() map[string]string {
	return map[string]string{
		GetXAPIKeyEnv:              dummyGetXAPIKey,
		CursorAPIKeyEnv:            dummyCursorAPIKey,
		GeminiAPIKeyEnv:            dummyGeminiAPIKey,
		GoogleOAuthClientIDEnv:     dummyGoogleOAuthClientID,
		GoogleOAuthClientSecretEnv: dummyGoogleOAuthClientSecret,
		GoogleOAuthRefreshTokenEnv: dummyGoogleOAuthRefreshToken,
		DriveFolderIDEnv:           dummyDriveFolderID,
	}
}

// lookupFrom はmapを入力源とするLookupEnvを返す。keyが無ければok=false。
func lookupFrom(env map[string]string) LookupEnv {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestLoad_returnsConfigMatchingContract_whenAllInputsValid(t *testing.T) {
	t.Parallel()

	// Given: 7 key全てへ有効値を持つenv
	env := fullValidEnv()

	// When: Loadする
	cfg, err := Load(lookupFrom(env))

	// Then: errorはなく、各fieldが投入値と一致する
	if err != nil {
		t.Fatal("Load() が有効入力でerrorを返した")
	}
	if cfg.Source.GetXAPIKey.Reveal() != dummyGetXAPIKey {
		t.Fatal("Source.GetXAPIKey が投入値と一致しない")
	}
	if cfg.Cursor.APIKey.Reveal() != dummyCursorAPIKey {
		t.Fatal("Cursor.APIKey が投入値と一致しない")
	}
	if cfg.Gemini.APIKey.Reveal() != dummyGeminiAPIKey {
		t.Fatal("Gemini.APIKey が投入値と一致しない")
	}
	if cfg.Drive.GoogleOAuthClientID != dummyGoogleOAuthClientID {
		t.Fatal("Drive.GoogleOAuthClientID が投入値と一致しない")
	}
	if cfg.Drive.GoogleOAuthClientSecret.Reveal() != dummyGoogleOAuthClientSecret {
		t.Fatal("Drive.GoogleOAuthClientSecret が投入値と一致しない")
	}
	if cfg.Drive.GoogleOAuthRefreshToken.Reveal() != dummyGoogleOAuthRefreshToken {
		t.Fatal("Drive.GoogleOAuthRefreshToken が投入値と一致しない")
	}
	if cfg.Drive.FolderID != dummyDriveFolderID {
		t.Fatal("Drive.FolderID が投入値と一致しない")
	}
}

func TestLoad_classifiesViolation_whenSingleKeyIsInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		mutate  func(env map[string]string) LookupEnv
		wantErr error
	}{
		{
			name: "missing_when_lookup_returns_not_ok",
			mutate: func(env map[string]string) LookupEnv {
				delete(env, GeminiAPIKeyEnv)
				return lookupFrom(env)
			},
			wantErr: ErrMissing,
		},
		{
			name: "empty_when_lookup_returns_ok_with_empty_string",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = ""
				return lookupFrom(env)
			},
			wantErr: ErrEmpty,
		},
		{
			name: "invalid_format_when_value_has_leading_whitespace",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = " gemini-key"
				return lookupFrom(env)
			},
			wantErr: ErrInvalidFormat,
		},
		{
			name: "invalid_format_when_value_has_trailing_tab",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = "gemini-key\t"
				return lookupFrom(env)
			},
			wantErr: ErrInvalidFormat,
		},
		{
			name: "invalid_format_when_value_has_trailing_newline",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = "gemini-key\n"
				return lookupFrom(env)
			},
			wantErr: ErrInvalidFormat,
		},
		{
			name: "invalid_format_when_value_has_leading_unicode_ideographic_space",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = "　gemini-key"
				return lookupFrom(env)
			},
			wantErr: ErrInvalidFormat,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: 1 keyだけを無効化したenv
			env := fullValidEnv()
			lookup := tc.mutate(env)

			// When: Loadする
			_, err := Load(lookup)

			// Then: 該当keyについて期待するsentinelでerrors.Isがtrueになる
			if err == nil {
				t.Fatal("Load() がviolationでerrorを返さなかった")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatal("errが期待するsentinelへ分類されなかった")
			}
			if !strings.Contains(err.Error(), GeminiAPIKeyEnv) {
				t.Fatal("err文字列へ違反keyが含まれなかった")
			}
		})
	}
}

func TestLoad_aggregatesViolationsInConfigFieldOrder_whenAllKeysMissing(t *testing.T) {
	t.Parallel()

	// Given: 全7 keyがmissingのenv
	lookup := lookupFrom(map[string]string{})

	// When: Loadする
	_, err := Load(lookup)

	// Then: Configのfield順で全7 keyが並ぶ
	if err == nil {
		t.Fatal("Load() がviolationでerrorを返さなかった")
	}
	lines := strings.Split(err.Error(), "\n")
	wantKeys := []string{
		GetXAPIKeyEnv,
		CursorAPIKeyEnv,
		GeminiAPIKeyEnv,
		GoogleOAuthClientIDEnv,
		GoogleOAuthClientSecretEnv,
		GoogleOAuthRefreshTokenEnv,
		DriveFolderIDEnv,
	}
	if len(lines) != len(wantKeys) {
		t.Fatal("集約された違反行数が7件ではない")
	}
	for i, key := range wantKeys {
		if !strings.HasPrefix(lines[i], key+": ") {
			t.Fatal("違反行の並びがConfigのfield順と一致しない")
		}
		if !strings.HasSuffix(lines[i], "missing") {
			t.Fatal("missingであるべき違反行の種別が異なる")
		}
	}
}

func TestLoad_aggregatesMixedViolationKinds_whenKeysFailDifferently(t *testing.T) {
	t.Parallel()

	// Given: 先頭keyがmissing、中央keyがempty、末尾keyがinvalid_formatのenv
	env := fullValidEnv()
	delete(env, GetXAPIKeyEnv)
	env[GeminiAPIKeyEnv] = ""
	env[DriveFolderIDEnv] = "drive-folder-id "

	// When: Loadする
	_, err := Load(lookupFrom(env))

	// Then: field順で3行、各行が対応する種別を持つ
	if err == nil {
		t.Fatal("Load() がviolationでerrorを返さなかった")
	}
	lines := strings.Split(err.Error(), "\n")
	want := []struct {
		key  string
		kind string
	}{
		{GetXAPIKeyEnv, "missing"},
		{GeminiAPIKeyEnv, "empty"},
		{DriveFolderIDEnv, "invalid_format"},
	}
	if len(lines) != len(want) {
		t.Fatal("集約された違反行数が3件ではない")
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w.key+": ") || !strings.HasSuffix(lines[i], w.kind) {
			t.Fatal("mixed violationの集約順または種別が期待と異なる")
		}
	}
	if !errors.Is(err, ErrMissing) || !errors.Is(err, ErrEmpty) || !errors.Is(err, ErrInvalidFormat) {
		t.Fatal("集約errが全種別のsentinelへ分類されなかった")
	}
}

func TestLoad_doesNotAutoTrim_whenValueIsSurroundedByWhitespace(t *testing.T) {
	t.Parallel()

	// Given: trimすれば有効になる値を持つenv
	env := fullValidEnv()
	env[CursorAPIKeyEnv] = "  cursor-key  "

	// When: Loadする
	cfg, err := Load(lookupFrom(env))

	// Then: auto-trimせずinvalid_formatとしてrejectし、有効値として受理しない
	if err == nil {
		t.Fatal("Load() がviolationでerrorを返さなかった")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatal("whitespace付き値がinvalid_formatへ分類されなかった")
	}
	if cfg.Cursor.APIKey != nil {
		t.Fatal("whitespace付き値がauto-trimされて受理された")
	}
}

func TestLoad_errorMessageDoesNotContainRawValue_whenValidationFails(t *testing.T) {
	t.Parallel()

	// Given: 秘匿価値のないdummyだが、raw値露出検査用に一意な文字列を含む値
	const rawMarker = "RAWMARKERdo-not-leak"
	env := fullValidEnv()
	env[GetXAPIKeyEnv] = rawMarker + " " // trailing whitespace で invalid_format

	// When: Loadする
	_, err := Load(lookupFrom(env))

	// Then: err文字列へraw値が含まれない
	//       （失敗時messageへrawMarkerを展開しないassertionで確認する）
	if err == nil {
		t.Fatal("Load() がviolationでerrorを返さなかった")
	}
	if strings.Contains(err.Error(), rawMarker) {
		t.Fatal("err文字列がraw値を含んだ")
	}
}

func TestLoad_returnsNoConfigValues_whenAnySingleViolationExists(t *testing.T) {
	t.Parallel()

	// Given: 1 keyだけ無効なenv
	env := fullValidEnv()
	delete(env, DriveFolderIDEnv)

	// When: Loadする
	cfg, err := Load(lookupFrom(env))

	// Then: errorを返し、Configは zero value（有効fieldも組み立てない）
	if err == nil {
		t.Fatal("violationがあるのにerrorを返さなかった")
	}
	if cfg.Source.GetXAPIKey != nil || cfg.Cursor.APIKey != nil || cfg.Gemini.APIKey != nil {
		t.Fatal("violation時にSecret fieldが組み立てられた")
	}
	if cfg.Drive.GoogleOAuthClientID != "" || cfg.Drive.FolderID != "" {
		t.Fatal("violation時にstring fieldが組み立てられた")
	}
}
