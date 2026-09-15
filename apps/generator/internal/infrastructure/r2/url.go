package r2

import (
	"fmt"
	"net/url"
	"strings"
)

// buildObjectURL は R2 の path-style Object URL を組み立てる。
// why: host は Account ID 由来だが呼び出し側で error へ載せない。
func buildObjectURL(accountID, bucket, objectName string) (string, error) {
	if accountID == "" || bucket == "" {
		return "", fmt.Errorf("endpoint incomplete")
	}
	u := &url.URL{
		Scheme: "https",
		Host:   accountID + ".r2.cloudflarestorage.com",
		Path:   "/" + bucket + "/" + objectName,
	}
	return u.String(), nil
}

// buildObjectURLFromBase は local S3 等の endpoint base 配下へ path-style Object URL を組む。
// why: 本番は Account ID host、gate peer は wrangler `/cdn-cgi/local/r2/s3` であり URL 形が違う。
func buildObjectURLFromBase(endpointBase, bucket, objectName string) (string, error) {
	base, err := parseEndpointBase(endpointBase, bucket)
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + bucket + "/" + objectName
	return base.String(), nil
}

// buildListURLFromBase は endpoint base 配下の ListObjectsV2 URL を組む。
func buildListURLFromBase(endpointBase, bucket, continuation string) (string, error) {
	base, err := parseEndpointBase(endpointBase, bucket)
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + bucket
	q := url.Values{}
	q.Set("list-type", "2")
	if continuation != "" {
		q.Set("continuation-token", continuation)
	}
	base.RawQuery = q.Encode()
	return base.String(), nil
}

func parseEndpointBase(endpointBase, bucket string) (*url.URL, error) {
	if endpointBase == "" || bucket == "" {
		return nil, fmt.Errorf("endpoint incomplete")
	}
	base, err := url.Parse(endpointBase)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("endpoint incomplete")
	}
	return base, nil
}
