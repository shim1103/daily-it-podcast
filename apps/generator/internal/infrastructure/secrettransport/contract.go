// Package secrettransport は秘密値を参照で注入する外向き HTTP transport の契約を提供する。
package secrettransport

import (
	"context"
	"net/http"
)

type secretRefToken struct {
	_ byte
}

// SecretRef は秘密値を公開せずに参照する不透明な識別子である。
// ゼロ値は無効であり、比較できる。
type SecretRef struct {
	token *secretRefToken
}

// NewSecretRef は呼び出すたびに異なる有効な SecretRef を返す。
func NewSecretRef() SecretRef {
	return SecretRef{token: &secretRefToken{}}
}

// Header は秘密値を含まない素通し HTTP header である。
type Header struct {
	Name  string
	Value string
}

// FieldInjection は field へ注入する秘密値の参照である。
type FieldInjection struct {
	Field  string
	Secret SecretRef
}

// Inject は bearer、header、form、JSON field への秘密値注入を表す。
// Headers、Form、JSON は index 順を保存する。
type Inject struct {
	Bearer  *SecretRef
	Headers []FieldInjection
	Form    []FieldInjection
	JSON    []FieldInjection
}

// Request は外向き HTTP request の秘密値を含まない入力である。
// TargetURL は host を持つ絶対 HTTPS URL でなければならず、Method が空なら GET になる。
// Inject の各 slice は index 順を保存し、空の Field または無効な Secret を持つ entry は skip する。
type Request struct {
	Method             string
	TargetURL          string
	Body               []byte
	PassthroughHeaders []Header
	Inject             Inject
}

// BindingResolver は SecretRef を秘密名へ解決する。
type BindingResolver interface {
	ResolveSecret(ref SecretRef) (name string, ok bool)
}

// Client は解決済みの秘密値を外向き HTTP request へ注入する。
type Client interface {
	// Do は request を送る。
	// 未解決の有効な SecretRef、または無効な非 nil Bearer は外部 I/O 前に error を返す。
	// error には秘密値、request body、stdin、child stderr を含めない。
	Do(ctx context.Context, request Request) (*http.Response, error)
}
