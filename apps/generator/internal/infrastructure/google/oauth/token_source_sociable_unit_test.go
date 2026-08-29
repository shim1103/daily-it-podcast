package oauth

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	oauthTestClientID     = "oauth-test-client-id-real-value"
	oauthTestClientSecret = "oauth-test-client-secret-real-value"
	oauthTestRefreshToken = "oauth-test-refresh-token-real-value"
)

type proxyProbe struct {
	targetURL    string
	method       string
	clientID     string
	clientSecret string
	refreshToken string
	body         string
	contentType  string
}

// newTokenSourceWithStub は本番 host（oauth2.googleapis.com）への接続を test TLS server へ差し替えた TokenSource を返す。
// why: Adapter は tokenURL を定数として持つため、DialTLSContext で接続先だけを test server へ redirect する。
func newTokenSourceWithStub(t *testing.T, status int, response string) (*TokenSource, *proxyProbe) {
	t.Helper()
	probe := &proxyProbe{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		probe.targetURL = r.URL.String()
		probe.method = r.Method
		probe.contentType = r.Header.Get("Content-Type")
		probe.clientID = r.PostForm.Get("client_id")
		probe.clientSecret = r.PostForm.Get("client_secret")
		probe.refreshToken = r.PostForm.Get("refresh_token")
		probe.body = r.PostForm.Encode()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	return NewTokenSource(httpClient, oauthTestClientID, oauthTestClientSecret, oauthTestRefreshToken), probe
}

func TestToken_returnsAccessToken_andSendsRefreshForm(t *testing.T) {
	// Given: token endpoint stub が access token を返す
	source, probe := newTokenSourceWithStub(t, http.StatusOK, `{"access_token":"ya29.test-token"}`)

	// When: OAuth refresh を実行する
	token, err := source.Token(context.Background())

	// Then: token と upstream request shape が contract を満たす
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if token != "ya29.test-token" {
		t.Fatalf("token = %q, want ya29.test-token", token)
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want POST", probe.method)
	}
	if probe.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", probe.contentType)
	}
	if probe.clientID != oauthTestClientID {
		t.Fatalf("client_id = %q, want real value", probe.clientID)
	}
	if probe.clientSecret != oauthTestClientSecret {
		t.Fatalf("client_secret = %q, want real value", probe.clientSecret)
	}
	if probe.refreshToken != oauthTestRefreshToken {
		t.Fatalf("refresh_token = %q, want real value", probe.refreshToken)
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
	// Given: token endpoint stub が 401 を返す
	source, _ := newTokenSourceWithStub(t, http.StatusUnauthorized, `{"error":"invalid_grant"}`)

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
	// Given: 直ちに閉じる upstream への接続（DialTLSContext が閉じた addr を指す）
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := server.Listener.Addr().String()
	server.Close()
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addrIgnored string) (net.Conn, error) {
				return tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	source := NewTokenSource(httpClient, oauthTestClientID, oauthTestClientSecret, oauthTestRefreshToken)

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}

func TestToken_returnsInfrastructureError_whenResponseIsInvalidJSON(t *testing.T) {
	// Given: token endpoint stub が JSON でない応答を返す
	source, _ := newTokenSourceWithStub(t, http.StatusOK, "not-json")

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}
