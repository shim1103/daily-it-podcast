package composition

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/agentsecrets"
)

func TestAgentsecretsSecretTransportClient_sendsBoundSecretNameToProxy_whenBearerIsCompositionRef(t *testing.T) {
	// Given: Composition binding（getXAPIKeySecret）と proxy probe へ結線した local Client
	var gotBearer string
	var gotTargetURL string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBearer = r.Header.Get("X-AS-Inject-Bearer")
		gotTargetURL = r.Header.Get("X-AS-Target-URL")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(proxy.Close)
	client := agentsecrets.NewClient(generatorSecretBindings, proxy.Client(), proxy.URL)

	// When: Composition の SecretRef を Bearer 注入して Do する
	res, err := client.Do(context.Background(), secrettransport.Request{
		TargetURL: "https://api.example.com/v1/posts",
		Inject:    secrettransport.Inject{Bearer: &getXAPIKeySecret},
	})

	// Then: X-AS-Inject-Bearer に binding の秘密名が載る（値ではない）
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}
	defer func() { _ = res.Body.Close() }()
	if gotBearer != secretnames.GetXAPIKeyName {
		t.Fatalf("X-AS-Inject-Bearer = %q, want %q", gotBearer, secretnames.GetXAPIKeyName)
	}
	if gotTargetURL != "https://api.example.com/v1/posts" {
		t.Fatalf("X-AS-Target-URL = %q", gotTargetURL)
	}
}

func TestAgentsecretsSecretTransportClient_returnsNonNilClient(t *testing.T) {
	t.Parallel()

	// Given/When: 既定の local factory を呼ぶ
	client := agentsecretsSecretTransportClient()

	// Then: Adapter へ渡せる secrettransport.Client が返る
	if client == nil {
		t.Fatal("agentsecretsSecretTransportClient() = nil, want non-nil")
	}
}

func TestProcessenvSecretTransportClient_returnsNonNilClient(t *testing.T) {
	t.Parallel()

	// Given/When: production 既定 factory を呼ぶ（退行防止）
	client := processenvSecretTransportClient()

	// Then: 引き続き結線可能
	if client == nil {
		t.Fatal("processenvSecretTransportClient() = nil, want non-nil")
	}
}
