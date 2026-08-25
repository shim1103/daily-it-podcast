package secrettransport_test

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

func TestValidateAbsoluteHTTPSURL_returnsNil_whenURLIsAbsoluteHTTPSWithHost(t *testing.T) {
	t.Parallel()

	// Given: host を持つ絶対 https URL
	targetURL := "https://example.test/upstream"

	// When: ValidateAbsoluteHTTPSURL する
	err := secrettransport.ValidateAbsoluteHTTPSURL(targetURL)

	// Then: error は nil
	if err != nil {
		t.Fatalf("ValidateAbsoluteHTTPSURL() error = %v, want nil", err)
	}
}

func TestValidateAbsoluteHTTPSURL_returnsError_whenURLIsBlank(t *testing.T) {
	t.Parallel()

	// Given: 空白のみの URL
	targetURL := "   "

	// When: ValidateAbsoluteHTTPSURL する
	err := secrettransport.ValidateAbsoluteHTTPSURL(targetURL)

	// Then: error を返す
	if err == nil {
		t.Fatal("ValidateAbsoluteHTTPSURL() error = nil, want non-nil")
	}
}

func TestValidateAbsoluteHTTPSURL_returnsError_whenSchemeIsNotHTTPS(t *testing.T) {
	t.Parallel()

	// Given: scheme が https ではない URL
	targetURL := "http://example.test/upstream"

	// When: ValidateAbsoluteHTTPSURL する
	err := secrettransport.ValidateAbsoluteHTTPSURL(targetURL)

	// Then: error を返す
	if err == nil {
		t.Fatal("ValidateAbsoluteHTTPSURL() error = nil, want non-nil")
	}
}

func TestValidateAbsoluteHTTPSURL_returnsError_whenHostIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: scheme は https だが host が空の URL
	targetURL := "https://"

	// When: ValidateAbsoluteHTTPSURL する
	err := secrettransport.ValidateAbsoluteHTTPSURL(targetURL)

	// Then: error を返す
	if err == nil {
		t.Fatal("ValidateAbsoluteHTTPSURL() error = nil, want non-nil")
	}
}

func TestValidateAbsoluteHTTPSURL_returnsError_whenURLIsUnparseable(t *testing.T) {
	t.Parallel()

	// Given: パース不能な URL（制御文字を含む）
	targetURL := "https://example.test/\x7f"

	// When: ValidateAbsoluteHTTPSURL する
	err := secrettransport.ValidateAbsoluteHTTPSURL(targetURL)

	// Then: error を返す
	if err == nil {
		t.Fatal("ValidateAbsoluteHTTPSURL() error = nil, want non-nil")
	}
}
