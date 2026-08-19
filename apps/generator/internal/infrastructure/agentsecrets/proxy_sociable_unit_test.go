package agentsecrets_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

type proxyProbe struct {
	TargetURL        string
	Method           string
	Bearer           string
	APIKey           string
	FormClientID     string
	FormClientSecret string
	FormRefreshToken string
	BodyParents0     string
	ContentType      string
	Authorization    string
}

func newClientWithProxyProbe(t *testing.T, responseBody string) (*agentsecrets.Client, *proxyProbe) {
	t.Helper()
	probe := &proxyProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probe.TargetURL = r.Header.Get("X-AS-Target-URL")
		probe.Method = r.Header.Get("X-AS-Method")
		probe.Bearer = r.Header.Get("X-AS-Inject-Bearer")
		probe.APIKey = r.Header.Get("X-AS-Inject-Header-X-API-Key")
		probe.FormClientID = r.Header.Get("X-AS-Inject-Form-client_id")
		probe.FormClientSecret = r.Header.Get("X-AS-Inject-Form-client_secret")
		probe.FormRefreshToken = r.Header.Get("X-AS-Inject-Form-refresh_token")
		probe.BodyParents0 = r.Header.Get("X-AS-Inject-Body-parents-0")
		probe.ContentType = r.Header.Get("Content-Type")
		probe.Authorization = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, responseBody)
	}))
	t.Cleanup(server.Close)
	return &agentsecrets.Client{HTTP: server.Client(), ProxyURL: server.URL}, probe
}

func TestDo_sendsKeyNameHeadersAndReturnsBody_whenBearerInjected(t *testing.T) {
	t.Parallel()

	// Given: 固定応答を返す proxy double と、Bearer にキー名を入れた Request
	const responseBody = `{"ok":true}`
	client, probe := newClientWithProxyProbe(t, responseBody)
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    agentsecrets.Inject{Bearer: secretnames.GetXAPIKeyName},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: proxy へキー名ヘッダが渡り、応答 body が返る
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(body) != responseBody {
		t.Fatalf("body = %q, want %q", body, responseBody)
	}
	if probe.TargetURL != "https://api.example.com/v1/posts" {
		t.Fatalf("X-AS-Target-URL = %q", probe.TargetURL)
	}
	if probe.Method != http.MethodGet {
		t.Fatalf("X-AS-Method = %q", probe.Method)
	}
	if probe.Bearer != secretnames.GetXAPIKeyName {
		t.Fatalf("X-AS-Inject-Bearer = %q, want %q", probe.Bearer, secretnames.GetXAPIKeyName)
	}
}

func TestDo_usesGETMethodHeader_whenMethodEmpty(t *testing.T) {
	t.Parallel()

	// Given: Method が空の Request と、ヘッダを記録する proxy double
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    agentsecrets.Inject{Bearer: secretnames.GetXAPIKeyName},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: X-AS-Method が GET になる
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.Method != http.MethodGet {
		t.Fatalf("X-AS-Method = %q, want GET", probe.Method)
	}
}

func TestDo_omitsBearerHeader_whenBearerEmpty(t *testing.T) {
	t.Parallel()

	// Given: Inject.Bearer が空の Request と、ヘッダを記録する proxy double
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "https://api.example.com/v1/posts",
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: X-AS-Inject-Bearer が付かない
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.Bearer != "" {
		t.Fatalf("X-AS-Inject-Bearer = %q, want empty", probe.Bearer)
	}
}

