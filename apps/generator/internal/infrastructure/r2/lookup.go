package r2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.CompletedEpisodeLookup = (*CompletedEpisodeLookup)(nil)

// CompletedEpisodeLookup は R2 上の完成ペア（同一 stem の json+mp3）を表示 date で照会する。
type CompletedEpisodeLookup struct {
	objectStore
}

// NewCompletedEpisodeLookup は R2 CompletedEpisodeLookup を返す。
//
// @require httpClient は非 nil（HasPair 時に検証）。accessKeyID / secretAccessKey / accountID / bucket は Composition が検証済み値を渡す。retry != nil（Composition Root の結線責務）。
// @ensure 戻りは非 nil の *CompletedEpisodeLookup（port.CompletedEpisodeLookup）。
// @invariant bucket・key・Account ID・Access Key・secret 実値を error / log へ載せない。
func NewCompletedEpisodeLookup(httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string, retry port.RetryReporter) *CompletedEpisodeLookup {
	return &CompletedEpisodeLookup{
		objectStore: objectStore{
			client:          httpClient,
			accessKeyID:     accessKeyID,
			secretAccessKey: secretAccessKey,
			accountID:       accountID,
			bucket:          bucket,
			now:             time.Now,
			retry:           retry,
		},
	}
}

// HasPair は所定空間に date 一致の完成ペアがあるとき true を返す。
//
// @require date は Port 契約どおり。本 Adapter は再検証しない。
// @ensure network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。
// @ensure 対象 stem への GET は最大 maxConcurrentGets 件まで同時実行してよい。
// @invariant error に bucket / key / credential 実値を載せない。
func (l *CompletedEpisodeLookup) HasPair(ctx context.Context, date string) (bool, error) {
	if l == nil || l.client == nil {
		return false, infraErr("has_pair", fmt.Errorf("client is nil"))
	}

	keys, err := l.listObjectKeys(ctx)
	if err != nil {
		return false, err
	}

	candidates := pairStemCandidates(keys)

	matches, err := l.matchDateOnStems(ctx, candidates, date)
	if err != nil {
		return false, err
	}
	for _, matched := range matches {
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// why: docs/decisions/2026-10-02T07-47-56 — List 済み key のメモリ分類。GET だけが fan-out 対象。
func pairStemCandidates(keys []string) []string {
	mp3Stems := make(map[string]struct{})
	var jsonStems []string
	for _, key := range keys {
		switch {
		case strings.HasSuffix(key, jsonExt):
			jsonStems = append(jsonStems, strings.TrimSuffix(key, jsonExt))
		case strings.HasSuffix(key, mp3Ext):
			mp3Stems[strings.TrimSuffix(key, mp3Ext)] = struct{}{}
		}
	}

	var candidates []string
	for _, stem := range jsonStems {
		if _, ok := mp3Stems[stem]; ok {
			candidates = append(candidates, stem)
		}
	}
	return candidates
}

func (l *CompletedEpisodeLookup) matchDateOnStems(ctx context.Context, candidates []string, date string) ([]bool, error) {
	matches := make([]bool, len(candidates))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentGets)
	for i, stem := range candidates {
		g.Go(func() error {
			raw, err := l.getObject(gctx, stem+jsonExt)
			if err != nil {
				return err
			}
			episodeDate, ok := manuscriptDate(raw)
			if !ok {
				return nil
			}
			if episodeDate == date {
				matches[i] = true
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return matches, nil
}

func manuscriptDate(raw []byte) (string, bool) {
	var doc struct {
		Date string `json:"date"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", false
	}
	date := strings.TrimSpace(doc.Date)
	if date == "" {
		return "", false
	}
	return date, true
}
