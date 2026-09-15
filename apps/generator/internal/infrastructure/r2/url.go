package r2

import (
	"fmt"
	"net/url"
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