func TestDo_sendsCustomHeaderKeyName_whenHeaderInjected(t *testing.T) {
	t.Parallel()

	// Given: Headers に upstream header 名とキー名を入れた Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "https://api.example.com/v1/posts",
		Inject: agentsecrets.Inject{
			Headers: map[string]string{"X-API-Key": "EXAMPLE_API_KEY"},
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: X-AS-Inject-Header-X-API-Key にキー名が載り、Bearer は付かない
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.APIKey != "EXAMPLE_API_KEY" {
		t.Fatalf("X-AS-Inject-Header-X-API-Key = %q, want EXAMPLE_API_KEY", probe.APIKey)
	}
	if probe.Bearer != "" {
		t.Fatalf("X-AS-Inject-Bearer = %q, want empty", probe.Bearer)
	}
}

func TestDo_returnsError_whenTargetURLEmpty(t *testing.T) {
	t.Parallel()

	// Given: TargetURL が空白のみの Request（proxy には到達しない）
	client := &agentsecrets.Client{HTTP: http.DefaultClient, ProxyURL: "http://127.0.0.1:9/proxy"}
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "   ",
	}

	// When: Client.Do を呼び出す
	_, err := client.Do(context.Background(), req)

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error for blank TargetURL")
	}
}

func TestDo_returnsError_whenTargetURLSchemeIsNotHTTPS(t *testing.T) {
	t.Parallel()

	// Given: TargetURL の scheme が https 以外の Request
	client := &agentsecrets.Client{HTTP: http.DefaultClient, ProxyURL: "http://127.0.0.1:9/proxy"}
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "http://api.example.com/v1/posts",
	}

	// When: Client.Do を呼び出す
	_, err := client.Do(context.Background(), req)

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error for non-https TargetURL")
	}
}

func TestDo_returnsError_whenTargetURLHostEmpty(t *testing.T) {
	t.Parallel()

	// Given: scheme は https だが host が空の Request
	client := &agentsecrets.Client{HTTP: http.DefaultClient, ProxyURL: "http://127.0.0.1:9/proxy"}
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "https://",
	}

	// When: Client.Do を呼び出す
	_, err := client.Do(context.Background(), req)

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error for https URL without host")
	}
}

func TestDo_returnsError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: receiver が nil
	var client *agentsecrets.Client

	// When: Client.Do を呼び出す
	_, err := client.Do(context.Background(), agentsecrets.Request{
		TargetURL: "https://api.example.com/v1/posts",
	})

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestDo_returnsError_whenContextNil(t *testing.T) {
	t.Parallel()

	// Given: ctx が nil
	client := &agentsecrets.Client{HTTP: http.DefaultClient, ProxyURL: "http://127.0.0.1:9/proxy"}

	// When: Client.Do を呼び出す
	_, err := client.Do(nil, agentsecrets.Request{ //nolint:staticcheck // 契約どおり nil ctx を検証する
		TargetURL: "https://api.example.com/v1/posts",
	})

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestDo_usesDefaultProxyAndHTTPClient_whenFieldsEmpty(t *testing.T) {
	t.Parallel()

	// Given: ProxyURL と HTTP が空の Client（Default を使う経路）と、到達不能な既定 proxy
	client := &agentsecrets.Client{}
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "https://api.example.com/v1/posts",
	}

	// When: Client.Do を呼び出す
	_, err := client.Do(context.Background(), req)

	// Then: 既定 proxy へ向かい、接続失敗の error が返る（Default 経路を踏んだ証拠）
	if err == nil {
		t.Fatal("expected error when default proxy is unreachable")
	}
}

