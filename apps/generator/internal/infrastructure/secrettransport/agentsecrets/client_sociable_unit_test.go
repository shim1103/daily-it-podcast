package agentsecrets_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/agentsecrets"
)

type stubBindings map[secrettransport.SecretRef]string

func (b stubBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := b[ref]
	return name, ok
}

type proxyProbe struct {
	called           bool
	targetURL        string
	method           string
	bearer           string
	apiKey           string
	formClientID     string
	formClientSecret string
	formRefreshToken string
	bodyParents0     string
	contentType      string
	authorization    string
	requestBody      string
}

// urlRecordingRoundTripper は実際の network へ出ず、request URL だけを記録する。
type urlRecordingRoundTripper struct {
	url string
}

func (t *urlRecordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.url = req.URL.String()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func newClientWithProxyProbe(t *testing.T, bindings secrettransport.BindingResolver) (*agentsecrets.Client, *proxyProbe) {
	t.Helper()
	probe := &proxyProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probe.called = true
		probe.targetURL = r.Header.Get("X-AS-Target-URL")
		probe.method = r.Header.Get("X-AS-Method")
		probe.bearer = r.Header.Get("X-AS-Inject-Bearer")
		probe.apiKey = r.Header.Get("X-AS-Inject-Header-X-API-Key")
		probe.formClientID = r.Header.Get("X-AS-Inject-Form-client_id")
		probe.formClientSecret = r.Header.Get("X-AS-Inject-Form-client_secret")
		probe.formRefreshToken = r.Header.Get("X-AS-Inject-Form-refresh_token")
		probe.bodyParents0 = r.Header.Get("X-AS-Inject-Body-parents-0")
		probe.contentType = r.Header.Get("Content-Type")
		probe.authorization = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		probe.requestBody = string(raw)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)
	return agentsecrets.NewClient(bindings, server.Client(), server.URL), probe
}

func TestDo_setsBearerInjectHeaderWithSecretName_whenBearerResolved(t *testing.T) {
	t.Parallel()

	// Given: 解決可能な Bearer SecretRef（名前のみ。値は process に無い）
	ref := secrettransport.NewSecretRef()
	const secretName = "AGENTSECRETS_ST_BEARER_KEY"
	const secretValue = "bearer-must-not-appear"
	bindings := stubBindings{ref: secretName}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: Bearer を注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: X-AS-Inject-Bearer に秘密名だけが載り、値は載らない
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.bearer != secretName {
		t.Fatalf("X-AS-Inject-Bearer = %q, want %q", probe.bearer, secretName)
	}
	if strings.Contains(probe.bearer, secretValue) {
		t.Fatalf("X-AS-Inject-Bearer contains secret value: %q", probe.bearer)
	}
	if probe.targetURL != "https://api.example.com/v1/posts" {
		t.Fatalf("X-AS-Target-URL = %q", probe.targetURL)
	}
}

func TestDo_setsHeaderInjectWithSecretName_whenHeadersInjected(t *testing.T) {
	t.Parallel()

	// Given: 解決可能な Header SecretRef
	ref := secrettransport.NewSecretRef()
	const secretName = "AGENTSECRETS_ST_HEADER_KEY"
	const secretValue = "header-must-not-appear"
	bindings := stubBindings{ref: secretName}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: Headers へ注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: "X-API-Key", Secret: ref}},
		},
	})

	// Then: X-AS-Inject-Header-X-API-Key に秘密名だけが載る
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.apiKey != secretName {
		t.Fatalf("X-AS-Inject-Header-X-API-Key = %q, want %q", probe.apiKey, secretName)
	}
	if strings.Contains(probe.apiKey, secretValue) {
		t.Fatalf("inject header contains secret value: %q", probe.apiKey)
	}
}

