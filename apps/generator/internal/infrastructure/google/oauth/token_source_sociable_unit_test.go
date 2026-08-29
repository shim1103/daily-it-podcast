package oauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

const (
	oauthTestClientID     = "oauth-test-client-id-real-value"
	oauthTestClientSecret = "oauth-test-client-secret-real-value"
	oauthTestRefreshToken = "oauth-test-refresh-token-real-value"
)

type stubClientResponse struct {
	status int
	body   string
	err    error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
type stubRoundTripper struct {
	response stubClientResponse
}

func (rt *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		_, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	if rt.response.err != nil {
		return nil, rt.response.err
	}
	return &http.Response{
		StatusCode: rt.response.status,
		Body:       io.NopCloser(strings.NewReader(rt.response.body)),
		Header:     make(http.Header),
	}, nil
}

func newTokenSourceWithStub(response stubClientResponse) *TokenSource {
	return NewTokenSource(
		&http.Client{Transport: &stubRoundTripper{response: response}},
		oauthTestClientID,
		oauthTestClientSecret,
		oauthTestRefreshToken,
	)
}

func TestToken_returnsInfrastructureError_whenUnauthorized(t *testing.T) {
	// Given: Client Stub が 401 を返す
	source := newTokenSourceWithStub(stubClientResponse{
		status: http.StatusUnauthorized,
		body:   `{"error":"invalid_grant"}`,
	})

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返し、Error / Unwrap が観測できる
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
	if !strings.HasPrefix(oauthErr.Error(), "google oauth:") {
		t.Fatalf("Error() = %q, want prefix google oauth:", oauthErr.Error())
	}
	if errors.Unwrap(oauthErr) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestToken_returnsInfrastructureError_whenAccessTokenIsEmpty(t *testing.T) {
	// Given: Client Stub が空 token を返す
	source := newTokenSourceWithStub(stubClientResponse{
		status: http.StatusOK,
		body:   `{"access_token":" "}`,
	})

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}

func TestToken_returnsInfrastructureError_whenClientIsNil(t *testing.T) {
	// Given: client が nil
	source := NewTokenSource(nil, oauthTestClientID, oauthTestClientSecret, oauthTestRefreshToken)

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}

func TestToken_returnsInfrastructureError_whenUpstreamConnectionFails(t *testing.T) {
	// Given: Client Stub が接続 error を返す
	source := newTokenSourceWithStub(stubClientResponse{
		err: errors.New("connection refused"),
	})

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}

func TestToken_returnsInfrastructureError_whenResponseIsInvalidJSON(t *testing.T) {
	// Given: Client Stub が JSON でない応答を返す
	source := newTokenSourceWithStub(stubClientResponse{
		status: http.StatusOK,
		body:   "not-json",
	})

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}
