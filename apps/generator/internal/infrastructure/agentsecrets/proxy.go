package agentsecrets

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
}

// Request は proxy 経由の外向き呼び出し 1 回分。
type Request struct {
	Method    string
	TargetURL string
	Body      io.Reader
	Inject    Inject
}

// Do は AgentSecrets proxy 経由で req を upstream へ送る。
//
// @require ctx != nil。req.TargetURL は絶対 https URL。Inject の各欄は秘密キー名のみ。
// @ensure 成功時、戻りは proxy の HTTP 応答である。リクエストは X-AS-Target-URL・X-AS-Method を載せ、Inject.Bearer 非空なら X-AS-Inject-Bearer にそのキー名を載せる。
// @ensure Method が空のとき X-AS-Method は GET になる。
// @ensure TargetURL が空（空白のみを含む）、または scheme が https 以外のとき error を返す。
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
	if !strings.HasPrefix(targetURL, "https://") {
		return nil, fmt.Errorf("agentsecrets: TargetURL must be absolute https URL")
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

	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("agentsecrets: proxy request: %w", err)
	}
	return res, nil
}
