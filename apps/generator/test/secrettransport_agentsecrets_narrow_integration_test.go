// Scope: Narrow Integration
// 実物境界: secrettransport/agentsecrets.Client が送信する AgentSecrets proxy request（test proxy server）
// Double: BindingResolver は Composition と同型の in-memory map。本番 keychain / 実 proxy は使わない。
// @require controllable な proxy probe。SecretRef は秘密名へ解決する（値は process に載せない）。
// @ensure 代表 success（Bearer）で秘密名が X-AS-Inject-Bearer へ届く。Header/Form/JSON の詳細 mapping は sociable unit が所有する。
// @ensure 未解決 secret は proxy I/O 前に失敗し、proxy は呼ばれない。
// @invariant secret 値・request body は error message へ出ない。
package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/agentsecrets"
)

func TestAgentSecretsSecretTransportClient_injectsBearerNameViaProxyHeader_whenSecretResolved(t *testing.T) {
	// Given: 秘密名へ解決する Bearer SecretRef（値は持たない）
	const (
		bearerName  = "NARROW_AS_BEARER_KEY"
		bearerValue = "narrow-bearer-must-not-appear"
	)
	bearerRef := secrettransport.NewSecretRef()
	bindings := narrowBindings{bearerRef: bearerName}

	var gotBearer string
	var gotTargetURL string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBearer = r.Header.Get("X-AS-Inject-Bearer")
		gotTargetURL = r.Header.Get("X-AS-Target-URL")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(proxy.Close)
	client := agentsecrets.NewClient(bindings, proxy.Client(), proxy.URL)

	// When: Bearer を注入した request を送る
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    secrettransport.Inject{Bearer: &bearerRef},
	})

	// Then: proxy へ秘密名だけが届く（Header/Form/JSON mapping は sociable unit 側）
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	_ = res.Body.Close()
	if gotBearer != bearerName {
		t.Fatalf("X-AS-Inject-Bearer = %q, want %q", gotBearer, bearerName)
	}
	if strings.Contains(gotBearer, bearerValue) {
		t.Fatalf("inject header contains secret value: %q", gotBearer)
	}
	if gotTargetURL != "https://api.example.com/v1/posts" {
		t.Fatalf("X-AS-Target-URL = %q", gotTargetURL)
	}
}

func TestAgentSecretsSecretTransportClient_failsBeforeProxyCall_whenSecretUnresolved(t *testing.T) {
	// Given: bindings に無い SecretRef と、混入させたくない識別文字列
	const mustNotLeak = "narrow-as-must-not-leak-value"
	ref := secrettransport.NewSecretRef()
	bindings := narrowBindings{}

	called := false
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(proxy.Close)
	client := agentsecrets.NewClient(bindings, proxy.Client(), proxy.URL)

	// When: 未解決 secret を Bearer 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Body:      []byte(mustNotLeak),
		Inject:    secrettransport.Inject{Bearer: &ref},
	})

	// Then: proxy I/O 前に失敗し、secret/body は error に出ない
	if err == nil {
		t.Fatal("Do() error = nil, want non-nil")
	}
	if res != nil {
		t.Fatalf("Do() response = %v, want nil", res)
	}
	if called {
		t.Fatal("proxy was called, want not called")
	}
	if strings.Contains(err.Error(), mustNotLeak) {
		t.Fatalf("error message %q contains secret/body value", err.Error())
	}
}
