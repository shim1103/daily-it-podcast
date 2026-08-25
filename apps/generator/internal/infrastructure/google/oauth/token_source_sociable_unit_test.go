package oauth

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
)

// stubBindings は test 用の BindingResolver fake。
type stubBindings map[secrettransport.SecretRef]string

func (b stubBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := b[ref]
	return name, ok
}

const (
	oauthTestClientIDSecretName     = "OAUTH_TEST_CLIENT_ID"
	oauthTestClientSecretSecretName = "OAUTH_TEST_CLIENT_SECRET"
	oauthTestRefreshTokenSecretName = "OAUTH_TEST_REFRESH_TOKEN"
)

type proxyProbe struct {
	targetURL    string
	method       string
	clientID     string
	clientSecret string
	refreshToken string
	body         string
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
		probe.clientID = r.PostForm.Get("client_id")
		probe.clientSecret = r.PostForm.Get("client_secret")
		probe.refreshToken = r.PostForm.Get("refresh_token")
		probe.body = r.PostForm.Encode()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	t.Setenv(oauthTestClientIDSecretName, "oauth-test-client-id-real-value")
	t.Setenv(oauthTestClientSecretSecretName, "oauth-test-client-secret-real-value")
	t.Setenv(oauthTestRefreshTokenSecretName, "oauth-test-refresh-token-real-value")
	clientIDSecret := secrettransport.NewSecretRef()
	clientSecretSecret := secrettransport.NewSecretRef()
	refreshTokenSecret := secrettransport.NewSecretRef()
	bindings := stubBindings{
		clientIDSecret:     oauthTestClientIDSecretName,
		clientSecretSecret: oauthTestClientSecretSecretName,
		refreshTokenSecret: oauthTestRefreshTokenSecretName,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	client := processenv.NewClient(bindings, httpClient, nil)
	return NewTokenSource(client, clientIDSecret, clientSecretSecret, refreshTokenSecret), probe
}

func TestToken_returnsAccessToken_andInjectsSecretRealValues(t *testing.T) {
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
	if probe.clientID != "oauth-test-client-id-real-value" {
		t.Fatalf("client_id = %q, want real value", probe.clientID)
	}
	if probe.clientSecret != "oauth-test-client-secret-real-value" {
		t.Fatalf("client_secret = %q, want real value", probe.clientSecret)
	}
	if probe.refreshToken != "oauth-test-refresh-token-real-value" {
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

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
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
	source := NewTokenSource(nil, secrettransport.NewSecretRef(), secrettransport.NewSecretRef(), secrettransport.NewSecretRef())

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

func TestToken_returnsInfrastructureError_whenSecretIsUnresolved(t *testing.T) {
	// Given: bindings に登録されていない SecretRef
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)
	source := NewTokenSource(client, unresolved, unresolved, unresolved)

	// When: OAuth refresh を実行する
	_, err := source.Token(context.Background())

	// Then: OAuth 固有の Infrastructure Error を返す
	var oauthErr *Error
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error = %T, want *oauth.Error", err)
	}
}
