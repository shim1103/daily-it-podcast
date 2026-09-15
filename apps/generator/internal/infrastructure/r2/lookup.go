package r2

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.CompletedEpisodeLookup = (*CompletedEpisodeLookup)(nil)

// CompletedEpisodeLookup は R2 上の完成ペア（同一 stem の json+mp3）を表示 date で照会する。
type CompletedEpisodeLookup struct {
	client          *http.Client
	accessKeyID     string
	secretAccessKey string
	accountID       string
	bucket          string
	endpointOverride
	now func() time.Time
}

// NewCompletedEpisodeLookup は R2 CompletedEpisodeLookup を返す。
//
// @require httpClient は非 nil（HasPair 時に検証）。accessKeyID / secretAccessKey / accountID / bucket は Composition が検証済み値を渡す。
// @ensure 戻りは非 nil の *CompletedEpisodeLookup（port.CompletedEpisodeLookup）。
// @invariant bucket・key・Account ID・Access Key・secret 実値を error / log へ載せない。
func NewCompletedEpisodeLookup(httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string) *CompletedEpisodeLookup {
	return &CompletedEpisodeLookup{
		client:          httpClient,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		accountID:       accountID,
		bucket:          bucket,
		now:             time.Now,
	}
}

// WithEndpointBase は本番 Account ID host の代わりに使う S3 互換 endpoint base を返す。
// local S3 gate peer 向け。空のままなら本番 URL。
//
// @require base は scheme+host を持つ絶対 URL（trailing slash 可）。
// @ensure 戻りは endpointBase を持つ別 *CompletedEpisodeLookup。受信者は変更しない。
func (l *CompletedEpisodeLookup) WithEndpointBase(base string) *CompletedEpisodeLookup {
	if l == nil {
		return nil
	}
	cp := *l
	cp.endpointOverride = cp.withBase(base)
	return &cp
}

// HasPair は所定空間に date 一致の完成ペアがあるとき true を返す。
//
// @require date は YYYY-MM-DD（Port 契約。本 Adapter は再検証しない）。
// @ensure 同一 stem の json+mp3 があり json の date が一致するとき true。片方・無し・不一致は false。
// @ensure network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @invariant storage 固有 id・MIME・vendor 型を露出せず、error に bucket / key / credential 実値を載せない。
func (l *CompletedEpisodeLookup) HasPair(ctx context.Context, date string) (bool, error) {
	if l == nil || l.client == nil {
		return false, infraErr("has_pair", fmt.Errorf("client is nil"))
	}

	keys, err := l.listObjectKeys(ctx)
	if err != nil {
		return false, err
	}

	mp3Stems := make(map[string]struct{})
	var jsonStems []string
	for _, key := range keys {
		switch {
		case strings.HasSuffix(key, jsonExt):
			jsonStems = append(jsonStems, strings.TrimSuffix(key, jsonExt))
		case strings.HasSuffix(key, mp3Ext):
			mp3Stems[strings.TrimSuffix(key, mp3Ext)] = struct{}{}
		}
	}

	for _, stem := range jsonStems {
		if _, ok := mp3Stems[stem]; !ok {
			continue
		}
		raw, err := l.getObject(ctx, stem+jsonExt)
		if err != nil {
			return false, err
		}
		episodeDate, ok := manuscriptDate(raw)
		if !ok {
			continue
		}
		if episodeDate == date {
			return true, nil
		}
	}
	return false, nil
}

func (l *CompletedEpisodeLookup) listObjectKeys(ctx context.Context) ([]string, error) {
	var out []string
	continuation := ""
	for {
		keys, next, err := l.listObjectKeysPage(ctx, continuation)
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		if next == "" {
			return out, nil
		}
		continuation = next
	}
}

type objectKeysPage struct {
	keys []string
	next string
}

func (l *CompletedEpisodeLookup) listObjectKeysPage(ctx context.Context, continuation string) ([]string, string, error) {
	page, err := retryLoop(maxPutAttempts, func() (bool, objectKeysPage, error) {
		return l.listObjectKeysPageOnce(ctx, continuation)
	})
	if err != nil {
		return nil, "", err
	}
	return page.keys, page.next, nil
}

