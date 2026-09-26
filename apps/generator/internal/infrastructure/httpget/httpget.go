// Package httpget は媒体非依存の HTTP GET+retry と共有 HTML 正規化を提供する。
// ItemSource / SourceID は知らない。媒体 Adapter が写像・URL・方言 parse を所有する。
package httpget

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
)

// GetWithRetry は GET を実行し body を返す。
// client.Do error / 5xx は 1 回だけ即再試行（backoff なし）。4xx（429 含む）/ 読み取り失敗は即 return。
//
// @require client != nil。url は絶対 URL。
// @ensure 成功時は応答 body。失敗時は非 nil error（Adapter 側で infraErr へ包む）。
func GetWithRetry(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	body, retryable, err := get(ctx, client, url)
	if err == nil {
		return body, nil
	}
	if !retryable {
		return nil, err
	}
	// why: 2 回目は retryable を問わず打ち切る（再試行は 1 回だけ）。
	body, _, err = get(ctx, client, url)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// get は GET を 1 回実行し body を返す。
// 2 つめの戻り値は再試行してよい失敗（client.Do error / 5xx）かどうか。
func get(ctx context.Context, client *http.Client, url string) (body []byte, retryable bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = res.Body.Close() }()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}
	if res.StatusCode >= 500 {
		return nil, true, fmt.Errorf("status %d", res.StatusCode)
	}
	if res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("status %d", res.StatusCode)
	}
	return body, false, nil
}

// pTagReplacer は <p> 段落区切りを改行へ変換する。
// why: NormalizeHTML は呼び出し元の 1 run で高頻度に呼ばれうる。Replacer 生成を package 初期化へ巻き上げる。
var pTagReplacer = strings.NewReplacer("<p>", "\n", "<P>", "\n")

// NormalizeHTML は HTML entity を unescape し、<p> を改行へ、他タグを除去する軽い正規化。
// token 効率が目的で full readability はしない。
func NormalizeHTML(s string) string {
	if s == "" {
		return ""
	}
	s = pTagReplacer.Replace(s)
	s = stripTags(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

// stripTags は "<" と ">" で囲まれた token を除去する。
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}
