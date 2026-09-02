//go:build system

// Scope: System（CursorCLI 単体到達・尺安定性）
// 実物: cursorcli.TextWriter が実 CURSOR_API_KEY で実 Cursor CLI（agent）を叩く。
// Double: なし。GetX / Gemini / OAuth / Drive は import も呼び出しもしない。
// 目的: ニュース生成に十分な固定の擬似 SourceItem を build.ComposeBrief へ渡し、
//
//	tw.Write → build.ManuscriptDraftFromWriterOutput を N=3 回連続で通す。
//	毎回 valid な Draft が返ること（尺 range・件数を満たすこと）と各試行の所要秒を検証・記録する
//	（Decision 2026-09-02T18-26-00）。
//
// @require process env に CURSOR_API_KEY がある（無ければ Skip）。Cursor CLI の `agent` が PATH で
//
//	解決できる（無ければ Skip）。他 env（GEMINI / GetX / OAuth / Drive）は要らない。
//
// @ensure 3 回の Write がすべて非 error。3 回の draft parse がすべて非 error（1 回でも失敗したら test fail）。
//
//	各回の所要秒・topic 数・全体文字数を Logf で出す。
//
// @invariant local に secret を置かない。Gemini / Drive を import も呼び出しもしない。Domain 定数を変更しない。
package system

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

// cursorDraftAttempts は valid Draft を連続で要求する試行回数。
// flaky 検証なので 1 回では足りない（Decision 2026-09-02T18-26-00）。
const cursorDraftAttempts = 3

// seedSourceItems は Cursor 単体到達 test 用の固定擬似ソース。
// 実在企業名は避け「あるクラウド事業者」等に一般化する。各 Context は複数トピック分の
// 素材になる濃さ（各 200〜400 字程度の日本語）にして、目安件数の topic を書けるだけの入力にする。
func seedSourceItems() []models.SourceItem {
	base := time.Now().UTC()
	return []models.SourceItem{
		{
			SourceID:   "seed-tweet-1",
			OccurredAt: base.Add(-6 * time.Hour),
			Context: "あるクラウド事業者が、マネージド型のコンテナ実行基盤に大型アップデートを発表した。" +
				"新しい仕組みでは、これまで利用者が手動で設定していたオートスケールの閾値を、過去の負荷傾向から自動で推定する。" +
				"あわせて、コールドスタートの待ち時間を短縮するためのプリウォーム機能が追加され、突発的なアクセス増にも即応できるという。" +
				"料金体系は従来の実行時間課金に加えて、確保しておく最小インスタンス数に応じた月額枠が選べるようになった。" +
				"公式ブログでは、Web サービスの開発チームが運用負荷を下げつつ応答性能を保てる点を利点として挙げている。",
		},
		{
			SourceID:   "seed-tweet-2",
			OccurredAt: base.Add(-5 * time.Hour),
			Context: "広く使われているオープンソースの Web フレームワークが、メジャーバージョンの候補版を公開した。" +
				"目玉は非同期処理まわりの見直しで、リクエスト単位のコンテキスト受け渡しが標準 API に組み込まれた。" +
				"これにより、これまで各プロジェクトが独自実装していたトレース情報の伝播が、追加の依存なしで書けるようになる。" +
				"一方で、旧来の同期前提で書かれた一部の拡張は改修が必要になるため、移行ガイドと自動変換ツールが同時に用意された。" +
				"メンテナは、正式版までに大きな仕様変更は入れない方針だと説明している。",
		},
		{
			SourceID:   "seed-tweet-3",
			OccurredAt: base.Add(-4 * time.Hour),
			Context: "ある半導体メーカーが、エッジ機器向けの推論アクセラレータの新モデルを公表した。" +
				"従来品と同じ消費電力のまま、画像認識のスループットが目安で二倍近くに向上したとされる。" +
				"開発キットには量子化を前提としたモデル変換の一式が含まれ、主要な学習フレームワークからの書き出しに対応する。" +
				"実際の製品化では、監視カメラやロボット掃除機のような常時稼働する機器での採用が見込まれている。" +
				"供給時期は次の四半期からで、まず開発者向けの評価ボードが先行して出荷される。",
		},
		{
			SourceID:   "seed-tweet-4",
			OccurredAt: base.Add(-3 * time.Hour),
			Context: "あるコード管理サービスが、プルリクエストのレビュー支援機能を拡充した。" +
				"変更差分から影響範囲を推定し、関連するテストが不足している箇所を指摘する仕組みが試験的に導入された。" +
				"また、レビュー担当者の割り当てを、過去の変更履歴と現在の作業負荷から自動で提案する。" +
				"提案はあくまで参考で、チームの設定で無効化もできる。" +
				"提供元は、レビュー待ち時間の短縮と属人化の解消を狙いとして説明しており、まずは有料プランの一部利用者から段階的に開放する。",
		},
		{
			SourceID:   "seed-tweet-5",
			OccurredAt: base.Add(-2 * time.Hour),
			Context: "複数のブラウザベンダーが参加する標準化団体が、Web アプリのオフライン動作に関する新しい仕様案を公開した。" +
				"従来のキャッシュ制御を置き換えるものではなく、バックグラウンドでの再同期をより宣言的に書けるようにする追加 API という位置づけだ。" +
				"開発者は同期のタイミングや失敗時の再試行方針を設定として渡せるようになり、実装ごとのばらつきが減ると期待される。" +
				"仕様案は意見募集の段階で、参照実装が一部のブラウザの開発版に入っている。" +
				"正式勧告までには時間がかかる見通しだが、主要ベンダーの足並みはそろっているという。",
		},
	}
}

