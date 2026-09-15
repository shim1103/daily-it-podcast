package r2

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
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
	return signV4(req, payload, accessKeyID, secretAccessKey, now)
}

// signV4Get は S3 互換 GetObject / ListObjectsV2 用に AWS Signature Version 4 を付与する。
//
// @require req は絶対 URL の GET。body 無し（payload は空）。
// @ensure Authorization / X-Amz-Date / X-Amz-Content-Sha256 を設定する。
// @invariant secret 実値を error message へ載せない。
func signV4Get(req *http.Request, accessKeyID, secretAccessKey string, now time.Time) error {
	return signV4(req, nil, accessKeyID, secretAccessKey, now)
}

func signV4(req *http.Request, payload []byte, accessKeyID, secretAccessKey string, now time.Time) error {
	if req == nil || req.URL == nil {
		return fmt.Errorf("request is nil")
	}
	if payload == nil {
		payload = []byte{}
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
	canonicalQuery := canonicalQueryString(req.URL.Query())

	signedHeadersList, canonicalHeaders := canonicalHeadersBlock(req.Header)
	canonicalRequest := strings.Join([]string{
		req.Method,
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

// canonicalQueryString は SigV4 用に query を名前順・URI encode で並べる。
// url.Values.Encode の + 空白は AWS 規約の %20 とずれるため使わない。
func canonicalQueryString(q url.Values) string {
	if len(q) == 0 {
		return ""
	}
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		vals := append([]string(nil), q[k]...)
		sort.Strings(vals)
		ek := uriEncode(k, true)
		for _, v := range vals {
			parts = append(parts, ek+"="+uriEncode(v, true))
		}
	}
	return strings.Join(parts, "&")
}

func uriEncode(s string, encodeSlash bool) string {
	var b strings.Builder
	b.Grow(len(s) * 3)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' || (!encodeSlash && c == '/') {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte("0123456789ABCDEF"[c>>4])
		b.WriteByte("0123456789ABCDEF"[c&15])
	}
	return b.String()
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
