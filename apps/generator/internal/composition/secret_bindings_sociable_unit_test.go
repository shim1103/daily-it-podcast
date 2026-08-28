package composition

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
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
