package oauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

type proxyProbe struct {
	targetURL    string
	method       string
	clientID     string
	clientSecret string
	refreshToken string
	body         string
}

func newTokenSourceWithStub(t *testing.T, status int, response string) (*TokenSource, *proxyProbe) {
	t.Helper()
	probe := &proxyProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("proxy body read: %v", err)
		}
		probe.targetURL = r.Header.Get("X-AS-Target-URL")
		probe.method = r.Header.Get("X-AS-Method")
		probe.clientID = r.Header.Get("X-AS-Inject-Form-client_id")
		probe.clientSecret = r.Header.Get("X-AS-Inject-Form-client_secret")
		probe.refreshToken = r.Header.Get("X-AS-Inject-Form-refresh_token")
		probe.body = string(body)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return NewTokenSource(&agentsecrets.Client{
		HTTP:     server.Client(),
		ProxyURL: server.URL,
	}), probe
}

func TestToken_returnsAccessToken_andInjectsSecretKeyNames(t *testing.T) {
	t.Parallel()

	// Given: token endpoint stub が access token を返す
	source, probe := newTokenSourceWithStub(t, http.StatusOK, `{"access_token":"ya29.test-token"}`)

	// When: OAuth refresh を実行する
	token, err := source.Token(context.Background())

	// Then: token と proxy request shape が contract を満たす
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if token != "ya29.test-token" {
		t.Fatalf("token = %q, want ya29.test-token", token)
	}
	if probe.targetURL != tokenURL {
		t.Fatalf("target URL = %q, want %q", probe.targetURL, tokenURL)
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want POST", probe.method)
	}
	if probe.clientID != secretnames.GoogleOAuthClientIDName {
		t.Fatalf("client_id key name = %q", probe.clientID)
	}
	if probe.clientSecret != secretnames.GoogleOAuthClientSecretName {
		t.Fatalf("client_secret key name = %q", probe.clientSecret)
	}
	if probe.refreshToken != secretnames.GoogleOAuthRefreshTokenName {
		t.Fatalf("refresh_token key name = %q", probe.refreshToken)
	}
	form, err := url.ParseQuery(probe.body)
	if err != nil {
		t.Fatalf("request body parse: %v", err)
	}
	if form.Get("grant_type") != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", form.Get("grant_type"))
	}
}

func TestToken_returnsInfrastructureError_whenUnauthorized(t *testing.T) {
	t.Parallel()

	// Given: token endpoint stub が 401 を返す
	source, _ := newTokenSourceWithStub(t, http.StatusUnauthorized, `{"error":"invalid_grant"}`)

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}

func TestToken_returnsInfrastructureError_whenAccessTokenIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: token endpoint stub が空 token を返す
	source, _ := newTokenSourceWithStub(t, http.StatusOK, `{"access_token":" "}`)

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}