func TestDo_setsFormInjectWithSecretName_whenFormInjected(t *testing.T) {
	t.Parallel()

	// Given: 解決可能な Form SecretRef と既存 body（実値 merge はしない）
	ref := secrettransport.NewSecretRef()
	const secretName = "AGENTSECRETS_ST_FORM_KEY"
	bindings := stubBindings{ref: secretName}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: Form 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: "https://oauth.example.com/token",
		Body:      []byte("grant_type=refresh_token"),
		PassthroughHeaders: []secrettransport.Header{
			{Name: "Content-Type", Value: "application/x-www-form-urlencoded"},
		},
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{{Field: "client_secret", Secret: ref}},
		},
	})

	// Then: X-AS-Inject-Form に秘密名が載り、body には実値を入れない
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.formClientSecret != secretName {
		t.Fatalf("X-AS-Inject-Form-client_secret = %q, want %q", probe.formClientSecret, secretName)
	}
	if probe.requestBody != "grant_type=refresh_token" {
		t.Fatalf("proxy request body = %q, want grant_type only (no secret value merge)", probe.requestBody)
	}
	if probe.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q", probe.contentType)
	}
}

func TestDo_setsBodyInjectWithSecretName_whenJSONInjectedByDotPath(t *testing.T) {
	t.Parallel()

	// Given: gdrive parents.0 相当の JSON SecretRef
	ref := secrettransport.NewSecretRef()
	const secretName = "AGENTSECRETS_ST_JSON_KEY"
	bindings := stubBindings{ref: secretName}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: JSON 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: "https://www.googleapis.com/drive/v3/files",
		Body:      []byte(`{"name":"episode.json","parents":[""]}`),
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: ref}},
		},
	})

	// Then: X-AS-Inject-Body-parents-0 に秘密名が載る（body へ実値 merge しない）
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.bodyParents0 != secretName {
		t.Fatalf("X-AS-Inject-Body-parents-0 = %q, want %q", probe.bodyParents0, secretName)
	}
	if !strings.Contains(probe.requestBody, `"parents":[""]`) {
		t.Fatalf("proxy body was mutated with secret value: %q", probe.requestBody)
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenBearerUnresolved(t *testing.T) {
	t.Parallel()

	// Given: bindings に無い Bearer SecretRef
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: 未解決 Bearer で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    secrettransport.Inject{Bearer: &unresolved},
	})

	// Then: proxy I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if probe.called {
		t.Fatal("proxy was called, want not called")
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenBearerIsInvalidNonNilZeroRef(t *testing.T) {
	t.Parallel()

	// Given: 非 nil だがゼロ値の Bearer（無効な非 nil Bearer）
	zero := secrettransport.SecretRef{}
	bindings := stubBindings{}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: 無効 Bearer で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    secrettransport.Inject{Bearer: &zero},
	})

	// Then: proxy I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if probe.called {
		t.Fatal("proxy was called, want not called")
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenBindingNameIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 解決はできるが名前が空の Bearer
	ref := secrettransport.NewSecretRef()
	bindings := stubBindings{ref: "   "}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: 空名 Bearer で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: proxy I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if probe.called {
		t.Fatal("proxy was called, want not called")
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenFieldInjectionUnresolved(t *testing.T) {
	t.Parallel()

	// Given: Inject 種別ごとの未解決 SecretRef（各 subtest で新規に作る）
	cases := []struct {
		name  string
		build func(unresolved secrettransport.SecretRef) secrettransport.Request
	}{
		{
			name: "headers",
			build: func(unresolved secrettransport.SecretRef) secrettransport.Request {
				return secrettransport.Request{
					TargetURL: "https://api.example.com/v1/posts",
					Inject: secrettransport.Inject{
						Headers: []secrettransport.FieldInjection{{Field: "X-API-Key", Secret: unresolved}},
					},
				}
			},
		},
		{
			name: "form",
			build: func(unresolved secrettransport.SecretRef) secrettransport.Request {
				return secrettransport.Request{
					TargetURL: "https://oauth.example.com/token",
					Inject: secrettransport.Inject{
						Form: []secrettransport.FieldInjection{{Field: "client_secret", Secret: unresolved}},
					},
				}
			},
		},
		{
			name: "json",
			build: func(unresolved secrettransport.SecretRef) secrettransport.Request {
				return secrettransport.Request{
					TargetURL: "https://www.googleapis.com/drive/v3/files",
					Inject: secrettransport.Inject{
						JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: unresolved}},
					},
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			unresolved := secrettransport.NewSecretRef()
			bindings := stubBindings{}
			client, probe := newClientWithProxyProbe(t, bindings)

			// When: 未解決 Field 注入で Do する
			res, err := client.Do(context.Background(), tc.build(unresolved))

			// Then: proxy I/O 前に error
			if err == nil {
				t.Fatal("Do() error = nil, want non-nil")
			}
			if res != nil {
				t.Fatalf("Do() response = %v, want nil", res)
			}
			if probe.called {
				t.Fatal("proxy was called, want not called")
			}
		})
	}
}

