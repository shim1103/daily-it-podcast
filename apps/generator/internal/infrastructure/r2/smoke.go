//go:build r2smoke

package r2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// 本 file は `workflow_dispatch` 疎通確認専用の最小 S3 互換往復を提供する。
// 既存 signV4* / buildObjectURL / retryLoop を再利用し、SigV4 署名の重複実装を避ける。
//
// @invariant EpisodeWriter / CompletedEpisodeLookup へ delete を追加しない
//   （「method は Write のみ」「List+Get のみ」という既存 Interface Segregation を保つ）。
//   本 file の関数は型に紐付かない独立の疎通専用 API であり、本番 Adapter の公開 API 集合を変更しない。
// @invariant `r2smoke` build tag 配下にのみ存在する。既定（タグ無し）の本番 build には一切含まれない
//   （`generator-produce-episode.yml` 等、タグを付けない通常 build から compile 対象外）。

// ListObjectKeysSmoke は疎通確認用に ListObjectsV2 を 1 回だけ呼び、object key 一覧を返す。
// ページング・pair 整合は見ない（疎通は probe 1 件のみを対象にする）。
//
// @require httpClient は非 nil。accessKeyID / secretAccessKey / accountID / bucket は呼び出し側で検証済み。
// @ensure network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @invariant error message に bucket / key / credential 実値を載せない。
func ListObjectKeysSmoke(ctx context.Context, httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string) ([]string, error) {
	if httpClient == nil {
		return nil, infraErr("list_smoke", fmt.Errorf("client is nil"))
	}
	target, err := smokeListURL(accountID, bucket)
	if err != nil {
		return nil, infraErr("build_url", err)
	}
	return retryLoop(maxPutAttempts, func() (bool, []string, error) {
		return listObjectKeysSmokeOnce(ctx, httpClient, target, accessKeyID, secretAccessKey)
	})
}

func listObjectKeysSmokeOnce(ctx context.Context, httpClient *http.Client, target, accessKeyID, secretAccessKey string) (retryable bool, keys []string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, nil, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, accessKeyID, secretAccessKey, time.Now()); err != nil {
		return false, nil, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return true, nil, infraErr("do", fmt.Errorf("request failed"))
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, nil, infraErr("read", fmt.Errorf("read body failed"))
	}
	if res.StatusCode == http.StatusOK {
		parsed, parseErr := parseListBucketResult(body)
		if parseErr != nil {
			return false, nil, infraErr("parse_list", fmt.Errorf("list parse failed"))
		}
		out := make([]string, 0, len(parsed.Contents))
		for _, c := range parsed.Contents {
			if c.Key != "" {
				out = append(out, c.Key)
			}
		}
		return false, out, nil
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return true, nil, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	return false, nil, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
}

// GetObjectSmoke は疎通確認用に GetObject を呼び、body を返す。
//
// @require httpClient は非 nil。accessKeyID / secretAccessKey / accountID / bucket / objectName は呼び出し側で検証済み。
// @ensure network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @invariant error message に bucket / key / credential 実値を載せない。
func GetObjectSmoke(ctx context.Context, httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket, objectName string) ([]byte, error) {
	if httpClient == nil {
		return nil, infraErr("get_smoke", fmt.Errorf("client is nil"))
	}
	target, err := buildObjectURL(accountID, bucket, objectName)
	if err != nil {
		return nil, infraErr("build_url", err)
	}
	return retryLoop(maxPutAttempts, func() (bool, []byte, error) {
		return getObjectSmokeOnce(ctx, httpClient, target, accessKeyID, secretAccessKey)
	})
}

func getObjectSmokeOnce(ctx context.Context, httpClient *http.Client, target, accessKeyID, secretAccessKey string) (retryable bool, body []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, nil, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, accessKeyID, secretAccessKey, time.Now()); err != nil {
		return false, nil, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return true, nil, infraErr("do", fmt.Errorf("request failed"))
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return false, nil, infraErr("read", fmt.Errorf("read body failed"))
	}
	if res.StatusCode == http.StatusOK {
		return false, raw, nil
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return true, nil, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	return false, nil, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
}

// DeleteObjectSmoke は疎通確認 probe object の自前削除に使う。
// 本番 Adapter（EpisodeWriter / CompletedEpisodeLookup）は delete を公開しない。本関数は疎通専用。
//
// @require httpClient は非 nil。accessKeyID / secretAccessKey / accountID / bucket / objectName は呼び出し側で検証済み。
// @ensure 2xx（204 等）で成功。network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @invariant error message に bucket / key / credential 実値を載せない。
func DeleteObjectSmoke(ctx context.Context, httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket, objectName string) error {
	if httpClient == nil {
		return infraErr("delete_smoke", fmt.Errorf("client is nil"))
	}
	target, err := buildObjectURL(accountID, bucket, objectName)
	if err != nil {
		return infraErr("build_url", err)
	}
	_, err = retryLoop(maxPutAttempts, func() (bool, struct{}, error) {
		retryable, opErr := deleteObjectSmokeOnce(ctx, httpClient, target, accessKeyID, secretAccessKey)
		return retryable, struct{}{}, opErr
	})
	return err
}

func deleteObjectSmokeOnce(ctx context.Context, httpClient *http.Client, target, accessKeyID, secretAccessKey string) (retryable bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, target, nil)
	if err != nil {
		return false, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	if err := signV4Get(req, accessKeyID, secretAccessKey, time.Now()); err != nil {
		return false, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return true, infraErr("do", fmt.Errorf("request failed"))
	}
	defer func() { _ = res.Body.Close() }()
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		return false, infraErr("read", fmt.Errorf("read body failed"))
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return false, nil
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return true, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	return false, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
}

func smokeListURL(accountID, bucket string) (string, error) {
	if accountID == "" || bucket == "" {
		return "", fmt.Errorf("endpoint incomplete")
	}
	q := url.Values{}
	q.Set("list-type", "2")
	u := &url.URL{
		Scheme:   "https",
		Host:     accountID + ".r2.cloudflarestorage.com",
		Path:     "/" + bucket,
		RawQuery: q.Encode(),
	}
	return u.String(), nil
}
