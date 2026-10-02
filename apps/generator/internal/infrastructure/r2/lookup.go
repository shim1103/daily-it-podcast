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

	"golang.org/x/sync/errgroup"

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
	now             func() time.Time
	retry           port.RetryReporter
}

// NewCompletedEpisodeLookup は R2 CompletedEpisodeLookup を返す。
//
// @require httpClient は非 nil（HasPair 時に検証）。accessKeyID / secretAccessKey / accountID / bucket は Composition が検証済み値を渡す。retry != nil（Composition Root の結線責務）。
// @ensure 戻りは非 nil の *CompletedEpisodeLookup（port.CompletedEpisodeLookup）。
// @invariant bucket・key・Account ID・Access Key・secret 実値を error / log へ載せない。
func NewCompletedEpisodeLookup(httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string, retry port.RetryReporter) *CompletedEpisodeLookup {
	return &CompletedEpisodeLookup{
		client:          httpClient,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		accountID:       accountID,
		bucket:          bucket,
		now:             time.Now,
		retry:           retry,
	}
}

// HasPair は所定空間に date 一致の完成ペアがあるとき true を返す。
//
// @require date は Port 契約どおり。本 Adapter は再検証しない。
// @ensure network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @ensure 対象 stem への GET は最大 maxConcurrentGets 件まで同時実行してよい。
// @invariant error に bucket / key / credential 実値を載せない。
func (l *CompletedEpisodeLookup) HasPair(ctx context.Context, date string) (bool, error) {
	if l == nil || l.client == nil {
		return false, infraErr("has_pair", fmt.Errorf("client is nil"))
	}

	keys, err := l.listObjectKeys(ctx)
	if err != nil {
		return false, err
	}

	candidates := pairStemCandidates(keys)

	matches, err := l.matchDateOnStems(ctx, candidates, date)
	if err != nil {
		return false, err
	}
	for _, matched := range matches {
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// why: docs/decisions/2026-10-02T07-47-56 — List 済み key のメモリ分類。GET だけが fan-out 対象。
func pairStemCandidates(keys []string) []string {
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

	var candidates []string
	for _, stem := range jsonStems {
		if _, ok := mp3Stems[stem]; ok {
			candidates = append(candidates, stem)
		}
	}
	return candidates
}

func (l *CompletedEpisodeLookup) matchDateOnStems(ctx context.Context, candidates []string, date string) ([]bool, error) {
	matches := make([]bool, len(candidates))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentGets)
	for i, stem := range candidates {
		g.Go(func() error {
			raw, err := l.getObject(gctx, stem+jsonExt)
			if err != nil {
				return err
			}
			episodeDate, ok := manuscriptDate(raw)
			if !ok {
				return nil
			}
			if episodeDate == date {
				matches[i] = true
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return matches, nil
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
	page, err := retryLoop(l.retry, "list_episode_keys", maxPutAttempts, func() (bool, objectKeysPage, error) {
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
	return retryLoop(l.retry, "get_episode_object", maxPutAttempts, func() (bool, []byte, error) {
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
	return buildObjectURL(l.accountID, l.bucket, objectName)
}

func (l *CompletedEpisodeLookup) listURL(continuation string) (string, error) {
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
