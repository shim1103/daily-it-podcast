// Package processenv は process environment から秘密値を解決し、外向き HTTP request へ実値そのものを注入する。
package processenv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

var _ secrettransport.Client = (*Client)(nil)

// Client は process environment 由来の秘密値を外向き HTTP request へ直接注入する。
// AgentSecrets proxy のような中間 proxy を経由せず、実値をこの process 内で注入して送信する。
type Client struct {
	bindings  secrettransport.BindingResolver
	http      *http.Client
	lookupEnv func(key string) (string, bool)
}

// NewClient は process-env 実装の Client を返す。
//
// @require bindings は非 nil。
// @ensure httpClient が nil のとき http.DefaultClient を使う。lookupEnv が nil のとき os.LookupEnv を使う。
func NewClient(bindings secrettransport.BindingResolver, httpClient *http.Client, lookupEnv func(key string) (string, bool)) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	return &Client{bindings: bindings, http: httpClient, lookupEnv: lookupEnv}
}

// Do は request を送る。
//
// @require ctx != nil。TargetURL は host を持つ絶対 https URL。
// @ensure 未解決の有効な SecretRef、無効な非 nil Bearer、または解決した secret 値が空のとき、外部 I/O 前に error を返す。
// @ensure error message に秘密値・request body を含めない。
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

	var bearerValue string
	if request.Inject.Bearer != nil {
		value, err := c.resolveSecretValue(*request.Inject.Bearer)
		if err != nil {
			return nil, infraErr("bearer", err)
		}
		bearerValue = value
	}

	headerValues, err := c.resolveFieldInjections(request.Inject.Headers)
	if err != nil {
		return nil, infraErr("headers", err)
	}
	formValues, err := c.resolveFieldInjections(request.Inject.Form)
	if err != nil {
		return nil, infraErr("form", err)
	}
	jsonValues, err := c.resolveFieldInjections(request.Inject.JSON)
	if err != nil {
		return nil, infraErr("json", err)
	}

	body := request.Body
	if len(formValues) > 0 {
		body, err = mergeForm(body, formValues)
		if err != nil {
			return nil, infraErr("merge_form", err)
		}
	}
	if len(jsonValues) > 0 {
		body, err = mergeJSON(body, jsonValues)
		if err != nil {
			return nil, infraErr("merge_json", err)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, infraErr("build_request", err)
	}
	for _, header := range request.PassthroughHeaders {
		name := strings.TrimSpace(header.Name)
		if name == "" {
			continue
		}
		httpReq.Header.Set(name, header.Value)
	}
	if request.Inject.Bearer != nil {
		httpReq.Header.Set("Authorization", "Bearer "+bearerValue)
	}
	for _, resolved := range headerValues {
		httpReq.Header.Set(resolved.field, resolved.value)
	}

	res, err := c.http.Do(httpReq)
	if err != nil {
		return nil, infraErr("request", err)
	}
	return res, nil
}

// resolveSecretValue は ref を secret 名へ解決し、process environment から実値を取得する。
func (c *Client) resolveSecretValue(ref secrettransport.SecretRef) (string, error) {
	name, ok := c.bindings.ResolveSecret(ref)
	if !ok || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("secret binding is unresolved")
	}
	value, ok := c.lookupEnv(name)
	if !ok || value == "" {
		return "", fmt.Errorf("secret is unset")
	}
	return value, nil
}

// resolvedField は FieldInjection 1 件分の Field と解決済み実値を、Inject の index 順のまま保持する。
type resolvedField struct {
	field string
	value string
}

// resolveFieldInjections は entries を index 順を保った resolvedField 列へ解決する。
// Field が空、または Secret がゼロ値の entry は skip する。
func (c *Client) resolveFieldInjections(entries []secrettransport.FieldInjection) ([]resolvedField, error) {
	out := make([]resolvedField, 0, len(entries))
	for _, entry := range entries {
		field := strings.TrimSpace(entry.Field)
		if field == "" || entry.Secret == (secrettransport.SecretRef{}) {
			continue
		}
		value, err := c.resolveSecretValue(entry.Secret)
		if err != nil {
			return nil, err
		}
		out = append(out, resolvedField{field: field, value: value})
	}
	return out, nil
}

// mergeForm は既存の application/x-www-form-urlencoded body へ values を index 順に merge する。
func mergeForm(body []byte, values []resolvedField) ([]byte, error) {
	form, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	for _, resolved := range values {
		form.Set(resolved.field, resolved.value)
	}
	return []byte(form.Encode()), nil
}

// mergeJSON は既存の JSON body へ dot-path で values を index 順に挿入する。
// path segment が非負整数なら配列 index、それ以外は object key として辿る。
func mergeJSON(body []byte, values []resolvedField) ([]byte, error) {
	var decoded any
	if len(bytes.TrimSpace(body)) == 0 {
		decoded = map[string]any{}
	} else if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, err
	}
	for _, resolved := range values {
		var err error
		decoded, err = setJSONPath(decoded, strings.Split(resolved.field, "."), resolved.value)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(decoded)
}

// setJSONPath は node 内の segments が指す位置へ value をセットした node を返す。
func setJSONPath(node any, segments []string, value string) (any, error) {
	if len(segments) == 0 {
		return value, nil
	}
	head := segments[0]
	rest := segments[1:]

	if index, err := strconv.Atoi(head); err == nil && index >= 0 {
		arr, ok := node.([]any)
		if !ok {
			return nil, fmt.Errorf("json path segment %q expects array", head)
		}
		if index >= len(arr) {
			return nil, fmt.Errorf("json path segment %q index out of range", head)
		}
		child, err := setJSONPath(arr[index], rest, value)
		if err != nil {
			return nil, err
		}
		arr[index] = child
		return arr, nil
	}

	obj, ok := node.(map[string]any)
	if !ok {
		obj = map[string]any{}
	}
	child, err := setJSONPath(obj[head], rest, value)
	if err != nil {
		return nil, err
	}
	obj[head] = child
	return obj, nil
}
