// Package r2locals3 は experimental local S3 peer の起動契約である。
// 本番 Adapter ではない。Adapter 振る舞い検証は httptest NI（r2_narrow_integration_test.go）が正。
//
// 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
//
// 既定 build（!r2locals3）の Start は Fake。実 wrangler 起動は `-tags r2locals3` の別 integration。
package r2locals3

import (
	"context"
	"strings"
)

// local S3 peer の注入値。本番 credential ではない。
const (
	localAccessKeyID     = "local-access-key-id"
	localSecretAccessKey = "local-secret-access-key"
	localAccountID       = "local"
	localBucket          = "generator-r2-local-s3"
	s3APIPathPrefix      = "/cdn-cgi/local/r2/s3"
	peerLogMaxRunes      = 400
)

// Peer は local S3 互換 endpoint の到達確認用注入値である。
//
// BaseURL は path-style S3 API の root（例: http://127.0.0.1:PORT/cdn-cgi/local/r2/s3）。
//
// @invariant AccessKey / Secret / AccountID / Bucket / BaseURL の実値を error message へ載せない。
type Peer struct {
	BaseURL         string
	AccessKeyID     string
	SecretAccessKey string
	AccountID       string
	Bucket          string
}

// Start は experimental local S3 peer を用意する。
//
// @require ctx は呼び出し元の寿命に従う。
// @ensure 既定 build は Fake Peer と no-op cleanup を返す（wrangler 未起動）。
// @ensure `-tags r2locals3` では実 endpoint の Peer と、必ず呼ぶ cleanup を返す。起動失敗は error（黙って skip しない）。
func Start(ctx context.Context) (peer *Peer, cleanup func(), err error) {
	return start(ctx)
}

// sanitizePeerLog は wrangler 起動 log から local peer 実値を伏せ、長さを制限する。
func sanitizePeerLog(raw string) string {
	out := raw
	for _, secret := range []string{localAccessKeyID, localSecretAccessKey, localBucket} {
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "[redacted]")
	}
	runes := []rune(out)
	if len(runes) > peerLogMaxRunes {
		out = string(runes[len(runes)-peerLogMaxRunes:])
	}
	return out
}
