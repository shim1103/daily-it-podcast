//go:build r2locals3

// Scope: Narrow Integration
// 実物境界: r2.EpisodeWriter / r2.CompletedEpisodeLookup が標準 *http.Client で experimental local S3 へ送る PutObject / ListObjectsV2 / GetObject
// Double: 本番 R2 は使わない。peer は wrangler local S3（Decision 2026-09-15T12-02-48）。httptest 経路は別 file で併存。
// @require r2locals3.Start が実 peer を返す。credential は peer 注入値。
// @ensure Writer 成功系（json→mp3）と Lookup 成功系（Write 後 HasPair）が local S3 で緑。
// @invariant error message に bucket・key・Account ID・Access Key・secret 実値を含めない。
package test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
	"github.com/shim1103/daily-it-podcast/apps/generator/test/r2locals3"
)

var (
	sharedLocalS3Peer    *r2locals3.Peer
	sharedLocalS3Cleanup func()
	sharedLocalS3Err     error
	sharedLocalS3Once    sync.Once
)

func localS3Peer(t *testing.T) *r2locals3.Peer {
	t.Helper()
	sharedLocalS3Once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		peer, cleanup, err := r2locals3.Start(ctx)
		sharedLocalS3Peer = peer
		sharedLocalS3Err = err
		sharedLocalS3Cleanup = func() {
			cancel()
			if cleanup != nil {
				cleanup()
			}
		}
	})
	if sharedLocalS3Err != nil {
		t.Fatalf("r2locals3.Start: %v", sharedLocalS3Err)
	}
	return sharedLocalS3Peer
}

func TestMain(m *testing.M) {
	code := m.Run()
	if sharedLocalS3Cleanup != nil {
		sharedLocalS3Cleanup()
	}
	os.Exit(code)
}

func newWriterOnPeer(t *testing.T, peer *r2locals3.Peer) *r2.EpisodeWriter {
	t.Helper()
	return r2.NewEpisodeWriter(
		http.DefaultClient,
		peer.AccessKeyID,
		peer.SecretAccessKey,
		peer.AccountID,
		peer.Bucket,
	).WithEndpointBase(peer.BaseURL)
}

func newLookupOnPeer(t *testing.T, peer *r2locals3.Peer) *r2.CompletedEpisodeLookup {
	t.Helper()
	return r2.NewCompletedEpisodeLookup(
		http.DefaultClient,
		peer.AccessKeyID,
		peer.SecretAccessKey,
		peer.AccountID,
		peer.Bucket,
	).WithEndpointBase(peer.BaseURL)
}

func assertLocalS3NoSecretLeak(t *testing.T, peer *r2locals3.Peer, msg string) {
	t.Helper()
	leaks := []string{peer.AccessKeyID, peer.SecretAccessKey, peer.AccountID, peer.Bucket, peer.BaseURL}
	for _, leak := range leaks {
		if leak != "" && strings.Contains(msg, leak) {
			t.Fatalf("Error() が peer 実値を含む")
		}
	}
}

func TestR2EpisodeWriter_putsJSONThenMP3_whenLocalS3PeerSucceeds(t *testing.T) {
	// Given: experimental local S3 peer
	peer := localS3Peer(t)
	writer := newWriterOnPeer(t, peer)
	ms := []byte(`{"episodeId":"local-s3-ep-1","date":"2026-09-15"}`)
	audio := models.SpeechAudio{Content: []byte("local-s3-mp3-body")}

	// When: Write する
	if err := writer.Write(context.Background(), "local-s3-ep-1", ms, audio); err != nil {
		assertLocalS3NoSecretLeak(t, peer, err.Error())
		t.Fatalf("Write: %v", err)
	}

	// Then: Lookup が同 date の完成ペアを true と読む（同一 peer 上の観測）
	lookup := newLookupOnPeer(t, peer)
	got, err := lookup.HasPair(context.Background(), "2026-09-15")
	if err != nil {
		assertLocalS3NoSecretLeak(t, peer, err.Error())
		t.Fatalf("HasPair: %v", err)
	}
	if !got {
		t.Fatal("HasPair = false, want true")
	}
}

func TestR2CompletedEpisodeLookup_returnsFalse_whenLocalS3PeerHasNoPair(t *testing.T) {
	// Given: local S3 peer（Writer test と共有。空 date 照会）
	peer := localS3Peer(t)
	lookup := newLookupOnPeer(t, peer)

	// When: 存在しない date を照会する
	got, err := lookup.HasPair(context.Background(), "2099-01-01")

	// Then: false・error 無し
	if err != nil {
		assertLocalS3NoSecretLeak(t, peer, err.Error())
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}
