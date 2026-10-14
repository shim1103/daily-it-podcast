package r2

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// objectStore は R2 SigV4 の list / get / delete を共有する。
type objectStore struct {
	client          *http.Client
	accessKeyID     string
	secretAccessKey string
	accountID       string
	bucket          string
	now             func() time.Time
	retry           port.RetryReporter
}

func (s *objectStore) listObjectKeys(ctx context.Context) ([]string, error) {
	var out []string
	continuation := ""
	for {
		keys, next, err := s.listObjectKeysPage(ctx, continuation)
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

func (s *objectStore) listObjectKeysPage(ctx context.Context, continuation string) ([]string, string, error) {
	page, err := retryLoop(s.retry, "list_episode_keys", maxPutAttempts, func() (bool, objectKeysPage, error) {
		return s.listObjectKeysPageOnce(ctx, continuation)
	})
	if err != nil {
		return nil, "", err
	}
	return page.keys, page.next, nil
}

func (s *objectStore) listObjectKeysPageOnce(ctx context.Context, continuation string) (retryable bool, page objectKeysPage, err error) {
	target, err := s.listURL(continuation)
	if err != nil {
		return false, objectKeysPage{}, infraErr("build_url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, objectKeysPage{}, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, s.accessKeyID, s.secretAccessKey, s.now()); err != nil {
		return false, objectKeysPage{}, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := s.client.Do(req)
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

func (s *objectStore) getObject(ctx context.Context, objectName string) ([]byte, error) {
	return retryLoop(s.retry, "get_episode_object", maxPutAttempts, func() (bool, []byte, error) {
		return s.getObjectOnce(ctx, objectName)
	})
}

func (s *objectStore) getObjectOnce(ctx context.Context, objectName string) (retryable bool, body []byte, err error) {
	target, err := s.objectURL(objectName)
	if err != nil {
		return false, nil, infraErr("build_url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, nil, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, s.accessKeyID, s.secretAccessKey, s.now()); err != nil {
		return false, nil, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := s.client.Do(req)
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

func (s *objectStore) deleteObject(ctx context.Context, objectName string) error {
	_, err := retryLoop(s.retry, "delete_episode_object", maxPutAttempts, func() (bool, struct{}, error) {
		retryable, opErr := s.deleteObjectOnce(ctx, objectName)
		return retryable, struct{}{}, opErr
	})
	return err
}

func (s *objectStore) deleteObjectOnce(ctx context.Context, objectName string) (retryable bool, err error) {
	target, err := s.objectURL(objectName)
	if err != nil {
		return false, infraErr("build_url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, target, nil)
	if err != nil {
		return false, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4(req, nil, s.accessKeyID, s.secretAccessKey, s.now()); err != nil {
		return false, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := s.client.Do(req)
	if err != nil {
		return true, infraErr("do", fmt.Errorf("request failed"))
	}
	defer func() { _ = res.Body.Close() }()
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		return false, infraErr("read", fmt.Errorf("read body failed"))
	}
	// why: DeleteObject は 204 が定番。冪等 cleanup のため 404 も成功。
	if (res.StatusCode >= 200 && res.StatusCode < 300) || res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return true, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	return false, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
}

func (s *objectStore) objectURL(objectName string) (string, error) {
	return buildObjectURL(s.accountID, s.bucket, objectName)
}

func (s *objectStore) listURL(continuation string) (string, error) {
	if s.accountID == "" || s.bucket == "" {
		return "", fmt.Errorf("endpoint incomplete")
	}
	q := url.Values{}
	q.Set("list-type", "2")
	if continuation != "" {
		q.Set("continuation-token", continuation)
	}
	u := &url.URL{
		Scheme:   "https",
		Host:     s.accountID + ".r2.cloudflarestorage.com",
		Path:     "/" + s.bucket,
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