func (l *CompletedEpisodeLookup) listObjectKeysPageOnce(ctx context.Context, continuation string) (retryable bool, page objectKeysPage, err error) {
	target, err := l.listURL(continuation)
	if err != nil {
		return false, objectKeysPage{}, infraErr("build_url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, objectKeysPage{}, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, l.accessKeyID, l.secretAccessKey, l.now()); err != nil {
		return false, objectKeysPage{}, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := l.client.Do(req)
	if err != nil {
		return true, objectKeysPage{}, infraErr("do", fmt.Errorf("request failed"))
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, objectKeysPage{}, infraErr("read", fmt.Errorf("read body failed"))
	}
	// why: ListObjectsV2 は正常時常に 200 固定（S3 仕様、R2 も S3 互換 API 準拠）。201/204 等の他 2xx は
	// PUT/POST 系専用で List には本来出現しないため、範囲判定ではなく 200 固定で判定する。
	if res.StatusCode == http.StatusOK {
		parsed, parseErr := parseListBucketResult(body)
		if parseErr != nil {
			return false, objectKeysPage{}, infraErr("parse_list", fmt.Errorf("list parse failed"))
		}
		out := make([]string, 0, len(parsed.Contents))
		for _, c := range parsed.Contents {
			if c.Key != "" {
				out = append(out, c.Key)
			}
		}
		nextToken := ""
		if parsed.IsTruncated {
			nextToken = parsed.NextContinuationToken
		}
		return false, objectKeysPage{keys: out, next: nextToken}, nil
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return true, objectKeysPage{}, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	return false, objectKeysPage{}, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
}

func (l *CompletedEpisodeLookup) getObject(ctx context.Context, objectName string) ([]byte, error) {
	return retryLoop(maxPutAttempts, func() (bool, []byte, error) {
		return l.getObjectOnce(ctx, objectName)
	})
}

func (l *CompletedEpisodeLookup) getObjectOnce(ctx context.Context, objectName string) (retryable bool, body []byte, err error) {
	target, err := l.objectURL(objectName)
	if err != nil {
		return false, nil, infraErr("build_url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, nil, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, l.accessKeyID, l.secretAccessKey, l.now()); err != nil {
		return false, nil, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := l.client.Do(req)
	if err != nil {
		return true, nil, infraErr("do", fmt.Errorf("request failed"))
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return false, nil, infraErr("read", fmt.Errorf("read body failed"))
	}
	// why: GetObject も正常時常に 200 固定。List 同様に範囲判定ではなく 200 固定で判定する。
	if res.StatusCode == http.StatusOK {
		return false, raw, nil
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return true, nil, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	return false, nil, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
}

func (l *CompletedEpisodeLookup) objectURL(objectName string) (string, error) {
	if l.endpointBase != "" {
		return buildObjectURLFromBase(l.endpointBase, l.bucket, objectName)
	}
	return buildObjectURL(l.accountID, l.bucket, objectName)
}

func (l *CompletedEpisodeLookup) listURL(continuation string) (string, error) {
	if l.endpointBase != "" {
		return buildListURLFromBase(l.endpointBase, l.bucket, continuation)
	}
	if l.accountID == "" || l.bucket == "" {
		return "", fmt.Errorf("endpoint incomplete")
	}
	q := url.Values{}
	q.Set("list-type", "2")
	if continuation != "" {
		q.Set("continuation-token", continuation)
	}
	u := &url.URL{
		Scheme:   "https",
		Host:     l.accountID + ".r2.cloudflarestorage.com",
		Path:     "/" + l.bucket,
		RawQuery: q.Encode(),
	}
	return u.String(), nil
}

type listBucketResultXML struct {
	XMLName  xml.Name `xml:"ListBucketResult"`
	Contents []struct {
		Key string `xml:"Key"`
	} `xml:"Contents"`
	IsTruncated           bool   `xml:"IsTruncated"`
	NextContinuationToken string `xml:"NextContinuationToken"`
}

func parseListBucketResult(raw []byte) (listBucketResultXML, error) {
	var parsed listBucketResultXML
	if err := xml.Unmarshal(raw, &parsed); err != nil {
		return listBucketResultXML{}, err
	}
	return parsed, nil
}

// manuscriptDate は原稿 JSON の date field だけを読む。schema 全体は検証しない。
func manuscriptDate(raw []byte) (string, bool) {
	var doc struct {
		Date string `json:"date"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", false
	}
	date := strings.TrimSpace(doc.Date)
	if date == "" {
		return "", false
	}
	return date, true
}
