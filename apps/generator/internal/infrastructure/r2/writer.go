// Package r2 は Cloudflare R2（S3 互換）へ episode を書く Driven Adapter である。
package r2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.EpisodeWriter = (*EpisodeWriter)(nil)

// EpisodeWriter は R2 へ {episodeId}.json / {episodeId}.mp3 を put する。
// method は Write のみ（delete を公開しない）。
type EpisodeWriter struct {
	client          *http.Client
	accessKeyID     string
	secretAccessKey string
	accountID       string
	bucket          string
	now             func() time.Time
	retry           port.RetryReporter
}

// NewEpisodeWriter は R2 EpisodeWriter を返す。
//
// @require httpClient は非 nil（Write 時に検証）。accessKeyID / secretAccessKey / accountID / bucket は Composition が検証済み値を渡す。retry != nil（Composition Root の結線責務）。
// @ensure 戻りは非 nil の *EpisodeWriter（port.EpisodeWriter）。
// @invariant bucket・key・Account ID・Access Key・secret 実値を error / log へ載せない。
func NewEpisodeWriter(httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string, retry port.RetryReporter) *EpisodeWriter {
	return &EpisodeWriter{
		client:          httpClient,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		accountID:       accountID,
		bucket:          bucket,
		now:             time.Now,
		retry:           retry,
	}
}

// Write は原稿と音声を所定 bucket へ公開順（json → mp3）で put する。同 key は upsert。
//
// @require episodeID は非空。manuscript / audio.Content は非空（Port 契約。本 Adapter は再検証しない）。
// @ensure 成功時、所定空間直下に {episodeID}.json と {episodeID}.mp3 がある。途中失敗は成功にしない。
// @ensure network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @invariant storage 固有 id・MIME・vendor 型を露出せず、error に bucket / key / credential 実値を載せない。
func (w *EpisodeWriter) Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	if w == nil || w.client == nil {
		return infraErr("write", fmt.Errorf("client is nil"))
	}
	if err := w.putObject(ctx, episodeID+jsonExt, jsonMIME, manuscript); err != nil {
		return err
	}
	if err := w.putObject(ctx, episodeID+mp3Ext, mp3MIME, audio.Content); err != nil {
		return err
	}
	return nil
}

func (w *EpisodeWriter) putObject(ctx context.Context, objectName, mime string, content []byte) error {
	_, err := retryLoop(w.retry, "write_episode", maxPutAttempts, func() (bool, struct{}, error) {
		retryable, opErr := w.putOnce(ctx, objectName, mime, content)
		return retryable, struct{}{}, opErr
	})
	return err
}

func (w *EpisodeWriter) putOnce(ctx context.Context, objectName, mime string, content []byte) (retryable bool, err error) {
	target, err := w.objectURL(objectName)
	if err != nil {
		return false, infraErr("build_url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, bytes.NewReader(content))
	if err != nil {
		return false, infraErr("build_request", fmt.Errorf("new request failed"))
	}
	req.Header.Set("Content-Type", mime)
	if err := signV4Put(req, content, w.accessKeyID, w.secretAccessKey, w.now()); err != nil {
		return false, infraErr("sign", fmt.Errorf("sign failed"))
	}
	res, err := w.client.Do(req)
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

func (w *EpisodeWriter) objectURL(objectName string) (string, error) {
	return buildObjectURL(w.accountID, w.bucket, objectName)
}
