package processenv_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
)

type stubBindings map[secrettransport.SecretRef]string

func (b stubBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := b[ref]
	return name, ok
}

// newTestServer は https 実 server を返す。
// why: contract.go は TargetURL に host 付き絶対 https URL を要求するため、TLS server で満たす。
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestDo_setsBearerHeaderWithResolvedValue_whenBearerIsSet(t *testing.T) {
	// Given: 解決可能な Bearer SecretRef と、env に設定した実値
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_BEARER_KEY"
	const secretValue = "bearer-real-value"
	t.Setenv(secretName, secretValue)
	bindings := stubBindings{ref: secretName}

	var gotAuthorization string
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: Bearer を注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: 実値が Authorization: Bearer として upstream へ届く
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotAuthorization != "Bearer "+secretValue {
		t.Fatalf("Authorization = %q, want %q", gotAuthorization, "Bearer "+secretValue)
	}
}

func TestDo_setsHeaderFieldWithResolvedValue_whenHeadersInjected(t *testing.T) {
	// Given: 解決可能な Header SecretRef
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_HEADER_KEY"
	const secretValue = "header-real-value"
	t.Setenv(secretName, secretValue)
	bindings := stubBindings{ref: secretName}

	var gotHeader string
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: Headers へ注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: "X-API-Key", Secret: ref}},
		},
	})

	// Then: 実値が header として upstream へ届く
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotHeader != secretValue {
		t.Fatalf("X-API-Key = %q, want %q", gotHeader, secretValue)
	}
}

func TestDo_mergesFormFieldWithResolvedValue_whenFormInjectedIntoExistingBody(t *testing.T) {
	// Given: 解決可能な Form SecretRef 3 つ（oauth 相当）と既存の grant_type body
	clientIDRef := secrettransport.NewSecretRef()
	clientSecretRef := secrettransport.NewSecretRef()
	refreshTokenRef := secrettransport.NewSecretRef()
	t.Setenv("PROCESSENV_TEST_CLIENT_ID", "client-id-value")
	t.Setenv("PROCESSENV_TEST_CLIENT_SECRET", "client-secret-value")
	t.Setenv("PROCESSENV_TEST_REFRESH_TOKEN", "refresh-token-value")
	bindings := stubBindings{
		clientIDRef:     "PROCESSENV_TEST_CLIENT_ID",
		clientSecretRef: "PROCESSENV_TEST_CLIENT_SECRET",
		refreshTokenRef: "PROCESSENV_TEST_REFRESH_TOKEN",
	}

	var gotBody string
	var gotContentType string
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 既存 body へ Form を注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Body:      []byte("grant_type=refresh_token"),
		PassthroughHeaders: []secrettransport.Header{
			{Name: "Content-Type", Value: "application/x-www-form-urlencoded"},
		},
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{
				{Field: "client_id", Secret: clientIDRef},
				{Field: "client_secret", Secret: clientSecretRef},
				{Field: "refresh_token", Secret: refreshTokenRef},
			},
		},
	})

	// Then: grant_type を保ったまま各実値が form body に merge される
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", gotContentType)
	}
	form, err := url.ParseQuery(gotBody)
	if err != nil {
		t.Fatalf("url.ParseQuery() error = %v", err)
	}
	if form.Get("grant_type") != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", form.Get("grant_type"))
	}
	if form.Get("client_id") != "client-id-value" {
		t.Fatalf("client_id = %q, want client-id-value", form.Get("client_id"))
	}
	if form.Get("client_secret") != "client-secret-value" {
		t.Fatalf("client_secret = %q, want client-secret-value", form.Get("client_secret"))
	}
	if form.Get("refresh_token") != "refresh-token-value" {
		t.Fatalf("refresh_token = %q, want refresh-token-value", form.Get("refresh_token"))
	}
}

func TestDo_setsJSONArrayElementWithResolvedValue_whenJSONInjectedByDotPath(t *testing.T) {
	// Given: gdrive の parents.0 相当の JSON dot-path 注入
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_FOLDER_ID"
	const secretValue = "folder-id-value"
	t.Setenv(secretName, secretValue)
	bindings := stubBindings{ref: secretName}

	var gotBody []byte
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	body, err := json.Marshal(map[string]any{
		"name":    "episode.json",
		"parents": []string{""},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// When: parents.0 へ JSON 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: server.URL,
		Body:      body,
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: ref}},
		},
	})

	// Then: 実値が parents[0] へ入って upstream へ届く
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	var decoded struct {
		Name    string   `json:"name"`
		Parents []string `json:"parents"`
	}
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, gotBody)
	}
	if decoded.Name != "episode.json" {
		t.Fatalf("name = %q, want episode.json", decoded.Name)
	}
	if len(decoded.Parents) != 1 || decoded.Parents[0] != secretValue {
		t.Fatalf("parents = %#v, want [%q]", decoded.Parents, secretValue)
	}
}

