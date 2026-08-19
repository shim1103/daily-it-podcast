package agentsecrets

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// what: AgentSecrets CLI の既定 proxy URL（port 8765）。
const DefaultProxyURL = "http://127.0.0.1:8765/proxy"

// why: env 注入だと秘密値が子 process の environ に載るため、キー名だけの proxy 注入にする。
// Client は AgentSecrets proxy 経由で外向き HTTP を送る。
type Client struct {
	HTTP     *http.Client
	ProxyURL string
}

// Inject は proxy 側注入用の秘密キー名。
type Inject struct {
	Bearer string
	// Headers は custom header 注入。key は upstream header 名、value は秘密キー名。
	Headers map[string]string
	// Form は form body 注入。key はフィールド名、value は秘密キー名。
	Form map[string]string
	// Body は JSON body 注入。key は JSON path（dot 区切り）、value は秘密キー名。
	Body map[string]string
}

// Request は proxy 経由の外向き呼び出し 1 回分。
type Request struct {
	Method    string
	TargetURL string
	Body      io.Reader
	// PassthroughHeaders は非秘密 header の素通し（Content-Type、導出した Authorization など）。
	PassthroughHeaders map[string]string
	Inject             Inject
}

// Do は AgentSecrets proxy 経由で req を upstream へ送る。
//
// @require ctx != nil。req.TargetURL は host 付きの絶対 https URL。Inject の各欄は秘密キー名のみ。
// @ensure 成功時、戻りは proxy の HTTP 応答である。リクエストは X-AS-Target-URL・X-AS-Method を載せ、Inject.Bearer 非空なら X-AS-Inject-Bearer にそのキー名を載せる。
// @ensure Inject.Headers の各非空組は X-AS-Inject-Header-<ヘッダ名> にキー名を載せる。header 名またはキー名が空ならその組は載せない。
// @ensure Inject.Form の各非空組は X-AS-Inject-Form-<フィールド名> にキー名を載せる。フィールド名またはキー名が空ならその組は載せない。
// @ensure Inject.Body の各非空組は X-AS-Inject-Body-<JSON path> にキー名を載せる。path の `.` は `-` に置換する。path またはキー名が空ならその組は載せない。
// @ensure Request.PassthroughHeaders の各非空組は同名 header として値をそのまま載せる。名前または値が空ならその組は載せない。
// @ensure Method が空のとき X-AS-Method は GET になる。
// @ensure TargetURL が空（空白のみを含む）、scheme が https 以外、または host が空のとき error を返す。
func (c *Client) Do(ctx context.Context, req Request) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("agentsecrets: client is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("agentsecrets: ctx is nil")
	}
	targetURL := strings.TrimSpace(req.TargetURL)
	if targetURL == "" {
		return nil, fmt.Errorf("agentsecrets: TargetURL is empty")
	}
	u, err := url.Parse(targetURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("agentsecrets: TargetURL must be absolute https URL with host")
	}
	method := strings.TrimSpace(req.Method)
	if method == "" {
		method = http.MethodGet
	}

	proxyURL := c.ProxyURL
	if proxyURL == "" {
		proxyURL = DefaultProxyURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, proxyURL, req.Body)
	if err != nil {
		return nil, fmt.Errorf("agentsecrets: build proxy request: %w", err)
	}
	httpReq.Header.Set("X-AS-Target-URL", targetURL)
	httpReq.Header.Set("X-AS-Method", method)
	if name := strings.TrimSpace(req.Inject.Bearer); name != "" {
		httpReq.Header.Set("X-AS-Inject-Bearer", name)
	}
	// why: 公式 PROXY.md は X-AS-Inject-Header-<HeaderName>。Bearer は Authorization 専用。
	setInjectHeaders(httpReq, "X-AS-Inject-Header-", req.Inject.Headers, false)
	// why: 公式 PROXY.md は X-AS-Inject-Form-<field>。CLI の --form-field に対応する。
	setInjectHeaders(httpReq, "X-AS-Inject-Form-", req.Inject.Form, false)
	// why: 公式 PROXY.md は JSON body 注入で dashes → dots。path の `.` を `-` にして載せる。
	setInjectHeaders(httpReq, "X-AS-Inject-Body-", req.Inject.Body, true)
	for headerName, value := range req.PassthroughHeaders {
		headerName = strings.TrimSpace(headerName)
		value = strings.TrimSpace(value)
		if headerName == "" || value == "" {
			continue
		}
		httpReq.Header.Set(headerName, value)
	}

	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("agentsecrets: proxy request: %w", err)
	}
	return res, nil
}

func setInjectHeaders(httpReq *http.Request, prefix string, pairs map[string]string, dotsToDashes bool) {
	for name, keyName := range pairs {
		name = strings.TrimSpace(name)
		keyName = strings.TrimSpace(keyName)
		if name == "" || keyName == "" {
			continue
		}
		if dotsToDashes {
			name = strings.ReplaceAll(name, ".", "-")
		}
		httpReq.Header.Set(prefix+name, keyName)
	}
}
