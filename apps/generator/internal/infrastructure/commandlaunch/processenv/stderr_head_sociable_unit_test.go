package processenv

import (
	"bytes"
	"strings"
	"testing"
)

func TestStderrHead_returnsEmpty_whenStderrIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 空の child stderr
	stderr := []byte{}

	// When: 診断 head を取る
	got := stderrHead(stderr, "secret", nil)

	// Then: 空
	if got != "" {
		t.Fatalf("stderrHead() = %q, want empty", got)
	}
}

func TestStderrHead_returnsFullText_whenShorterThanLimitAndNoSecretOrStdin(t *testing.T) {
	t.Parallel()

	// Given: 上限未満で secret / stdin を含まない stderr
	stderr := []byte("agent failed")

	// When: 診断 head を取る
	got := stderrHead(stderr, "secret", []byte("brief"))

	// Then: 全文が残る
	if got != "agent failed" {
		t.Fatalf("stderrHead() = %q, want %q", got, "agent failed")
	}
}

func TestStderrHead_returnsFirstLimitBytes_whenLongerThanLimit(t *testing.T) {
	t.Parallel()

	// Given: 上限を超える stderr
	stderr := bytes.Repeat([]byte("a"), stderrHeadLimit+20)

	// When: 診断 head を取る
	got := stderrHead(stderr, "secret", nil)

	// Then: 先頭 300 byte だけ
	if len(got) != stderrHeadLimit {
		t.Fatalf("stderrHead() len = %d, want %d", len(got), stderrHeadLimit)
	}
	if got != strings.Repeat("a", stderrHeadLimit) {
		t.Fatalf("stderrHead() = %q, want first %d a", got, stderrHeadLimit)
	}
}

func TestStderrHead_returnsEmpty_whenStderrContainsSecret(t *testing.T) {
	t.Parallel()

	// Given: secret 値を含む stderr
	const secret = "processenv-test-dummy-secret-value"
	stderr := []byte("fail token=" + secret)

	// When: 診断 head を取る
	got := stderrHead(stderr, secret, nil)

	// Then: head 自体を省略する
	if got != "" {
		t.Fatalf("stderrHead() = %q, want empty when secret is present", got)
	}
}

func TestStderrHead_returnsEmpty_whenStderrContainsStdin(t *testing.T) {
	t.Parallel()

	// Given: stdin 本文を含む stderr
	stdin := []byte("processenv-test-stdin-token")
	stderr := []byte("echoed " + string(stdin))

	// When: 診断 head を取る
	got := stderrHead(stderr, "secret", stdin)

	// Then: head 自体を省略する
	if got != "" {
		t.Fatalf("stderrHead() = %q, want empty when stdin is present", got)
	}
}

func TestStderrHead_returnsHead_whenSecretValueIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 空 secret と非空 stderr
	stderr := []byte("agent failed")

	// When: 診断 head を取る
	got := stderrHead(stderr, "", nil)

	// Then: 空 secret は一致とみなさない
	if got != "agent failed" {
		t.Fatalf("stderrHead() = %q, want %q", got, "agent failed")
	}
}

func TestStderrHead_returnsHead_whenStdinIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 空 stdin と非空 stderr
	stderr := []byte("agent failed")

	// When: 診断 head を取る
	got := stderrHead(stderr, "secret", []byte{})

	// Then: 空 stdin は一致とみなさない
	if got != "agent failed" {
		t.Fatalf("stderrHead() = %q, want %q", got, "agent failed")
	}
}

func TestStderrHead_oneLinesNewlines_whenStderrHasNewline(t *testing.T) {
	t.Parallel()

	// Given: 改行を含む stderr
	stderr := []byte("first\nsecond")

	// When: 診断 head を取る
	got := stderrHead(stderr, "secret", nil)

	// Then: 1行化される
	if got != "first second" {
		t.Fatalf("stderrHead() = %q, want %q", got, "first second")
	}
}

func TestWithStderrHead_wrapsCause_whenHeadIsNonEmpty(t *testing.T) {
	t.Parallel()

	// Given: 非空 head になる stderr と元 error
	base := errString("exit status 1")
	stderr := []byte("agent failed")

	// When: cause へ head を載せる
	got := withStderrHead(base, stderr, "secret", nil)

	// Then: 元 error を保持し head を診断へ載せる
	if got == nil {
		t.Fatal("withStderrHead() = nil, want wrapped error")
	}
	if !strings.Contains(got.Error(), "stderr_head=agent failed") {
		t.Fatalf("withStderrHead() = %q, want stderr_head", got.Error())
	}
	if !strings.Contains(got.Error(), base.Error()) {
		t.Fatalf("withStderrHead() = %q, want base %q", got.Error(), base.Error())
	}
}

func TestWithStderrHead_returnsBaseOnly_whenHeadIsOmittedForSecret(t *testing.T) {
	t.Parallel()

	// Given: secret を含む stderr
	const secret = "processenv-test-dummy-secret-value"
	base := errString("exit status 1")
	stderr := []byte("token=" + secret)

	// When: cause へ head を載せようとする
	got := withStderrHead(base, stderr, secret, nil)

	// Then: 元 error のまま（secret を挿入しない）
	if got != base {
		t.Fatalf("withStderrHead() = %v, want the base error", got)
	}
	if strings.Contains(got.Error(), secret) {
		t.Fatalf("withStderrHead() = %q, contains secret", got.Error())
	}
}

func TestWithStderrHead_returnsBaseOnly_whenHeadIsOmittedForStdin(t *testing.T) {
	t.Parallel()

	// Given: stdin 本文を含む stderr
	stdin := []byte("processenv-test-stdin-token")
	base := errString("exit status 1")
	stderr := []byte("echoed " + string(stdin))

	// When: cause へ head を載せようとする
	got := withStderrHead(base, stderr, "secret", stdin)

	// Then: 元 error のまま（stdin を挿入しない）
	if got != base {
		t.Fatalf("withStderrHead() = %v, want the base error", got)
	}
	if strings.Contains(got.Error(), string(stdin)) {
		t.Fatalf("withStderrHead() = %q, contains stdin", got.Error())
	}
}

type errString string

func (e errString) Error() string { return string(e) }
