package build

import (
	"encoding/json"
	"fmt"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// speechSegmentsBeforeTopics は topic 群より前に来る固定 segment 数（Greeting+Intro を 1 本に束ねる）。
const speechSegmentsBeforeTopics = 1

// speechSegmentsPerTopic は 1 topic あたりの segment 数（Preface+Detail を 1 本に束ねる）。
const speechSegmentsPerTopic = 1

// speechTextBundleDelimiter は greeting/intro・preface/detail・closing/farewell の境界。
// Cursor が topic.detail 内で使う改行（最大1個）と区別するため改行3個。
const speechTextBundleDelimiter = "\n\n\n"

// SpeechTexts は TTS へ渡す朗読 text 列を発話順で返す。
//
// TTS の request 回数を無料枠 quota 内へ抑えるため、朗読 text を topic+2 束へまとめる（Decision 2026-09-02T13-55-00）。
// 境界 delimiter は改行3個（Decision 2026-09-04T17-05-00）。
//
// @require greeting は非空。farewell も非空（date 注入済みの文）。d は検証済み ManuscriptDraft。
// @ensure 本数は 1 + len(d.Topics) + 1。順序は次のとおり:
//
//	texts[0]  = greeting + "\n\n\n" + d.Intro
//	texts[1..] = 各 topic の Preface + "\n\n\n" + Detail（topic 順）
//	texts[末尾] = d.ClosingSummary + "\n\n\n" + farewell
func SpeechTexts(greeting, farewell string, d models.ManuscriptDraft) []string {
	texts := make([]string, 0, speechSegmentsBeforeTopics+speechSegmentsPerTopic*len(d.Topics)+1)
	texts = append(texts, greeting+speechTextBundleDelimiter+d.Intro)
	for _, tp := range d.Topics {
		texts = append(texts, tp.Preface+speechTextBundleDelimiter+tp.Detail)
	}
	texts = append(texts, d.ClosingSummary+speechTextBundleDelimiter+farewell)
	return texts
}

// Timeline は segment 尺列から各 topic 束の開始秒と episode 全体尺を求める。
// 隣接 segment 間に constants.SegmentSilenceSec を加算する。
//
// SpeechTexts が topic+2 束を返すため、期待 segment 本数と topicStartSecs の意味も束ね後の単位に合わせる（Decision 2026-09-02T13-55-00）。
//
// @require len(segmentDurations) は SpeechTexts と同じ固定本数（1 + topicCount + 1 = greeting+intro 束 / 各 topic の preface+detail 束 / closingSummary+farewell 束）。
// @ensure topicStartSecs[i] は i 番目 topic 束（Preface+Detail 連結）の開始累積秒。
// @ensure endingStartSec は末尾 segment（closingSummary+farewell 束）の開始累積秒。
// @ensure totalDurationSec == Σ segmentDurations + SegmentSilenceSec*(len(segmentDurations)-1)。
// @ensure segment 本数が固定本数と不一致、または topicCount < 1 のとき Domain Error（Op = inconsistent_episode_assembly）。
func Timeline(segmentDurations []float64, topicCount int) (topicStartSecs []float64, endingStartSec float64, totalDurationSec float64, err error) {
	if topicCount < 1 {
		return nil, 0, 0, domainerrors.DomainErr(
			domainerrors.OpInconsistentEpisodeAssembly,
			fmt.Errorf("topicCount must be >= 1, got %d", topicCount),
		)
	}
	wantCount := speechSegmentsBeforeTopics + speechSegmentsPerTopic*topicCount + 1
	if n := len(segmentDurations); n != wantCount {
		return nil, 0, 0, domainerrors.DomainErr(
			domainerrors.OpInconsistentEpisodeAssembly,
			fmt.Errorf("segment count %d is inconsistent with topicCount %d (want %d)", n, topicCount, wantCount),
		)
	}

	starts := make([]float64, topicCount)
	endingIdx := len(segmentDurations) - 1
	var cursor float64
	for i, dur := range segmentDurations {
		if i > 0 {
			cursor += constants.SegmentSilenceSec
		}
		// topic i の束は segment index (speechSegmentsBeforeTopics + i*speechSegmentsPerTopic)。
		if idx := i - speechSegmentsBeforeTopics; idx >= 0 && idx%speechSegmentsPerTopic == 0 {
			if topic := idx / speechSegmentsPerTopic; topic < topicCount {
				starts[topic] = cursor
			}
		}
		if i == endingIdx {
			endingStartSec = cursor
		}
		cursor += dur
	}
	return starts, endingStartSec, cursor, nil
}

// ManuscriptInput は MarshalManuscript の入力を 1 つにまとめた Parameter Object。
type ManuscriptInput struct {
	EpisodeID      string
	Date           string
	Title          string
	DurationSec    float64
	Opening        string // 朗読全文: 定型挨拶 + intro（SpeechTexts[0] と同一）。body.opening.text へそのまま入る。
	Draft          models.ManuscriptDraft
	TopicStartSecs []float64
	Ending         string  // 朗読全文: closingSummary + 定型締め（SpeechTexts 末尾と同一）。body.ending.text へそのまま入る。
	EndingStartSec float64 // 末尾 segment（closingSummary+farewell 束）の音声上の開始秒。body.ending.startSec へ入る。
}

// MarshalManuscript は完成 manuscript.schema.json 形の JSON bytes を組む。
//
// @require in.TopicStartSecs と in.Draft.Topics は同数。in.Opening / in.Ending は朗読全文（定型込み）。
// @ensure 戻りは manuscript.schema.json の required（episodeId/date/title/durationSec/body）を満たす JSON bytes。Validate は行わない（Gate = WriteEpisode の責務）。
// @ensure body.opening.text / body.ending.text は入力をそのまま書く（TTS が読む原稿そのものを契約へ入れる。application 都合で定型を落とさない）。
// @ensure body.opening.startSec は先頭 segment なので 0 直書き。body.ending.startSec は in.EndingStartSec。
// @ensure len(in.TopicStartSecs) != len(in.Draft.Topics) のとき Domain Error（Op = inconsistent_episode_assembly）。
func MarshalManuscript(in ManuscriptInput) ([]byte, error) {
	if len(in.TopicStartSecs) != len(in.Draft.Topics) {
		return nil, domainerrors.DomainErr(
			domainerrors.OpInconsistentEpisodeAssembly,
			fmt.Errorf("topicStartSecs count %d != topics count %d", len(in.TopicStartSecs), len(in.Draft.Topics)),
		)
	}

	topics := make([]manuscriptTopicJSON, len(in.Draft.Topics))
	for i, tp := range in.Draft.Topics {
		topics[i] = manuscriptTopicJSON{
			Title:    tp.Title,
			Preface:  tp.Preface,
			Detail:   tp.Detail,
			StartSec: in.TopicStartSecs[i],
		}
	}

	doc := manuscriptJSON{
		EpisodeID:   in.EpisodeID,
		Date:        in.Date,
		Title:       in.Title,
		DurationSec: in.DurationSec,
		Body: manuscriptBodyJSON{
			Opening: manuscriptBookendJSON{
				Text: in.Opening,
				// opening は先頭 segment（greeting+intro 束）なので開始位置は定義上つねに 0。
				StartSec: 0,
			},
			Topics: topics,
			Ending: manuscriptBookendJSON{
				Text:     in.Ending,
				StartSec: in.EndingStartSec,
			},
		},
	}
	return json.Marshal(doc)
}

type manuscriptJSON struct {
	EpisodeID   string             `json:"episodeId"`
	Date        string             `json:"date"`
	Title       string             `json:"title"`
	DurationSec float64            `json:"durationSec"`
	Body        manuscriptBodyJSON `json:"body"`
}

type manuscriptBodyJSON struct {
	Opening manuscriptBookendJSON `json:"opening"`
	Topics  []manuscriptTopicJSON `json:"topics"`
	Ending  manuscriptBookendJSON `json:"ending"`
}

// manuscriptBookendJSON は body.opening / body.ending 共通の「朗読全文 + 音声上の開始秒」形。
type manuscriptBookendJSON struct {
	Text     string  `json:"text"`
	StartSec float64 `json:"startSec"`
}

type manuscriptTopicJSON struct {
	Title    string  `json:"title"`
	Preface  string  `json:"preface"`
	Detail   string  `json:"detail"`
	StartSec float64 `json:"startSec"`
}
