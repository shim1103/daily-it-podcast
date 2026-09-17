package config

import (
	"errors"
	"strings"
	"testing"
)

// dummy* はformat制約（前後にwhitespaceなし・非空）だけ満たすtest用の値である。
const (
	dummyCursorAPIKey      = "cursor-key"
	dummyGeminiAPIKey      = "gemini-key"
	dummySpareGeminiAPIKey = "spare-gemini-key"
	dummyR2AccessKeyID     = "r2-access-key"
	dummyR2SecretAccessKey = "r2-secret-access-key"
	dummyR2AccountID       = "r2-account-id"
	dummyR2Bucket          = "r2-bucket"
)

// fullValidEnv は全 key へ有効値を持つenv mapである。
func fullValidEnv() map[string]string {
	return map[string]string{
		CursorAPIKeyEnv:      dummyCursorAPIKey,
		GeminiAPIKeyEnv:      dummyGeminiAPIKey,
		SpareGeminiAPIKeyEnv: dummySpareGeminiAPIKey,
		R2AccessKeyIDEnv:     dummyR2AccessKeyID,
		R2SecretAccessKeyEnv: dummyR2SecretAccessKey,
		R2AccountIDEnv:       dummyR2AccountID,
		R2BucketEnv:          dummyR2Bucket,
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

	// Given: 全 key へ有効値を持つenv
	env := fullValidEnv()

	// When: Loadする
	cfg, err := Load(lookupFrom(env))

	// Then: errorはなく、各fieldが投入値と一致する
	if err != nil {
		t.Fatal("Load() が有効入力でerrorを返した")
	}
	if cfg.Cursor.APIKey.Reveal() != dummyCursorAPIKey {
		t.Fatal("Cursor.APIKey が投入値と一致しない")
	}
	if cfg.Gemini.APIKey.Reveal() != dummyGeminiAPIKey {
		t.Fatal("Gemini.APIKey が投入値と一致しない")
	}
	if cfg.Gemini.SpareAPIKey.Reveal() != dummySpareGeminiAPIKey {
		t.Fatal("Gemini.SpareAPIKey が投入値と一致しない")
	}
	if cfg.R2.AccessKeyID.Reveal() != dummyR2AccessKeyID {
		t.Fatal("R2.AccessKeyID が投入値と一致しない")
	}
	if cfg.R2.SecretAccessKey.Reveal() != dummyR2SecretAccessKey {
		t.Fatal("R2.SecretAccessKey が投入値と一致しない")
	}
	if cfg.R2.AccountID != dummyR2AccountID {
		t.Fatal("R2.AccountID が投入値と一致しない")
	}
	if cfg.R2.Bucket != dummyR2Bucket {
		t.Fatal("R2.Bucket が投入値と一致しない")
	}
}

func TestLoad_classifiesViolation_whenSingleKeyIsInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		mutate   func(env map[string]string) LookupEnv
		wantKind string
	}{
		{
			name: "missing_when_lookup_returns_not_ok",
			mutate: func(env map[string]string) LookupEnv {
				delete(env, GeminiAPIKeyEnv)
				return lookupFrom(env)
			},
			wantKind: KindMissing,
		},
		{
			name: "empty_when_lookup_returns_ok_with_empty_string",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = ""
				return lookupFrom(env)
			},
			wantKind: KindEmpty,
		},
		{
			name: "invalid_format_when_value_has_leading_whitespace",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = " gemini-key"
				return lookupFrom(env)
			},
			wantKind: KindInvalidFormat,
		},
		{
			name: "invalid_format_when_value_has_trailing_tab",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = "gemini-key\t"
				return lookupFrom(env)
			},
			wantKind: KindInvalidFormat,
		},
		{
			name: "invalid_format_when_value_has_trailing_newline",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = "gemini-key\n"
				return lookupFrom(env)
			},
			wantKind: KindInvalidFormat,
		},
		{
			name: "invalid_format_when_value_has_leading_unicode_ideographic_space",
			mutate: func(env map[string]string) LookupEnv {
				env[GeminiAPIKeyEnv] = "　gemini-key"
				return lookupFrom(env)
			},
			wantKind: KindInvalidFormat,
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

			// Then: 該当keyについて期待するKindへ分類される
			if err == nil {
				t.Fatal("Load() がviolationでerrorを返さなかった")
			}
			var single *Error
			if !errors.As(err, &single) {
				t.Fatal("errから *config.Error を取り出せなかった")
			}
			if single.Key != GeminiAPIKeyEnv {
				t.Fatalf("Key = %q", single.Key)
			}
			if single.Kind != tc.wantKind {
				t.Fatalf("Kind = %q, want %q", single.Kind, tc.wantKind)
			}
		})
	}
}

