package r2

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func signingKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

// signV4Put は S3 互換 PutObject 用に AWS Signature Version 4 を付与する。
//
// @require req は絶対 URL の PUT。payload は body と一致する。
// @ensure Authorization / X-Amz-Date / X-Amz-Content-Sha256 を設定する。
// @invariant secret 実値を error message へ載せない。
func signV4Put(req *http.Request, payload []byte, accessKeyID, secretAccessKey string, now time.Time) error {
	if req == nil || req.URL == nil {
		return fmt.Errorf("request is nil")
	}
	payloadHash := sha256Hex(payload)
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")

	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	const canonicalQuery = ""

	signedHeadersList, canonicalHeaders := canonicalHeadersBlock(req.Header)
	canonicalRequest := strings.Join([]string{
		http.MethodPut,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeadersList,
		payloadHash,
	}, "\n")

	credentialScope := strings.Join([]string{dateStamp, awsRegion, awsService, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	sig := hex.EncodeToString(hmacSHA256(
		signingKey(secretAccessKey, dateStamp, awsRegion, awsService),
		stringToSign,
	))
	auth := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKeyID, credentialScope, signedHeadersList, sig,
	)
	req.Header.Set("Authorization", auth)
	return nil
}

func canonicalHeadersBlock(h http.Header) (signedHeaders, canonicalHeaders string) {
	type pair struct{ k, v string }
	var pairs []pair
	for k, vals := range h {
		lk := strings.ToLower(k)
		joined := strings.TrimSpace(strings.Join(vals, ","))
		joined = strings.Join(strings.Fields(joined), " ")
		pairs = append(pairs, pair{k: lk, v: joined})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	names := make([]string, 0, len(pairs))
	var b strings.Builder
	for _, p := range pairs {
		names = append(names, p.k)
		b.WriteString(p.k)
		b.WriteByte(':')
		b.WriteString(p.v)
		b.WriteByte('\n')
	}
	return strings.Join(names, ";"), b.String()
}
