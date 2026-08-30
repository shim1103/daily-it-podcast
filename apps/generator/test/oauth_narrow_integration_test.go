// Scope: Narrow Integration
// 実物境界: oauth.TokenSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 Google OAuth API は使わない。DialTLSContext で本番 host 宛先だけを test server へ redirect する。
// @require dummy credential を Adapter へ直接渡す。upstream は controllable な test server。
// @ensure upstream は POST を受け取り、refresh grant form に必要な field が届く。
// @ensure 成功時 Token は非空 access token を返す。
// @invariant dummy credential 実値は error message へ出ない。
package test

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
)

const (
	narrowOAuthClientID     = "narrow-oauth-client-id-real-value"
	narrowOAuthClientSecret = "narrow-oauth-client-secret-real-value"
	narrowOAuthRefreshToken = "narrow-oauth-refresh-token-real-value"
)

type oauthNarrowProbe struct {
	method       string
	contentType  string
	clientID     string
	clientSecret string
	refreshToken string
	grantType    string
	body         string
}

// newTokenSourceWithProxy は本番 host（oauth2.googleapis.com）への接続を test TLS server へ差し替えた TokenSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure dummy credential は Adapter へ直接渡し、標準 *http.Client が refresh form へ載せる。
func newTokenSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*oauth.TokenSource, *oauthNarrowProbe) {
	t.Helper()
	probe := &oauthNarrowProbe{}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		probe.method = r.Method
		probe.contentType = r.Header.Get("Content-Type")
		probe.clientID = r.PostForm.Get("client_id")
		probe.clientSecret = r.PostForm.Get("client_secret")
		probe.refreshToken = r.PostForm.Get("refresh_token")
		probe.grantType = r.PostForm.Get("grant_type")
		probe.body = r.PostForm.Encode()
		handler(w, r)
	}))
	t.Cleanup(upstream.Close)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	return oauth.NewTokenSource(httpClient, narrowOAuthClientID, narrowOAuthClientSecret, narrowOAuthRefreshToken), probe
}

func TestOAuthTokenSource_deliversPostWithRefreshGrantForm_whenUpstreamSucceeds(t *testing.T) {
	// Given: dummy credential と、成功応答を返す upstream double
	const wantToken = "ya29.narrow-access-token"
	source, probe := newTokenSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"` + wantToken + `"}`))
	})

	// When: Token refresh を実行する
	got, err := source.Token(context.Background())

	// Then: upstream は POST を受け、refresh grant form が contract を満たし、access token が返る
	if err != nil {
		t.Fatalf("Token() error = %v, want nil", err)
	}
	if got != wantToken {
		t.Fatalf("token mismatch, want non-empty access token")
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want %q", probe.method, http.MethodPost)
	}
	if probe.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", probe.contentType)
	}
	if probe.clientID != narrowOAuthClientID {
		t.Fatalf("client_id mismatch")
	}
	if probe.clientSecret != narrowOAuthClientSecret {
		t.Fatalf("client_secret mismatch")
	}
	if probe.refreshToken != narrowOAuthRefreshToken {
		t.Fatalf("refresh_token mismatch")
	}
	if probe.grantType != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", probe.grantType)
	}
}

func TestOAuthTokenSource_excludesDummyCredentialsFromErrorMessage_whenUpstreamFails(t *testing.T) {
	// Given: dummy credential と、常に 401 を返す upstream double
	source, _ := newTokenSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	})

	// When: Token refresh を実行する
	_, err := source.Token(context.Background())

	// Then: error は返るが、dummy credential 実値は error message に出ない
	if err == nil {
		t.Fatal("Token() error = nil, want non-nil")
	}
	errMsg := err.Error()
	for _, secret := range []string{narrowOAuthClientID, narrowOAuthClientSecret, narrowOAuthRefreshToken} {
		if strings.Contains(errMsg, secret) {
			t.Fatalf("error message contains dummy credential value")
		}
	}
}
