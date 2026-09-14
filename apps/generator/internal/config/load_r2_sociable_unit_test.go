package config

import (
	"errors"
	"strings"
	"testing"
)

const (
	dummyR2AccessKeyID     = "r2-access-key"
	dummyR2SecretAccessKey = "r2-secret-access-key"
	dummyR2AccountID       = "r2-account-id"
	dummyR2Bucket          = "r2-bucket"
)

func fullValidR2Env() map[string]string {
	return map[string]string{
		R2AccessKeyIDEnv:     dummyR2AccessKeyID,
		R2SecretAccessKeyEnv: dummyR2SecretAccessKey,
		R2AccountIDEnv:       dummyR2AccountID,
		R2BucketEnv:          dummyR2Bucket,
	}
}

func TestLoadR2_returnsConfigMatchingContract_whenAllInputsValid(t *testing.T) {
	t.Parallel()

	cfg, err := LoadR2(lookupFrom(fullValidR2Env()))
	if err != nil {
		t.Fatal("LoadR2() が有効入力でerrorを返した")
	}
	if cfg.AccessKeyID.Reveal() != dummyR2AccessKeyID {
		t.Fatal("AccessKeyID が投入値と一致しない")
	}
	if cfg.SecretAccessKey.Reveal() != dummyR2SecretAccessKey {
		t.Fatal("SecretAccessKey が投入値と一致しない")
	}
	if cfg.AccountID != dummyR2AccountID {
		t.Fatal("AccountID が投入値と一致しない")
	}
	if cfg.Bucket != dummyR2Bucket {
		t.Fatal("Bucket が投入値と一致しない")
	}
}

func TestLoadR2_classifiesViolation_whenSingleKeyIsInvalid(t *testing.T) {
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
			env := fullValidR2Env()
			tc.mutate(env)
			_, err := LoadR2(lookupFrom(env))
			if err == nil {
				t.Fatal("expected error")
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

func TestLoad_doesNotRequireR2Env(t *testing.T) {
	t.Parallel()

	// Given: 現行 Load 必須だけ（R2 なし）
	env := fullValidEnv()

	// When
	cfg, err := Load(lookupFrom(env))

	// Then: Drive 正本 Load は R2 無しでも成功
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Drive.FolderID != dummyDriveFolderID {
		t.Fatal("Drive.FolderID mismatch")
	}
}
