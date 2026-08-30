package build

import (
	"encoding/json"
	"fmt"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// speechSegmentsBeforeTopics は topic 群より前に来る固定 segment 数（Greeting + Intro）。
const speechSegmentsBeforeTopics = 2

// speechSegmentsPerTopic は 1 topic あたりの segment 数（Preface + Detail）。
const speechSegmentsPerTopic = 2

// SpeechTexts は TTS へ渡す朗読 text 列を発話順で返す。
//
// @require greeting は非空。farewell も非空（date 注入済みの文）。d は検証済み ManuscriptDraft。
// @ensure 順序は Greeting, Intro, 各 topic の Preface→Detail（topic 順）, ClosingSummary, ClosingFarewell。
func SpeechTexts(greeting, farewell string, d models.ManuscriptDraft) []string {
	texts := make([]string, 0, speechSegmentsBeforeTopics+speechSegmentsPerTopic*len(d.Topics)+2)
	texts = append(texts, greeting, d.Intro)
	for _, tp := range d.Topics {
		texts = append(texts, tp.Preface, tp.Detail)
	}
	texts = append(texts, d.ClosingSummary, farewell)
	return texts
}

// Timeline は segment 尺列から各 topic の Preface 開始秒と episode 全体尺を求める。
// 隣接 segment 間に constants.SegmentSilenceSec を加算する。
//
// @require len(segmentDurations) は SpeechTexts と同じ固定本数（speechSegmentsBeforeTopics + speechSegmentsPerTopic*topicCount + 2 = Greeting + Intro + Preface/Detail×topic + ClosingSummary + Farewell）。
// @ensure topicStartSecs[i] は i 番目 topic の Preface segment の開始累積秒（Detail は startSec に含めない）。
// @ensure totalDurationSec == Σ segmentDurations + SegmentSilenceSec*(len(segmentDurations)-1)。
// @ensure segment 本数が固定本数と不一致、または topicCount < 1 のとき Domain Error（Op = inconsistent_episode_assembly）。
func Timeline(segmentDurations []float64, topicCount int) (topicStartSecs []float64, totalDurationSec float64, err error) {
	if topicCount < 1 {
		return nil, 0, domainerrors.DomainErr(
			domainerrors.OpInconsistentEpisodeAssembly,
			fmt.Errorf("topicCount must be >= 1, got %d", topicCount),
		)
	}
	wantCount := speechSegmentsBeforeTopics + speechSegmentsPerTopic*topicCount + 2
	if n := len(segmentDurations); n != wantCount {
		return nil, 0, domainerrors.DomainErr(
			domainerrors.OpInconsistentEpisodeAssembly,
			fmt.Errorf("segment count %d is inconsistent with topicCount %d (want %d)", n, topicCount, wantCount),
		)
	}

	starts := make([]float64, topicCount)
	var cursor float64
	for i, dur := range segmentDurations {
		if i > 0 {
			cursor += constants.SegmentSilenceSec
		}
		// topic i の Preface は segment index (speechSegmentsBeforeTopics + i*speechSegmentsPerTopic)。
		if idx := i - speechSegmentsBeforeTopics; idx >= 0 && idx%speechSegmentsPerTopic == 0 {
			if topic := idx / speechSegmentsPerTopic; topic < topicCount {
				starts[topic] = cursor
			}
		}
		cursor += dur
	}
	return starts, cursor, nil
}

// ManuscriptInput は MarshalManuscript の入力を 1 つにまとめた Parameter Object。
type ManuscriptInput struct {
	EpisodeID      string
	Date           string
	Title          string
	DurationSec    float64
	Opening        string
	Draft          models.ManuscriptDraft
	TopicStartSecs []float64
	Closing        string
}

// MarshalManuscript は完成 manuscript.schema.json 形の JSON bytes を組む。
//
// @require in.TopicStartSecs と in.Draft.Topics は同数。
// @ensure 戻りは manuscript.schema.json の required（episodeId/date/title/durationSec/body）を満たす JSON bytes。Validate は行わない（Gate = WriteEpisode の責務）。
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
			Opening: in.Opening,
			Topics:  topics,
			Closing: in.Closing,
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
	Opening string                `json:"opening"`
	Topics  []manuscriptTopicJSON `json:"topics"`
	Closing string                `json:"closing"`
}

type manuscriptTopicJSON struct {
	Title    string  `json:"title"`
	Preface  string  `json:"preface"`
	Detail   string  `json:"detail"`
	StartSec float64 `json:"startSec"`
}
