package config

import (
	"fmt"
	"strings"
	"testing"
)

// dummySecretRaw はformat制約だけ満たし秘匿価値のないtest用raw値である。
const dummySecretRaw = "dummy-secret-raw-value"

func TestNewSecret_Reveal_returnsRawValue_whenConstructedFromRaw(t *testing.T) {
	t.Parallel()

	// Given: raw値から生成したSecret
	secret := newSecret(dummySecretRaw)

	// When: Revealする
	revealed := secret.Reveal()

	// Then: 投入したraw値と一致する
	if revealed != dummySecretRaw {
		t.Fatal("Reveal() がraw値を返さなかった")
	}
}

func TestNewSecret_String_returnsRedaction_whenFormatted(t *testing.T) {
	t.Parallel()

	// Given: raw値から生成したSecret
	secret := newSecret(dummySecretRaw)

	// When: String() 直接と fmt の default verb 経由で表示する
	viaMethod := secret.String()
	viaVerbV := fmt.Sprintf("%v", secret)

	// Then: いずれも固定のredaction表記でありraw値を含まない
	if viaMethod != secretRedaction {
		t.Fatal("String() がredaction表記を返さなかった")
	}
	if strings.Contains(viaMethod, dummySecretRaw) || strings.Contains(viaVerbV, dummySecretRaw) {
		t.Fatal("String系出力がraw値を含んだ")
	}
	if viaVerbV != secretRedaction {
		t.Fatal("default verb経由の表示がredaction表記を返さなかった")
	}
}

func TestNewSecret_GoString_returnsRedaction_whenFormattedWithHashV(t *testing.T) {
	t.Parallel()

	// Given: raw値から生成したSecret
	secret := newSecret(dummySecretRaw)

	// When: GoString / %#v で表示する
	viaMethod := secret.GoString()
	viaVerbHashV := fmt.Sprintf("%#v", secret)

	// Then: いずれも固定のredaction表記でありraw値を含まない
	if viaMethod != secretRedaction {
		t.Fatal("GoString() がredaction表記を返さなかった")
	}
	if strings.Contains(viaVerbHashV, dummySecretRaw) {
		t.Fatal("hash-v verb経由の表示がraw値を含んだ")
	}
	if viaVerbHashV != secretRedaction {
		t.Fatal("hash-v verb経由の表示がredaction表記を返さなかった")
	}
}

func TestNewSecret_satisfiesSecretInterface_whenAssigned(t *testing.T) {
	t.Parallel()

	// Given / When: newSecretの戻りをSecret型へ代入する
	var s Secret = newSecret(dummySecretRaw)

	// Then: nilではない
	if s == nil {
		t.Fatal("newSecret() がSecretを返さなかった")
	}
}