func TestDo_skipsEntry_whenHeaderFieldNameIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: Field が空の Headers entry
	bindings := stubBindings{}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: 空 Field を含めて Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: ""}},
		},
	})

	// Then: skip され、エラーにならず proxy へ到達する
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if !probe.called {
		t.Fatal("proxy was not called")
	}
	if probe.apiKey != "" {
		t.Fatalf("unexpected inject header = %q", probe.apiKey)
	}
}

func TestDo_skipsEntry_whenHeaderSecretIsZero(t *testing.T) {
	t.Parallel()

	// Given: Field は非空だが Secret がゼロ値の Headers entry
	bindings := stubBindings{}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: ゼロ Secret の entry を含めて Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: "X-API-Key"}},
		},
	})

	// Then: skip され、エラーにならず inject header も付かない
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if !probe.called {
		t.Fatal("proxy was not called")
	}
	if probe.apiKey != "" {
		t.Fatalf("unexpected inject header = %q", probe.apiKey)
	}
}

func TestDo_errorMessageOmitsSecretValue_whenBearerUnresolved(t *testing.T) {
	t.Parallel()

	// Given: 未解決 Bearer と、error へ混入させたくない識別文字列
	const mustNotLeak = "super-secret-value-must-not-leak"
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	client, _ := newClientWithProxyProbe(t, bindings)

	// When: Do する
	_, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Body:      []byte(mustNotLeak),
		Inject:    secrettransport.Inject{Bearer: &unresolved},
	})

	// Then: error に秘密値・request body を含めない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), mustNotLeak) {
		t.Fatalf("error message %q contains secret/body value", err.Error())
	}
}

func TestDo_returnsError_whenReceiverIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var client *agentsecrets.Client

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: "https://api.example.com/v1/posts"})

	// Then: error を返す。Infrastructure Error として判別できる
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	var infra *agentsecrets.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *agentsecrets.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "agentsecrets:") {
		t.Fatalf("Error() = %q, want prefix agentsecrets:", infra.Error())
	}
}

func TestDo_returnsError_whenContextIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil ctx
	client := agentsecrets.NewClient(stubBindings{}, http.DefaultClient, "http://127.0.0.1:9/proxy")

	// When: Do する
	// why: nil ctx の防御分岐を意図的に検証する。
	res, err := client.Do(nil, secrettransport.Request{TargetURL: "https://api.example.com/v1/posts"})

	// Then: error を返す
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsError_whenBindingsIsNil(t *testing.T) {
	t.Parallel()

	// Given: bindings が nil
	client := agentsecrets.NewClient(nil, http.DefaultClient, "http://127.0.0.1:9/proxy")

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: "https://api.example.com/v1/posts"})

	// Then: error を返す
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenTargetURLIsBlank(t *testing.T) {
	t.Parallel()

	// Given: 空白のみの TargetURL
	client := agentsecrets.NewClient(stubBindings{}, http.DefaultClient, "http://127.0.0.1:9/proxy")

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: "   "})

	// Then: 外部 I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestNewClient_returnsNonNil_whenHTTPClientAndProxyURLAreNilOrEmpty(t *testing.T) {
	t.Parallel()

	// Given: httpClient / proxyURL に nil / 空を渡す
	bindings := stubBindings{}

	// When: NewClient する
	client := agentsecrets.NewClient(bindings, nil, "")

	// Then: nil でも構築できる（DefaultClient / DefaultProxyURL にフォールバック）
	if client == nil {
		t.Fatal("NewClient() = nil, want non-nil")
	}
}

func TestDo_setsMethodHeaderToGET_whenMethodEmpty(t *testing.T) {
	t.Parallel()

	// Given: Method が空の Request
	client, probe := newClientWithProxyProbe(t, stubBindings{})

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
	})

	// Then: X-AS-Method は GET
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.method != http.MethodGet {
		t.Fatalf("X-AS-Method = %q, want GET", probe.method)
	}
}

