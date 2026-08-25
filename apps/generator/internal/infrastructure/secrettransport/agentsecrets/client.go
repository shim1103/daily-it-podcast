// Package agentsecrets は AgentSecrets proxy 経由で秘密名だけを注入する secrettransport.Client 実装を提供する。
package agentsecrets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	infraas "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

var _ secrettransport.Client = (*Client)(nil)

// Client は SecretRef を秘密名へ解決し、AgentSecrets proxy の X-AS-* inject header へ名前だけを載せる。
// 秘密値そのものはこの process の request へ入れない。
type Client struct {
	bindings secrettransport.BindingResolver
	proxy    *infraas.Client
}

// NewClient は AgentSecrets proxy 実装の Client を返す。
//
// @require bindings は Do 呼び出し時点で非 nil。
// @ensure httpClient が nil のとき proxy 側が http.DefaultClient を使う。proxyURL が空のとき DefaultProxyURL を使う。
func NewClient(bindings secrettransport.BindingResolver, httpClient *http.Client, proxyURL string) *Client {
	return &Client{
		bindings: bindings,
		proxy:    &infraas.Client{HTTP: httpClient, ProxyURL: proxyURL},
	}
}

// Do は request を AgentSecrets proxy 経由で送る。
//
// @require ctx != nil。TargetURL は host を持つ絶対 https URL。
// @ensure 未解決の有効な SecretRef、空の秘密名、または無効な非 nil Bearer は外部 I/O 前に error を返す。
// @ensure error message に秘密値・request body・proxy response body を含めない。
func (c *Client) Do(ctx context.Context, request secrettransport.Request) (*http.Response, error) {
	if c == nil {
		return nil, infraErr("do", fmt.Errorf("client is nil"))
	}
	if ctx == nil {
		return nil, infraErr("do", fmt.Errorf("ctx is nil"))
	}
	if c.bindings == nil {
		return nil, infraErr("do", fmt.Errorf("bindings is nil"))
	}
	targetURL := strings.TrimSpace(request.TargetURL)
	if err := secrettransport.ValidateAbsoluteHTTPSURL(targetURL); err != nil {
		return nil, infraErr("do", err)
	}

	var bearerName string
	if request.Inject.Bearer != nil {
		name, err := c.resolveSecretName(*request.Inject.Bearer)
		if err != nil {
			return nil, infraErr("bearer", err)
		}
		bearerName = name
	}

	headerNames, err := c.resolveFieldNames(request.Inject.Headers)
	if err != nil {
		return nil, infraErr("headers", err)
	}
	formNames, err := c.resolveFieldNames(request.Inject.Form)
	if err != nil {
		return nil, infraErr("form", err)
	}
	jsonNames, err := c.resolveFieldNames(request.Inject.JSON)
	if err != nil {
		return nil, infraErr("json", err)
	}

	passthrough := make(map[string]string, len(request.PassthroughHeaders))
	for _, header := range request.PassthroughHeaders {
		name := strings.TrimSpace(header.Name)
		value := strings.TrimSpace(header.Value)
		if name == "" || value == "" {
			continue
		}
		passthrough[name] = value
	}

	res, err := c.proxy.Do(ctx, infraas.Request{
		Method:             request.Method,
		TargetURL:          targetURL,
		Body:               bytes.NewReader(request.Body),
		PassthroughHeaders: passthrough,
		Inject: infraas.Inject{
			Bearer:  bearerName,
			Headers: headerNames,
			Form:    formNames,
			Body:    jsonNames,
		},
	})
	if err != nil {
		return nil, infraErr("request", err)
	}
	return res, nil
}

func (c *Client) resolveSecretName(ref secrettransport.SecretRef) (string, error) {
	name, ok := c.bindings.ResolveSecret(ref)
	if !ok || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("secret binding is unresolved")
	}
	return strings.TrimSpace(name), nil
}

func (c *Client) resolveFieldNames(entries []secrettransport.FieldInjection) (map[string]string, error) {
	out := make(map[string]string)
	for _, entry := range entries {
		field := strings.TrimSpace(entry.Field)
		if field == "" || entry.Secret == (secrettransport.SecretRef{}) {
			continue
		}
		name, err := c.resolveSecretName(entry.Secret)
		if err != nil {
			return nil, err
		}
		out[field] = name
	}
	return out, nil
}
