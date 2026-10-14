package r2

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// SameDayObjectCleaner は表示 date が一致する json（と sibling mp3）を削除する。
// EpisodeWriter は Write のみ。本型は system-test の fixture reset 専用で、Compose しない。
type SameDayObjectCleaner struct {
	objectStore
}

// NewSameDayObjectCleaner は同日 object 削除用の cleaner を返す。
//
// @require httpClient・retry は非 nil。credential は呼び出し側が検証済み。
// @require bucket 選択は呼び出し側の責任（本 cleaner は本番名を知らない・禁止しない）。
// @ensure 戻りは非 nil。
// @invariant bucket・key・credential 実値を error へ載せない。Compose 経路から呼ばない。
func NewSameDayObjectCleaner(httpClient *http.Client, accessKeyID, secretAccessKey, accountID, bucket string, retry port.RetryReporter) *SameDayObjectCleaner {
	return &SameDayObjectCleaner{
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

// RemoveForDate は date 一致 json と sibling mp3 を削除する。完成ペア要件は無い。
//
// @require date は表示暦日（YYYY-MM-DD）。呼び出し側が非 nil の本 cleaner を渡す（nil は programmer error）。
// @ensure 戻り removed は date 一致で削除した json 数。対象無しは 0。DELETE の 404 は成功。
// @ensure mp3-only（json 無し）は date が分からないため残す。別日の object は触らない。
func (c *SameDayObjectCleaner) RemoveForDate(ctx context.Context, date string) (removed int, err error) {
	keys, err := c.listObjectKeys(ctx)
	if err != nil {
		return 0, err
	}

	for _, key := range keys {
		if !strings.HasSuffix(key, jsonExt) {
			continue
		}
		raw, err := c.getObject(ctx, key)
		if err != nil {
			return removed, err
		}
		episodeDate, ok := manuscriptDate(raw)
		if !ok || episodeDate != date {
			continue
		}
		stem := strings.TrimSuffix(key, jsonExt)
		if err := c.deleteObject(ctx, key); err != nil {
			return removed, err
		}
		if err := c.deleteObject(ctx, stem+mp3Ext); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}
