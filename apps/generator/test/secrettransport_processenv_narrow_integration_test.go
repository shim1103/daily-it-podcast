// Scope: Narrow Integration
// 実物境界: processenv.Client が送信する外向き HTTP request（test upstream server）
// Double: BindingResolver は Composition と同型の in-memory map。本番 credential は使わない。
// @require dummy process environment（t.Setenv）に secret 実値をセットする。upstream は controllable な test server。
// @ensure Bearer / Header / Form / JSON の実値が upstream request へ正しく届く。
// @ensure 未解決/未設定 secret は外部 I/O 前に失敗し、upstream は呼ばれない。
// @invariant secret 実値は error message へ出ない。
package test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
)

func TestProcessEnvSecretTransportClient_injectsBearerAndHeader_whenSecretsResolved(t *testing.T) {
	// Given: dummy process environment に設定した Bearer / Header の実値と、それに解決する SecretRef
	const (
		bearerName = "NARROW_ST_BEARER_KEY"
		bearerVal  = "narrow-bearer-real-value"
		headerName = "NARROW_ST_HEADER_KEY"
		headerVal  = "narrow-header-real-value"
	)
	t.Setenv(bearerName, bearerVal)
	t.Setenv(headerName, headerVal)

	bearerRef := secrettransport.NewSecretRef()
	headerRef := secrettransport.NewSecretRef()
	bindings := narrowBindings{
		bearerRef: bearerName,
		headerRef: headerName,
	}

	var gotAuthorization string
	var gotCustomHeader string
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		gotCustomHeader = r.Header.Get("X-Narrow-Key")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	client := processenv.NewClient(bindings, upstream.Client(), nil)

	// When: Bearer + Header を注入した request を送る
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: upstream.URL,
		Inject: secrettransport.Inject{
			Bearer:  &bearerRef,
			Headers: []secrettransport.FieldInjection{{Field: "X-Narrow-Key", Secret: headerRef}},
		},
	})

	// Then: Bearer と Header の実値が upstream request へ届く
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	_ = res.Body.Close()
	if gotAuthorization != "Bearer "+bearerVal {
		t.Fatalf("Authorization = %q, want %q", gotAuthorization, "Bearer "+bearerVal)
	}
	if gotCustomHeader != headerVal {
		t.Fatalf("X-Narrow-Key = %q, want %q", gotCustomHeader, headerVal)
	}
}

func TestProcessEnvSecretTransportClient_mergesFormValue_whenSecretResolved(t *testing.T) {
	// Given: dummy process environment に設定した Form 実値と、既存の grant_type body
	const (
		formName = "NARROW_ST_FORM_KEY"
		formVal  = "narrow-form-real-value"
	)
	t.Setenv(formName, formVal)

	formRef := secrettransport.NewSecretRef()
	bindings := narrowBindings{formRef: formName}

	var formBody string
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		formBody = string(raw)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	client := processenv.NewClient(bindings, upstream.Client(), nil)

	// When: Form 注入で request を送る
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: upstream.URL,
		Body:      []byte("grant_type=refresh_token"),
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{{Field: "client_secret", Secret: formRef}},
		},
	})

	// Then: grant_type を保ったまま実値が form body に merge される
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	_ = res.Body.Close()
	form, err := url.ParseQuery(formBody)
	if err != nil {
		t.Fatalf("url.ParseQuery() error = %v", err)
	}
	if form.Get("grant_type") != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", form.Get("grant_type"))
	}
	if form.Get("client_secret") != formVal {
		t.Fatalf("client_secret = %q, want %q", form.Get("client_secret"), formVal)
	}
}

func TestProcessEnvSecretTransportClient_mergesJSONDotPath_whenSecretResolved(t *testing.T) {
	// Given: dummy process environment に設定した JSON 実値（gdrive の parents.0 相当）
	const (
		jsonName = "NARROW_ST_JSON_KEY"
		jsonVal  = "narrow-json-real-value"
	)
	t.Setenv(jsonName, jsonVal)

	jsonRef := secrettransport.NewSecretRef()
	bindings := narrowBindings{jsonRef: jsonName}

	var jsonBody []byte
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	client := processenv.NewClient(bindings, upstream.Client(), nil)
	requestBody, err := json.Marshal(map[string]any{"parents": []string{""}})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// When: JSON dot-path 注入で request を送る
	res, err := client.Do(context.Background(), secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: upstream.URL,
		Body:      requestBody,
		Inject: secrettransport.Inject{
			JSON: []secrettransport.FieldInjection{{Field: "parents.0", Secret: jsonRef}},
		},
	})

	// Then: parents[0] に実値が入る
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	_ = res.Body.Close()
	var decoded struct {
		Parents []string `json:"parents"`
	}
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, jsonBody)
	}
	if len(decoded.Parents) != 1 || decoded.Parents[0] != jsonVal {
		t.Fatalf("parents = %#v, want [%q]", decoded.Parents, jsonVal)
	}
}

func TestProcessEnvSecretTransportClient_failsBeforeUpstreamCall_whenSecretIsUnsetInProcessEnvironment(t *testing.T) {
	// Given: bindings は解決できるが、dummy process environment に secret が未設定
	const secretName = "NARROW_ST_UNSET_KEY"
	const secretValue = "narrow-must-not-leak-value"
	ref := secrettransport.NewSecretRef()
	bindings := narrowBindings{ref: secretName}

	called := false
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	client := processenv.NewClient(bindings, upstream.Client(), nil)

	// When: 未設定 secret を Bearer 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: upstream.URL,
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: 外部 I/O 前に失敗し、upstream は呼ばれず、secret 値は error に出ない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("upstream was called, want not called")
	}
	if strings.Contains(err.Error(), secretValue) {
		t.Fatalf("error message %q contains secret value", err.Error())
	}
}
