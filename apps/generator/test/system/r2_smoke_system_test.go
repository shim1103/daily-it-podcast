//go:build r2smoke

// Scope: System（R2 疎通確認。workflow_dispatch 専用。cron なし）
// 実物境界: r2.EpisodeWriter.Write（Put）/ r2.ListObjectKeysSmoke（List）/ r2.GetObjectSmoke（Get）が
//
//	test bucket へ実 HTTPS で往復する。r2.DeleteObjectSmoke（Delete）は検証対象ではなく probe の後始末。
//
// Double: なし（TEST_R2_* の実 credential のみ。本番 credential は読まない）。
// 目的: 1 個の probe object が test bucket へ Put→List→Get 往復できるか、という単一の疎通仕様だけを見る。
//
//	stem pair 整合・schema 適合は見ない（Decision 2026-09-16T10-45-14 §1-2）。test/prod 切替引数は持たない（test 固定）。
//
// @require process env に TEST_R2_ACCOUNT_ID / TEST_R2_BUCKET / TEST_R2_ACCESS_KEY_ID /
//
//	TEST_R2_SECRET_ACCESS_KEY がある（無ければ Skip = 環境要因、smoke 対象外。本番 R2_* env 名は読まない）。
//
// @ensure probe object の Put→List→Get が単一往復として成功する。List に probe key が現れる。Get の byte が Put した内容と一致する。
// @ensure probe object は dispatch 末尾で必ず自前 delete する（成功・失敗いずれの経路でも t.Cleanup で保証）。
// @invariant 既定 -tags なしでは compile されない。secret 実値を t.Log / t.Fatalf に出さない。
package system

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

func TestR2Smoke_roundTripsProbeObject_overPutListGet(t *testing.T) {
	// Given: 実 TEST_R2_* credential（欠けたら Skip = 環境要因、smoke 対象外）
	accountID := strings.TrimSpace(os.Getenv("TEST_R2_ACCOUNT_ID"))
	bucket := strings.TrimSpace(os.Getenv("TEST_R2_BUCKET"))
	accessKeyID := strings.TrimSpace(os.Getenv("TEST_R2_ACCESS_KEY_ID"))
	secretAccessKey := strings.TrimSpace(os.Getenv("TEST_R2_SECRET_ACCESS_KEY"))
	if accountID == "" || bucket == "" || accessKeyID == "" || secretAccessKey == "" {
		t.Skip("smoke precondition: TEST_R2_ACCOUNT_ID / TEST_R2_BUCKET / TEST_R2_ACCESS_KEY_ID / TEST_R2_SECRET_ACCESS_KEY のいずれかが無い（R2 疎通 smoke を skip）")
	}

	// Given: dispatch run ごとに一意な probe stem（並行実行・前回残骸との衝突を避ける）と、
	//   Put→List→Get で使う固定 payload
	httpClient := &http.Client{Timeout: 30 * time.Second}
	probeStem := "r2-smoke-probe-" + time.Now().UTC().Format("20060102T150405.000000000Z07")
	probeJSONKey := probeStem + ".json"
	probeMP3Key := probeStem + ".mp3"
	wantJSON := []byte(`{"probe":true}`)
	wantMP3 := []byte("r2-smoke-probe-audio")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// @invariant probe object は成功・失敗いずれの経路でも dispatch 末尾で自前 delete する（検証対象外の後始末）。
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		for _, key := range []string{probeJSONKey, probeMP3Key} {
			if err := r2.DeleteObjectSmoke(cleanupCtx, httpClient, accessKeyID, secretAccessKey, accountID, bucket, key); err != nil {
				t.Logf("probe delete 失敗（%s）: %v", key, err)
			}
		}
	})

	// When: probe object を Put→List→Get の順で 1 回だけ往復させる
	writer := r2.NewEpisodeWriter(httpClient, accessKeyID, secretAccessKey, accountID, bucket)
	putErr := writer.Write(ctx, probeStem, wantJSON, models.SpeechAudio{Content: wantMP3})
	keys, listErr := r2.ListObjectKeysSmoke(ctx, httpClient, accessKeyID, secretAccessKey, accountID, bucket)
	gotJSON, getJSONErr := r2.GetObjectSmoke(ctx, httpClient, accessKeyID, secretAccessKey, accountID, bucket, probeJSONKey)
	gotMP3, getMP3Err := r2.GetObjectSmoke(ctx, httpClient, accessKeyID, secretAccessKey, accountID, bucket, probeMP3Key)

	// Then: 往復の各 property が成立する（同一 probe の複数 property 検証）
	if putErr != nil {
		t.Fatalf("Put（Write）疎通失敗: %v", putErr)
	}
	if listErr != nil {
		t.Fatalf("List 疎通失敗: %v", listErr)
	}
	if !containsKey(keys, probeJSONKey) {
		t.Fatalf("List に probe json key が見えない（probe key を隠して報告: 件数=%d）", len(keys))
	}
	if !containsKey(keys, probeMP3Key) {
		t.Fatalf("List に probe mp3 key が見えない（probe key を隠して報告: 件数=%d）", len(keys))
	}
	if getJSONErr != nil {
		t.Fatalf("Get（json）疎通失敗: %v", getJSONErr)
	}
	if getMP3Err != nil {
		t.Fatalf("Get（mp3）疎通失敗: %v", getMP3Err)
	}
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Fatal("Get（json）body が Put した内容と一致しない")
	}
	if !bytes.Equal(gotMP3, wantMP3) {
		t.Fatal("Get（mp3）body が Put した内容と一致しない")
	}

	t.Logf("R2 疎通 OK（Put→List→Get 往復成功。probe は Cleanup で delete 済み）")
}

func containsKey(keys []string, want string) bool {
	for _, k := range keys {
		if k == want {
			return true
		}
	}
	return false
}
