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

func TestGeneratorSecretBindings_resolvesCursorAPIKeySecretName(t *testing.T) {
	t.Parallel()

	// Given/When: 表の Cursor API key binding
	name, ok := generatorSecretBindings.ResolveSecret(cursorAPIKeySecret)

	// Then: Cursor 秘密名へ結線される
	if !ok {
		t.Fatal("ResolveSecret(cursorAPIKeySecret) = false, want true")
	}
	if name != secretnames.CursorAPIKeyName {
		t.Fatalf("secret name = %q, want %q", name, secretnames.CursorAPIKeyName)
	}
}

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