func TestLoad_classifiesR2Violation_whenSingleR2KeyIsInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		mutate   func(env map[string]string)
		wantKey  string
		wantKind string
	}{
		{
			name: "missing_access_key",
			mutate: func(env map[string]string) {
				delete(env, R2AccessKeyIDEnv)
			},
			wantKey:  R2AccessKeyIDEnv,
			wantKind: KindMissing,
		},
		{
			name: "empty_bucket",
			mutate: func(env map[string]string) {
				env[R2BucketEnv] = ""
			},
			wantKey:  R2BucketEnv,
			wantKind: KindEmpty,
		},
		{
			name: "invalid_format_account_id",
			mutate: func(env map[string]string) {
				env[R2AccountIDEnv] = " account "
			},
			wantKey:  R2AccountIDEnv,
			wantKind: KindInvalidFormat,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: R2 の 1 key だけを無効化した env
			env := fullValidEnv()
			tc.mutate(env)

			// When: Loadする
			_, err := Load(lookupFrom(env))

			// Then: 該当 R2 key について期待する Kind へ分類され、secret 実値は Error() に出ない
			if err == nil {
				t.Fatal("Load() がviolationでerrorを返さなかった")
			}
			var bundled *Errors
			if !errors.As(err, &bundled) {
				t.Fatalf("error type %T, want *Errors", err)
			}
			if len(bundled.Violations) != 1 {
				t.Fatalf("violations = %d, want 1", len(bundled.Violations))
			}
			if bundled.Violations[0].Key != tc.wantKey || bundled.Violations[0].Kind != tc.wantKind {
				t.Fatalf("violation = %+v, want key=%s kind=%s", bundled.Violations[0], tc.wantKey, tc.wantKind)
			}
			if strings.Contains(err.Error(), dummyR2SecretAccessKey) {
				t.Fatalf("Error() に secret 実値: %q", err.Error())
			}
		})
	}
}

func TestLoad_aggregatesViolationsInConfigFieldOrder_whenAllKeysMissing(t *testing.T) {
	t.Parallel()

	// Given: 全 keyがmissingのenv
	lookup := lookupFrom(map[string]string{})

	// When: Loadする
	_, err := Load(lookup)

	// Then: Configのfield順で全 keyが並ぶ
	if err == nil {
		t.Fatal("Load() がviolationでerrorを返さなかった")
	}
	lines := strings.Split(err.Error(), "\n")
	wantKeys := []string{
		CursorAPIKeyEnv,
		GeminiAPIKeyEnv,
		SpareGeminiAPIKeyEnv,
		R2AccessKeyIDEnv,
		R2SecretAccessKeyEnv,
		R2AccountIDEnv,
		R2BucketEnv,
	}
	if len(lines) != len(wantKeys) {
		t.Fatal("集約された違反行数がkey数と一致しない")
	}
	for i, key := range wantKeys {
		// 各行は "<prefix>: <key>: <kind>" の3段パターン（Domain/Infra と対称）
		if !strings.HasPrefix(lines[i], "generator config: "+key+": ") {
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
	delete(env, CursorAPIKeyEnv)
	env[GeminiAPIKeyEnv] = ""
	env[R2BucketEnv] = "r2-bucket "

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
		{CursorAPIKeyEnv, "missing"},
		{GeminiAPIKeyEnv, "empty"},
		{R2BucketEnv, "invalid_format"},
	}
	if len(lines) != len(want) {
		t.Fatal("集約された違反行数が3件ではない")
	}
	for i, w := range want {
		// 各行は "<prefix>: <key>: <kind>" の3段パターン
		if lines[i] != "generator config: "+w.key+": "+w.kind {
			t.Fatalf("行 %d = %q", i, lines[i])
		}
	}
}

func TestLoad_reachesEachViolationViaErrorsAs_whenKeysFailDifferently(t *testing.T) {
	t.Parallel()

	// Given: 先頭keyがmissing、末尾keyがinvalid_formatのenv
	env := fullValidEnv()
	delete(env, CursorAPIKeyEnv)
	env[R2BucketEnv] = "r2-bucket "

	// When: Loadする
	_, err := Load(lookupFrom(env))

	// Then: 束ねたerrorから個別の *config.Error へ errors.As で到達でき、
	//       その Key と Kind が読める
	if err == nil {
		t.Fatal("Load() がviolationでerrorを返さなかった")
	}
	var single *Error
	if !errors.As(err, &single) {
		t.Fatal("束ねたerrorから *config.Error へ errors.As で到達できなかった")
	}
	if single.Key != CursorAPIKeyEnv {
		t.Fatalf("errors.As が最初の違反 *config.Error を返さなかった: Key = %q", single.Key)
	}
	if single.Kind != KindMissing {
		t.Fatalf("Kind = %q, want %q", single.Kind, KindMissing)
	}
	// 単体 *Error も "<prefix>: <key>: <kind>" の3段パターン（Domain/Infra と対称）
	if got := single.Error(); got != "generator config: "+CursorAPIKeyEnv+": missing" {
		t.Fatalf("single.Error() = %q", got)
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
	var single *Error
	if !errors.As(err, &single) || single.Kind != KindInvalidFormat {
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
	env[CursorAPIKeyEnv] = rawMarker + " " // trailing whitespace で invalid_format

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
	delete(env, R2BucketEnv)

	// When: Loadする
	cfg, err := Load(lookupFrom(env))

	// Then: errorを返し、Configは zero value（有効fieldも組み立てない）
	if err == nil {
		t.Fatal("violationがあるのにerrorを返さなかった")
	}
	if cfg.Cursor.APIKey != nil || cfg.Gemini.APIKey != nil || cfg.Gemini.SpareAPIKey != nil {
		t.Fatal("violation時にSecret fieldが組み立てられた")
	}
	if cfg.R2.AccessKeyID != nil || cfg.R2.SecretAccessKey != nil || cfg.R2.AccountID != "" || cfg.R2.Bucket != "" {
		t.Fatal("violation時に R2 fieldが組み立てられた")
	}
}
