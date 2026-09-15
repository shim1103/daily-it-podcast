package r2

import (
	"context"
	"net/http"

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
}

// NewCompletedEpisodeLookup は R2 CompletedEpisodeLookup を返す。
//
// @require httpClient は非 nil（HasPair 時に検証予定）。accessKeyID / secretAccessKey / accountID / bucket は Composition が検証済み値を渡す。
// @ensure 戻りは非 nil の *CompletedEpisodeLookup（port.CompletedEpisodeLookup）。
// @invariant bucket・key・Account ID・Access Key・secret 実値を error / log へ載せない。
func NewCompletedEpisodeLookup(httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string) *CompletedEpisodeLookup {
	return &CompletedEpisodeLookup{
		client:          httpClient,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		accountID:       accountID,
		bucket:          bucket,
	}
}

// HasPair は所定空間に date 一致の完成ペアがあるとき true を返す。
//
// @require date は YYYY-MM-DD（Port 契約。本 Adapter は再検証しない）。
// @ensure 本 stub は常に false, nil（List/Get 未実装）。
// @invariant storage 固有 id・MIME・vendor 型を露出せず、error に bucket / key / credential 実値を載せない。
func (l *CompletedEpisodeLookup) HasPair(_ context.Context, _ string) (bool, error) {
	return false, nil
}