// draftTotalRunes は build.ManuscriptDraftFromWriterOutput の合計対象と同じ数え方で
// 全体文字数を数える（intro + closingSummary + Σ_topics(preface + detail)）。
// title / topic.title は朗読されない見出しなので数えない。
func draftTotalRunes(d models.ManuscriptDraft) int {
	total := utf8.RuneCountInString(strings.TrimSpace(d.Intro)) +
		utf8.RuneCountInString(strings.TrimSpace(d.ClosingSummary))
	for _, tp := range d.Topics {
		total += utf8.RuneCountInString(strings.TrimSpace(tp.Preface))
		total += utf8.RuneCountInString(strings.TrimSpace(tp.Detail))
	}
	return total
}

func TestCursorCLIDraftSystem_returnsValidDraftOnEveryAttempt_whenRealAPIKeyPresent(t *testing.T) {
	// Given: 実 CURSOR_API_KEY と PATH 上の `agent`（どちらか欠けたら Skip）
	apiKey := strings.TrimSpace(os.Getenv(config.CursorAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("System precondition: %s が無い（CursorCLI 単体到達 test を skip）", config.CursorAPIKeyEnv)
	}
	if _, err := exec.LookPath(cursorcli.BinaryName); err != nil {
		t.Skipf("System precondition: Cursor CLI %q が PATH で解決できない（CursorCLI 単体到達 test を skip）: %v", cursorcli.BinaryName, err)
	}

	// Given: 固定の擬似ソースから組んだ brief
	brief, err := build.ComposeBrief(seedSourceItems())
	if err != nil {
		t.Fatalf("ComposeBrief: %v", err)
	}

	// Given: 実 Cursor CLI 経由の TextWriter
	cursorFactory := processenv.NewSecretEnvLauncherFactory(os.Getenv(config.CursorAPIKeyEnv), os.LookupEnv)
	tw := cursorcli.NewTextWriter(cursorFactory)

	// ctx timeout: 15 分（3 回 × Cursor draft 数分）。
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// When / Then: N 回連続で Write → draft parse。1 回でも parse 失敗したら test は fail。
	for i := 1; i <= cursorDraftAttempts; i++ {
		attemptStart := time.Now()
		raw, err := tw.Write(ctx, brief)
		if err != nil {
			t.Fatalf("Cursor Write（%d/%d 回目）: %v", i, cursorDraftAttempts, err)
		}

		draft, err := build.ManuscriptDraftFromWriterOutput(raw)
		elapsed := time.Since(attemptStart)
		if err != nil {
			t.Errorf("draft parse（%d/%d 回目）が失敗: %v（所要 %.1fs）", i, cursorDraftAttempts, err, elapsed.Seconds())
			continue
		}
		t.Logf("draft OK（%d/%d 回目）: topic %d 件 / 全体 %d 文字 / 所要 %.1fs",
			i, cursorDraftAttempts, len(draft.Topics), draftTotalRunes(draft), elapsed.Seconds())
	}
}