func TestDo_skipsBlankHeaderInjectPairs(t *testing.T) {
	t.Parallel()

	// Given: Headers に空の header 名またはキー名を含む Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: "https://api.example.com/v1/posts",
		Inject: agentsecrets.Inject{
			Headers: map[string]string{
				"X-API-Key": "EXAMPLE_API_KEY",
				"":          "IGNORED",
				"X-Empty":   "  ",
			},
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: 有効な組だけが載る
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.APIKey != "EXAMPLE_API_KEY" {
		t.Fatalf("X-AS-Inject-Header-X-API-Key = %q", probe.APIKey)
	}
}

func TestDo_sendsFormInjectKeyNames_whenFormInjected(t *testing.T) {
	t.Parallel()

	// Given: Form に OAuth 用フィールド名とキー名を入れた Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: "https://oauth2.googleapis.com/token",
		Inject: agentsecrets.Inject{
			Form: map[string]string{
				"client_id":     secretnames.GoogleOAuthClientIDName,
				"client_secret": secretnames.GoogleOAuthClientSecretName,
				"refresh_token": secretnames.GoogleOAuthRefreshTokenName,
			},
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: X-AS-Inject-Form-* にキー名だけが載る
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.FormClientID != secretnames.GoogleOAuthClientIDName {
		t.Fatalf("X-AS-Inject-Form-client_id = %q, want %q", probe.FormClientID, secretnames.GoogleOAuthClientIDName)
	}
	if probe.FormClientSecret != secretnames.GoogleOAuthClientSecretName {
		t.Fatalf("X-AS-Inject-Form-client_secret = %q, want %q", probe.FormClientSecret, secretnames.GoogleOAuthClientSecretName)
	}
	if probe.FormRefreshToken != secretnames.GoogleOAuthRefreshTokenName {
		t.Fatalf("X-AS-Inject-Form-refresh_token = %q, want %q", probe.FormRefreshToken, secretnames.GoogleOAuthRefreshTokenName)
	}
}

func TestDo_sendsBodyInjectKeyNameWithDashes_whenNestedJSONPath(t *testing.T) {
	t.Parallel()

	// Given: Body に JSON path parents.0 とキー名を入れた Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: "https://www.googleapis.com/drive/v3/files",
		Inject: agentsecrets.Inject{
			Body: map[string]string{"parents.0": secretnames.DriveFolderIDName},
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: X-AS-Inject-Body-parents-0 にキー名が載る（dot は dash へ）
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.BodyParents0 != secretnames.DriveFolderIDName {
		t.Fatalf("X-AS-Inject-Body-parents-0 = %q, want %q", probe.BodyParents0, secretnames.DriveFolderIDName)
	}
}

func TestDo_passthroughsNonSecretHeaders_whenHeadersSet(t *testing.T) {
	t.Parallel()

	// Given: 非秘密の Content-Type と Authorization を載せた Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: "https://www.googleapis.com/drive/v3/files",
		PassthroughHeaders: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer dummy-access-token",
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: 値そのものが proxy へ渡り、Inject header は付かない
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.ContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", probe.ContentType)
	}
	if probe.Authorization != "Bearer dummy-access-token" {
		t.Fatalf("Authorization = %q", probe.Authorization)
	}
	if probe.Bearer != "" {
		t.Fatalf("X-AS-Inject-Bearer = %q, want empty", probe.Bearer)
	}
}

func TestDo_skipsBlankFormAndBodyInjectPairs(t *testing.T) {
	t.Parallel()

	// Given: Form / Body に空の名前またはキー名を含む Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: "https://oauth2.googleapis.com/token",
		Inject: agentsecrets.Inject{
			Form: map[string]string{
				"client_id": secretnames.GoogleOAuthClientIDName,
				"":          "IGNORED",
				"x-empty":   "  ",
			},
			Body: map[string]string{
				"parents.0": secretnames.DriveFolderIDName,
				"":          "IGNORED",
				"x-empty":   "  ",
			},
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: 有効な組だけが載る
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.FormClientID != secretnames.GoogleOAuthClientIDName {
		t.Fatalf("X-AS-Inject-Form-client_id = %q", probe.FormClientID)
	}
	if probe.BodyParents0 != secretnames.DriveFolderIDName {
		t.Fatalf("X-AS-Inject-Body-parents-0 = %q", probe.BodyParents0)
	}
}

func TestDo_skipsBlankPassthroughHeaders(t *testing.T) {
	t.Parallel()

	// Given: Headers に空の名前または値を含む Request
	client, probe := newClientWithProxyProbe(t, `{}`)
	req := agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: "https://www.googleapis.com/drive/v3/files",
		PassthroughHeaders: map[string]string{
			"Content-Type": "application/json",
			"":             "IGNORED",
			"X-Empty":      "  ",
		},
	}

	// When: Client.Do を呼び出す
	res, err := client.Do(context.Background(), req)

	// Then: 有効な組だけが載る
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if probe.ContentType != "application/json" {
		t.Fatalf("Content-Type = %q", probe.ContentType)
	}
}