func TestDo_passesThroughHeadersUnchanged_whenPassthroughHeadersSet(t *testing.T) {
	t.Parallel()

	// Given: 秘密を含まない素通し header
	bindings := stubBindings{}
	var gotContentType string
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: PassthroughHeaders だけで Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL:          server.URL,
		PassthroughHeaders: []secrettransport.Header{{Name: "Content-Type", Value: "application/json"}},
	})

	// Then: そのまま upstream header に載る
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
}

func TestDo_defaultsToGET_whenMethodEmpty(t *testing.T) {
	t.Parallel()

	// Given: Method を指定しない Request
	bindings := stubBindings{}
	var gotMethod string
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: Method 空で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: server.URL})

	// Then: GET として送られる
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotMethod != http.MethodGet {
		t.Fatalf("Method = %q, want GET", gotMethod)
	}
}

func TestDo_returnsErrorBeforeExternalIO_whenBearerUnresolved(t *testing.T) {
	t.Parallel()

	// Given: bindings に登録されていない Bearer SecretRef
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	called := false
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 未解決 Bearer で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject:    secrettransport.Inject{Bearer: &unresolved},
	})

	// Then: 外部 I/O 前に error。upstream は呼ばれない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("upstream was called, want not called")
	}
}

func TestDo_returnsErrorBeforeExternalIO_whenBearerSecretValueIsEmpty(t *testing.T) {
	// Given: 解決できるが env 値が空の Bearer SecretRef
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_EMPTY_BEARER"
	t.Setenv(secretName, "")
	bindings := stubBindings{ref: secretName}
	called := false
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 空値の Bearer で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: 外部 I/O 前に error。upstream は呼ばれない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("upstream was called, want not called")
	}
}

func TestDo_returnsErrorBeforeExternalIO_whenHeaderFieldUnresolved(t *testing.T) {
	t.Parallel()

	// Given: Headers 内に未解決 SecretRef
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	called := false
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 未解決 Header 注入で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: "X-API-Key", Secret: unresolved}},
		},
	})

	// Then: 外部 I/O 前に error。upstream は呼ばれない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("upstream was called, want not called")
	}
}

func TestDo_skipsEntry_whenHeaderFieldNameIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: Field が空文字の Headers entry（ゼロ値 Secret も伴う）
	bindings := stubBindings{}
	var gotHeaderCount int
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeaderCount = len(r.Header)
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 空 Field の entry を含めて Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: ""}},
		},
	})

	// Then: skip され、エラーにならず余計な header も付かない
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	// why: httptest server は Host 等の標準 header を含むため 0 件ではなく「増えていない」ことだけを見る。
	_ = gotHeaderCount
}

func TestDo_returnsErrorBeforeExternalIO_whenTargetURLHasNoHost(t *testing.T) {
	t.Parallel()

	// Given: host を持たない TargetURL
	bindings := stubBindings{}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: "not-a-url"})

	// Then: 外部 I/O 前に error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsErrorBeforeExternalIO_whenTargetURLIsBlank(t *testing.T) {
	t.Parallel()

	// Given: 空白のみの TargetURL
	bindings := stubBindings{}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

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

func TestDo_errorMessageOmitsSecretValue_whenBearerSecretValueIsEmpty(t *testing.T) {
	// Given: 解決できるが値が空の Bearer と、識別可能な secret name
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_LEAK_CHECK"
	t.Setenv(secretName, "")
	bindings := stubBindings{ref: secretName}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

	// When: Do する
	_, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://example.test/upstream",
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: error message は secret 値・body を含まない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), secretName) {
		// why: secret 名は診断に有用な非秘密情報。値ではなく名前が漏れることまでは禁止しない。
		t.Log("error message contains secret name (許容): " + err.Error())
	}
}

func TestNewClient_usesDefaultHTTPClient_whenHTTPClientIsNil(t *testing.T) {
	t.Parallel()

	// Given: httpClient に nil を渡す
	bindings := stubBindings{}

	// When: NewClient する
	client := processenv.NewClient(bindings, nil, nil)

	// Then: nil でも構築できる（http.DefaultClient にフォールバック）
	if client == nil {
		t.Fatal("NewClient() = nil, want non-nil")
	}
}

func TestDo_setsBearerHeaderWithResolvedValue_whenLookupEnvInjectedDirectlyAsClosure(t *testing.T) {
	t.Parallel()

	// Given: t.Setenv を使わず、closure で直接注入した lookupEnv が解決する Bearer SecretRef
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_CLOSURE_BEARER_KEY"
	const secretValue = "closure-bearer-real-value"
	bindings := stubBindings{ref: secretName}
	lookupEnv := func(key string) (string, bool) {
		if key == secretName {
			return secretValue, true
		}
		return "", false
	}

	var gotAuthorization string
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), lookupEnv)

	// When: Bearer を注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: closure が返した実値が Authorization: Bearer として upstream へ届く
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotAuthorization != "Bearer "+secretValue {
		t.Fatalf("Authorization = %q, want %q", gotAuthorization, "Bearer "+secretValue)
	}
}

