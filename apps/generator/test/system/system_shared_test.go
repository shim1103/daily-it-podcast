//go:build system

// Scope: System 共通 helper（build tag system 全体で共有）
// Double: なし。固定の擬似 SourceItem と draft の rune 数え方だけを持つ。
// @invariant 実 I/O をしない。Domain 定数を変更しない。
package system

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// seedSourceItems は Cursor 単体到達 / draft rate 計測 test 用の固定擬似ソース。
// 実在企業名は避け「あるクラウド事業者」等に一般化する。各 Context は複数トピック分の
// 素材になる濃さ（各 200〜400 字程度の日本語）にして、目安件数の topic を書けるだけの入力にする。
func seedSourceItems() []models.SourceItem {
	base := time.Now().UTC()
	return []models.SourceItem{
		{
			SourceID:   "seed-item-1",
			OccurredAt: base.Add(-6 * time.Hour),
			Context: "あるクラウド事業者が、マネージド型のコンテナ実行基盤に大型アップデートを発表した。" +
				"新しい仕組みでは、これまで利用者が手動で設定していたオートスケールの閾値を、過去の負荷傾向から自動で推定する。" +
				"あわせて、コールドスタートの待ち時間を短縮するためのプリウォーム機能が追加され、突発的なアクセス増にも即応できるという。" +
				"料金体系は従来の実行時間課金に加えて、確保しておく最小インスタンス数に応じた月額枠が選べるようになった。" +
				"公式ブログでは、Web サービスの開発チームが運用負荷を下げつつ応答性能を保てる点を利点として挙げている。",
		},
		{
			SourceID:   "seed-item-2",
			OccurredAt: base.Add(-5 * time.Hour),
			Context: "広く使われているオープンソースの Web フレームワークが、メジャーバージョンの候補版を公開した。" +
				"目玉は非同期処理まわりの見直しで、リクエスト単位のコンテキスト受け渡しが標準 API に組み込まれた。" +
				"これにより、これまで各プロジェクトが独自実装していたトレース情報の伝播が、追加の依存なしで書けるようになる。" +
				"一方で、旧来の同期前提で書かれた一部の拡張は改修が必要になるため、移行ガイドと自動変換ツールが同時に用意された。" +
				"メンテナは、正式版までに大きな仕様変更は入れない方針だと説明している。",
		},
		{
			SourceID:   "seed-item-3",
			OccurredAt: base.Add(-4 * time.Hour),
			Context: "ある半導体メーカーが、エッジ機器向けの推論アクセラレータの新モデルを公表した。" +
				"従来品と同じ消費電力のまま、画像認識のスループットが目安で二倍近くに向上したとされる。" +
				"開発キットには量子化を前提としたモデル変換の一式が含まれ、主要な学習フレームワークからの書き出しに対応する。" +
				"実際の製品化では、監視カメラやロボット掃除機のような常時稼働する機器での採用が見込まれている。" +
				"供給時期は次の四半期からで、まず開発者向けの評価ボードが先行して出荷される。",
		},
		{
			SourceID:   "seed-item-4",
			OccurredAt: base.Add(-3 * time.Hour),
			Context: "あるコード管理サービスが、プルリクエストのレビュー支援機能を拡充した。" +
				"変更差分から影響範囲を推定し、関連するテストが不足している箇所を指摘する仕組みが試験的に導入された。" +
				"また、レビュー担当者の割り当てを、過去の変更履歴と現在の作業負荷から自動で提案する。" +
				"提案はあくまで参考で、チームの設定で無効化もできる。" +
				"提供元は、レビュー待ち時間の短縮と属人化の解消を狙いとして説明しており、まずは有料プランの一部利用者から段階的に開放する。",
		},
		{
			SourceID:   "seed-item-5",
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