func TestDo_omitsBearerInjectHeader_whenBearerNil(t *testing.T) {
	t.Parallel()

	// Given: Bearer を付けない Request
	client, probe := newClientWithProxyProbe(t, stubBindings{})

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
	})

	// Then: X-AS-Inject-Bearer は付かない
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.bearer != "" {
		t.Fatalf("X-AS-Inject-Bearer = %q, want empty", probe.bearer)
	}
}

func TestDo_setsMultipleFormInjectHeaders_whenMultipleFormFieldsResolved(t *testing.T) {
	t.Parallel()

	// Given: OAuth 用 Form の複数 SecretRef
	clientID := secrettransport.NewSecretRef()
	clientSecret := secrettransport.NewSecretRef()
	refreshToken := secrettransport.NewSecretRef()
	bindings := stubBindings{
		clientID:     "OAUTH_CLIENT_ID",
		clientSecret: "OAUTH_CLIENT_SECRET",
		refreshToken: "OAUTH_REFRESH_TOKEN",
	}
	client, probe := newClientWithProxyProbe(t, bindings)

	// When: Form 複数注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: "https://oauth2.googleapis.com/token",
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{
				{Field: "client_id", Secret: clientID},
				{Field: "client_secret", Secret: clientSecret},
				{Field: "refresh_token", Secret: refreshToken},
			},
		},
	})

	// Then: 各 X-AS-Inject-Form-* に秘密名だけが載る
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.formClientID != "OAUTH_CLIENT_ID" {
		t.Fatalf("X-AS-Inject-Form-client_id = %q", probe.formClientID)
	}
	if probe.formClientSecret != "OAUTH_CLIENT_SECRET" {
		t.Fatalf("X-AS-Inject-Form-client_secret = %q", probe.formClientSecret)
	}
	if probe.formRefreshToken != "OAUTH_REFRESH_TOKEN" {
		t.Fatalf("X-AS-Inject-Form-refresh_token = %q", probe.formRefreshToken)
	}
}

func TestDo_passthroughsNonSecretHeaders_whenAuthorizationAndContentTypeSet(t *testing.T) {
	t.Parallel()

	// Given: 非秘密の Content-Type と Authorization
	client, probe := newClientWithProxyProbe(t, stubBindings{})

	// When: PassthroughHeaders 付きで Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: "https://www.googleapis.com/drive/v3/files",
		PassthroughHeaders: []secrettransport.Header{
			{Name: "Content-Type", Value: "application/json"},
			{Name: "Authorization", Value: "Bearer dummy-access-token"},
			{Name: "", Value: "IGNORED"},
			{Name: "X-Empty", Value: "  "},
		},
	})

	// Then: 有効な組だけが値ごと渡り、Inject Bearer は付かない
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if probe.contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", probe.contentType)
	}
	if probe.authorization != "Bearer dummy-access-token" {
		t.Fatalf("Authorization = %q", probe.authorization)
	}
	if probe.bearer != "" {
		t.Fatalf("X-AS-Inject-Bearer = %q, want empty", probe.bearer)
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenTargetURLSchemeIsNotHTTPS(t *testing.T) {
	t.Parallel()

	// Given: scheme が https 以外の TargetURL
	client, probe := newClientWithProxyProbe(t, stubBindings{})

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "http://api.example.com/v1/posts",
	})

	// Then: 外部 I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if probe.called {
		t.Fatal("proxy was called, want not called")
	}
}

func TestDo_returnsErrorBeforeProxyIO_whenTargetURLHostEmpty(t *testing.T) {
	t.Parallel()

	// Given: host が空の https URL
	client, probe := newClientWithProxyProbe(t, stubBindings{})

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://",
	})

	// Then: 外部 I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if probe.called {
		t.Fatal("proxy was called, want not called")
	}
}

func TestDo_requestsDefaultProxyURL_whenProxyURLEmpty(t *testing.T) {
	t.Parallel()

	// Given: proxyURL 空で NewClient した Client と、到達 URL を記録する RoundTripper
	recorder := &urlRecordingRoundTripper{}
	client := agentsecrets.NewClient(stubBindings{}, &http.Client{Transport: recorder}, "")

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
	})

	// Then: 実際の request URL は DefaultProxyURL（network / port 占有に依存しない）
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if recorder.url != agentsecrets.DefaultProxyURL {
		t.Fatalf("proxy request URL = %q, want %q", recorder.url, agentsecrets.DefaultProxyURL)
	}
}