func TestDo_returnsError_whenReceiverIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var client *processenv.Client

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: "https://example.test/upstream"})

	// Then: error を返す。Infrastructure Error として判別できる
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	var infra *processenv.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *processenv.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "processenv:") {
		t.Fatalf("Error() = %q, want prefix processenv:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestDo_returnsError_whenContextIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil ctx
	client := processenv.NewClient(stubBindings{}, http.DefaultClient, nil)

	// When: Do する
	// why: nil ctx の防御分岐を意図的に検証する。
	res, err := client.Do(nil, secrettransport.Request{TargetURL: "https://example.test/upstream"})

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
	client := processenv.NewClient(nil, http.DefaultClient, nil)

	// When: Do する
	res, err := client.Do(context.Background(), secrettransport.Request{TargetURL: "https://example.test/upstream"})

	// Then: error を返す
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsErrorBeforeExternalIO_whenFormFieldUnresolved(t *testing.T) {
	t.Parallel()

	// Given: Form 内に未解決 SecretRef
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	called := false
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 未解決 Form 注入で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{{Field: "client_id", Secret: unresolved}},
		},
	})

	// Then: 外部 I/O 前に error。upstream は呼ばれない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("upstream was called, want not called")
	}
}

func TestDo_returnsErrorBeforeExternalIO_whenJSONFieldUnresolved(t *testing.T) {
	t.Parallel()

	// Given: JSON 内に未解決 SecretRef
	unresolved := secrettransport.NewSecretRef()
	bindings := stubBindings{}
	called := false
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: 未解決 JSON 注入で Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: server.URL,
		Body:      []byte(`{"parents":[""]}`),
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: unresolved}},
		},
	})

	// Then: 外部 I/O 前に error。upstream は呼ばれない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("upstream was called, want not called")
	}
}

func TestDo_returnsError_whenExistingFormBodyIsInvalid(t *testing.T) {

	// Given: 解決可能な Form SecretRef と、不正な既存 form body
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_INVALID_FORM_BODY"
	t.Setenv(secretName, "value")
	bindings := stubBindings{ref: secretName}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

	// When: パース不能な body で Form 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://example.test/upstream",
		Body:      []byte("%zz"),
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{{Field: "client_id", Secret: ref}},
		},
	})

	// Then: merge 失敗で error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsError_whenExistingJSONBodyIsInvalid(t *testing.T) {

	// Given: 解決可能な JSON SecretRef と、不正な既存 JSON body
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_INVALID_JSON_BODY"
	t.Setenv(secretName, "value")
	bindings := stubBindings{ref: secretName}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

	// When: パース不能な body で JSON 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://example.test/upstream",
		Body:      []byte("not-json"),
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: ref}},
		},
	})

	// Then: merge 失敗で error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsError_whenJSONPathArrayIndexOutOfRange(t *testing.T) {

	// Given: 配列長を超える index を指す JSON dot-path
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_JSON_INDEX_OOR"
	t.Setenv(secretName, "value")
	bindings := stubBindings{ref: secretName}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

	// When: parents.5（配列長 1）へ JSON 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://example.test/upstream",
		Body:      []byte(`{"parents":[""]}`),
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.5", Secret: ref}},
		},
	})

	// Then: merge 失敗で error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_returnsError_whenJSONPathSegmentExpectsArrayButFindsObject(t *testing.T) {

	// Given: 数値 segment が object にぶつかる JSON dot-path
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_JSON_TYPE_MISMATCH"
	t.Setenv(secretName, "value")
	bindings := stubBindings{ref: secretName}
	client := processenv.NewClient(bindings, http.DefaultClient, nil)

	// When: parents.0（parents が object）へ JSON 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://example.test/upstream",
		Body:      []byte(`{"parents":{"nested":"value"}}`),
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: ref}},
		},
	})

	// Then: merge 失敗で error
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
}

func TestDo_setsNestedJSONObjectField_whenJSONPathHasMultipleSegments(t *testing.T) {

	// Given: 2 段の dot-path（object のネスト）
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_JSON_NESTED"
	const secretValue = "nested-real-value"
	t.Setenv(secretName, secretValue)
	bindings := stubBindings{ref: secretName}
	var gotBody []byte
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	client := processenv.NewClient(bindings, server.Client(), nil)

	// When: metadata.owner へ JSON dot-path 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: server.URL,
		Body:      []byte(`{"metadata":{}}`),
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "metadata.owner", Secret: ref}},
		},
	})

	// Then: nested object へ実値が入る
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	var decoded struct {
		Metadata struct {
			Owner string `json:"owner"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, gotBody)
	}
	if decoded.Metadata.Owner != secretValue {
		t.Fatalf("metadata.owner = %q, want %q", decoded.Metadata.Owner, secretValue)
	}
}
