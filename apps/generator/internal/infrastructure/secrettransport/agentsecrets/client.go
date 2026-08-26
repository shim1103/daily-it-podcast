// Package agentsecrets は AgentSecrets proxy 経由で秘密名だけを注入する secrettransport.Client 実装を提供する。
package agentsecrets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

// what: AgentSecrets CLI の既定 proxy URL（port 8765）。
const DefaultProxyURL = "http://127.0.0.1:8765/proxy"

var _ secrettransport.Client = (*Client)(nil)

// Client は SecretRef を秘密名へ解決し、AgentSecrets proxy の X-AS-* inject header へ名前だけを載せる。
// 秘密値そのものはこの process の request へ入れない。
type Client struct {
	bindings secrettransport.BindingResolver
	http     *http.Client
	proxyURL string
}

// NewClient は AgentSecrets proxy 実装の Client を返す。
//
// @require bindings は Do 呼び出し時点で非 nil。
// @ensure httpClient が nil のとき http.DefaultClient を使う。proxyURL が空のとき DefaultProxyURL を使う。
func NewClient(bindings secrettransport.BindingResolver, httpClient *http.Client, proxyURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if proxyURL == "" {
		proxyURL = DefaultProxyURL
	}
	return &Client{bindings: bindings, http: httpClient, proxyURL: proxyURL}
}

// Do は request を AgentSecrets proxy 経由で送る。
//
// @require ctx != nil。TargetURL は host を持つ絶対 https URL。
// @ensure 未解決の有効な SecretRef、空の秘密名、または無効な非 nil Bearer は外部 I/O 前に error を返す。
// @ensure 成功時、proxy へは X-AS-Target-URL・X-AS-Method を載せる。Inject.Bearer が解決済みなら X-AS-Inject-Bearer に秘密名を載せる。
// @ensure Inject.Headers の各非空組は X-AS-Inject-Header-<ヘッダ名> に秘密名を載せる。header 名または秘密名が空ならその組は載せない。
// @ensure Inject.Form の各非空組は X-AS-Inject-Form-<フィールド名> に秘密名を載せる。フィールド名または秘密名が空ならその組は載せない。
// @ensure Inject.JSON の各非空組は X-AS-Inject-Body-<JSON path> に秘密名を載せる。path の `.` は `-` に置換する。path または秘密名が空ならその組は載せない。
// @ensure PassthroughHeaders の各非空組は同名 header として値をそのまま載せる。名前または値が空ならその組は載せない。
// @ensure Method が空のとき X-AS-Method は GET になる。
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
	method := strings.TrimSpace(request.Method)
	if method == "" {
		method = http.MethodGet
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

	httpReq, err := http.NewRequestWithContext(ctx, method, c.proxyURL, bytes.NewReader(request.Body))
	if err != nil {
		return nil, infraErr("build_request", err)
	}
	httpReq.Header.Set("X-AS-Target-URL", targetURL)
	httpReq.Header.Set("X-AS-Method", method)
	if bearerName != "" {
		httpReq.Header.Set("X-AS-Inject-Bearer", bearerName)
	}
	// why: 公式 PROXY.md は X-AS-Inject-Header-<HeaderName>。Bearer は Authorization 専用。
	setInjectHeaders(httpReq, "X-AS-Inject-Header-", headerNames, false)
	// why: 公式 PROXY.md は X-AS-Inject-Form-<field>。CLI の --form-field に対応する。
	setInjectHeaders(httpReq, "X-AS-Inject-Form-", formNames, false)
	// why: 公式 PROXY.md は JSON body 注入で dashes → dots。path の `.` を `-` にして載せる。
	setInjectHeaders(httpReq, "X-AS-Inject-Body-", jsonNames, true)
	for _, header := range request.PassthroughHeaders {
		name := strings.TrimSpace(header.Name)
		value := strings.TrimSpace(header.Value)
		if name == "" || value == "" {
			continue
		}
		httpReq.Header.Set(name, value)
	}

	res, err := c.http.Do(httpReq)
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

func setInjectHeaders(httpReq *http.Request, prefix string, pairs map[string]string, dotsToDashes bool) {
	for name, keyName := range pairs {
		// why: resolveFieldNames が trim 済み非空の field/秘密名だけを入れる。ここでの再 trim は冗長。
		if name == "" || keyName == "" {
			continue
		}
		if dotsToDashes {
			name = strings.ReplaceAll(name, ".", "-")
		}
		httpReq.Header.Set(prefix+name, keyName)
	}
}
