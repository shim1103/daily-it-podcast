package build

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// ComposeBrief は Fetch 結果から TextWriter へ渡す brief 平文 1 本を組み立てる。
// 固定 Prompt は entities/constants.TextWriterBriefPrompt。本 func は埋め込みのみ。
//
// @require items は Fetch 成功後の slice。
// @ensure len(items) > 0 のとき戻りは trim 後に非空の brief 平文 1 本。
// @ensure len(items) == 0 のとき ("", Domain Error Op = no_source_items) を返す。
// @ensure constants.TextWriterBriefPrompt の {{SOURCES}} {{JSON_EXAMPLE}} と数値 placeholder を置換して完成させる。
// @ensure 数値 placeholder は manuscript_draft_limits 定数を embedManuscriptDraftLimits で埋める。{{SOURCES}} は各 item の SourceID・OccurredAt・Context を平文列挙（窓幅説明なし）。{{JSON_EXAMPLE}} は models.WriterOutput から生成。
// @ensure OpeningGreeting / ClosingFarewell は含めない。
// @invariant Prompt 散文を本 package に hardcode しない。Context を structured parse しない。
func ComposeBrief(items []models.SourceItem) (string, error) {
	return ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt)
}

// ComposeBriefWithTemplate は brief template 文字列を引数で受け取る ComposeBrief の一般版。
// rate 計測（system && ratemeasure）が prompt variant を差し替えて A/B するための注入口
// （Decision 2026-09-03T14-47-00）。
//
// @require items は Fetch 成功後の slice。template は brief prompt template（parse しない）。
// @ensure template の {{…_MIN}} 等の数値 placeholder を manuscript_draft_limits 定数で、
//
//	{{SOURCES}} を items の平文列挙で、{{JSON_EXAMPLE}} を models.WriterOutput 形式例で埋める。
//
// @ensure ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt) は ComposeBrief(items) と同一出力。
// @ensure len(items) == 0 のとき ("", Domain Error Op = no_source_items) を返す。
func ComposeBriefWithTemplate(items []models.SourceItem, template string) (string, error) {
	if len(items) == 0 {
		return "", domainerrors.DomainErr(domainerrors.OpNoSourceItems, nil)
	}

	brief := embedManuscriptDraftLimits(template)
	brief = strings.Replace(brief, "{{SOURCES}}", formatSourceItems(items), 1)
	jsonExample, err := marshalWriterOutputExample()
	if err != nil {
		return "", err
	}
	brief = strings.Replace(brief, "{{JSON_EXAMPLE}}", jsonExample, 1)
	return strings.TrimSpace(brief), nil
}

func formatSourceItems(items []models.SourceItem) string {
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
		b.WriteString("source_id: ")
		b.WriteString(item.SourceID)
		b.WriteString("\noccurred_at: ")
		b.WriteString(item.OccurredAt.UTC().Format(time.RFC3339))
		b.WriteByte('\n')
		b.WriteString(item.Context)
	}
	return b.String()
}

// marshalWriterOutputExample は {{JSON_EXAMPLE}} に埋める形式例を生成する。
// 実在の固有名詞は入れず、各 field は Prompt 各セクションの target 付近の長さのダミー文にする。
// これは形式例であり、topic 数と全体合計は Prompt の指定（目安 5 件・合計下限以上）に従うこと。
// example 内の 3 topic はあくまで shape を示すための最小構成で、各 field 単体は
// manuscript_draft_limits の rune 数 range 内に収めてある。
func marshalWriterOutputExample() (string, error) {
	example := models.WriterOutput{
		Title: "本日の主要トピックを一望できる形式例のタイトル文字列で長さの目安を示すものです",
		Intro: "本日は複数の技術ニュースをまとめて取り上げます。まず全体像を短く示し、そのうえでそれぞれの背景と要点を順番に整理していきます。今日の話題は運用基盤の更新から標準仕様の提案まで幅があり、共通する流れも見えてきます。専門用語は都度かみ砕いて説明するので、これから学ぶ方も前提から順に追ってください。それでは最初の話題から見ていきましょう。",
		Topics: []models.WriterOutputTopic{
			{
				Title:   "一つ目のトピックの題名の形式例",
				Preface: "一つ目の話題は運用基盤の大きな更新です。何がどう変わるのか、なぜ多くの開発チームが注目しているのかを、まずは全体像として短く押さえてから詳しい中身に入っていきます。前提となる用語や、これまでどんな課題があったのかも、ここで簡単に整理しておきます。",
				Detail:  "一つ目の話題の本文です。まず背景として、これまでこの分野では利用者が多くの設定を手作業で調整する必要があり、運用の負担が大きいという課題が長く指摘されてきました。設定を一つ誤るだけで性能が急に落ちることもあり、経験の浅いチームには扱いが難しい領域でした。監視のためのダッシュボードを整えるだけでも手間がかかり、専任の担当者を置けない現場では後回しになりがちでした。今回の更新では、過去の利用傾向をもとに主要なパラメータを自動で推定する仕組みが導入され、手作業の調整がほぼ不要になります。あわせて、急なアクセス増を見越して事前に処理能力を確保しておく機能も追加され、想定外の集中に対しても落ち着いて対応できるようになりました。この変化により、小規模なチームでも安定した性能を保ちやすくなり、これまで運用にかけていた時間を機能開発に回せるようになります。一方で、自動推定に任せきりにすると想定外の課金が発生する場合があるため、上限の設定と定期的な見直しは引き続き必要です。提供元は、今後さらに推定の精度を高め、対応する構成の種類も段階的に増やしていく方針だと説明しています。",
			},
			{
				Title:   "二つ目のトピックの題名の形式例",
				Preface: "二つ目の話題は、広く使われている開発ツールの新しい版についてです。全体の中での位置づけと、押さえておきたい論点を先に示してから、従来との違いや移行のしやすさを具体的に見ていきます。学び始めの方にも関わる変更なので、丁寧に追っていきます。",
				Detail:  "二つ目の話題の本文です。まず発表の要旨として、次のメジャー版の候補が公開され、非同期処理まわりの仕組みが標準の機能として取り込まれました。従来は各プロジェクトが独自に実装していた処理の受け渡しが、追加の部品を入れなくても書けるようになります。実際の開発では、複数の処理をまたいで情報を引き継ぐコードが短くなり、書き方のばらつきや設定ミスも減ると期待されます。とくに、いくつものチームが並行して開発する規模の大きな現場では、共通の書き方が決まっていることの価値が大きいと説明されています。ただし、古い前提で書かれた一部の拡張は手直しが必要になるため、移行の手順書と、機械的に書き換えを助ける道具が同時に用意されました。手順書には、変更の影響が出やすい箇所と、確認しておくべきテストの観点がまとめられています。メンテナは、正式版までに大きな仕様変更は入れない方針だと説明しており、今のうちに候補版で試して不具合を報告しておく価値があります。関連する周辺ツールも順次この版への対応を進めており、数か月のうちに主要なものはそろう見通しです。",
			},
			{
				Title:   "三つ目のトピックの題名の形式例",
				Preface: "三つ目の話題は、ブラウザに関わる標準仕様の新しい提案です。なぜこの話題を最後に置くのか、全体の流れの中での位置づけを短く述べてから、提案が何を解決しようとしているのか、その中身に入っていきます。",
				Detail:  "三つ目の話題の本文です。まず位置づけとして、これは既存の仕組みを置き換えるものではなく、機能を追加で足す提案であるという点が重要です。公表された内容によると、これまで実装ごとにばらつきがあった動作を、開発者が方針を明示的に指定できる形に整えます。指定できるのは、処理を実行する条件や、失敗したときにどう再試行するかといった考え方です。これまでは同じ書き方をしても環境によって挙動が変わることがあり、その差を吸収するための場当たりのコードが増えていました。新しい提案では、その差を仕様の側で埋めることを狙っています。これにより、利用者の環境が異なっても同じように動くことが期待でき、不具合が起きたときの切り分けもしやすくなります。提案はまだ意見募集の段階で、参照用の実装が一部のブラウザの開発版に入っている状況です。正式に決まるまでには時間がかかる見込みですが、主要な関係者の方向性はおおむね一致しているとされています。聞き手としては、対応状況の発表を折に触れて確認し、自分の使っている環境で試せるようになったら小さく動かしてみるとよいでしょう。",
			},
		},
		ClosingSummary: "本日取り上げた話題を振り返ります。運用基盤の更新は手作業の調整を減らし、開発ツールの新版は共通処理を標準化し、標準仕様の提案は動作のばらつきを抑える方向でした。いずれも、利用者の負担を下げて仕組みに任せる範囲を広げるという共通の流れがあります。細かな条件や対応時期はこれから変わることもあるため、詳細は各自で一次情報を確認してください。それではまた次回お会いしましょう。",
	}
	data, err := json.Marshal(example)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
