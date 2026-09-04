package build_test

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// assertInconsistentEpisodeAssembly は err が Op = inconsistent_episode_assembly の Domain Error で
// cause 非空であることを確認する。
func assertInconsistentEpisodeAssembly(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var de *domainerrors.Error
	if !errors.As(err, &de) {
		t.Fatalf("error type %T (%v), want *domainerrors.Error", err, err)
	}
	if de.Op != domainerrors.OpInconsistentEpisodeAssembly {
		t.Fatalf("Op = %q, want %q", de.Op, domainerrors.OpInconsistentEpisodeAssembly)
	}
	if de.Err == nil || strings.TrimSpace(de.Err.Error()) == "" {
		t.Fatal("cause が空")
	}
}

// draftFixture は helper test 共通の 2 topic 分の ManuscriptDraft を返す。
func draftFixture() models.ManuscriptDraft {
	return models.ManuscriptDraft{
		Title:          "きょうの IT ニュースまとめ",
		Intro:          "本日の導入です。",
		ClosingSummary: "本日のまとめです。",
		Topics: []models.ManuscriptDraftTopic{
			{Title: "話題いち", Preface: "前置きいち。", Detail: "詳細いち。"},
			{Title: "話題にい", Preface: "前置きにい。", Detail: "詳細にい。"},
		},
	}
}

// --- SpeechTexts ---

func TestSpeechTexts_returnsTopicPlusTwoBundles_inGreetingIntroTopicsClosingSummaryFarewellOrder(t *testing.T) {
	t.Parallel()

	// Given: greeting・farewell（date 注入済みの非空文）+ 2 topic の draft
	d := draftFixture()

	// When: TTS text 列を組む
	got := build.SpeechTexts("あいさつ文", "おわりの文。", d)

	// Then: 本数は 1 + topic 数 + 1。greeting+intro / topic ごと preface+detail / closingSummary+farewell を改行連結
	want := []string{
		"あいさつ文\n本日の導入です。",
		"前置きいち。\n詳細いち。",
		"前置きにい。\n詳細にい。",
		"本日のまとめです。\nおわりの文。",
	}
	assertStrings(t, got, want)
}

// --- Timeline ---

func TestTimeline_accumulatesSegmentDurationsWithSilence_andRecordsBundleStartPerTopic(t *testing.T) {
	t.Parallel()

	// Given: greeting+intro 束 + topic 束×2 + closingSummary+farewell 束 の 4 segment、各尺は既知、topic 数 2
	//        [greetingIntro, topic0, topic1, closingSummaryFarewell]
	durs := []float64{5, 9, 13, 17}
	s := constants.SegmentSilenceSec

	// When: timeline を組む
	starts, closingStart, total, err := build.Timeline(durs, 2)

	// Then: topic0 束の開始 = greetingIntro + S
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	want0 := 5 + s
	want1 := want0 + 9 + s
	if len(starts) != 2 {
		t.Fatalf("starts len = %d, want 2", len(starts))
	}
	if math.Abs(starts[0]-want0) > 1e-9 {
		t.Fatalf("starts[0] = %v, want %v", starts[0], want0)
	}
	if math.Abs(starts[1]-want1) > 1e-9 {
		t.Fatalf("starts[1] = %v, want %v", starts[1], want1)
	}

	// Then: closingSummary+farewell 束（末尾 segment）の開始 = topic1 束開始 + topic1 尺 + S
	wantClosing := want1 + 13 + s
	if math.Abs(closingStart-wantClosing) > 1e-9 {
		t.Fatalf("closingStart = %v, want %v", closingStart, wantClosing)
	}

	// Then: total = 全 segment 尺合計 + S*(segment数-1)
	wantTotal := (5.0 + 9 + 13 + 17) + s*3
	if math.Abs(total-wantTotal) > 1e-9 {
		t.Fatalf("total = %v, want %v", total, wantTotal)
	}
}

func TestTimeline_returnsSingleTopicStart_whenTopicCountIsOne(t *testing.T) {
	t.Parallel()

	// Given: 1 topic、3 segment [greetingIntro, topic0, closingSummaryFarewell]
	durs := []float64{1, 2, 3}
	s := constants.SegmentSilenceSec

	// When: timeline を組む
	starts, closingStart, total, err := build.Timeline(durs, 1)

	// Then: topic0 束の開始は greetingIntro + S
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	if len(starts) != 1 {
		t.Fatalf("starts len = %d, want 1", len(starts))
	}
	want0 := 1 + s
	if math.Abs(starts[0]-want0) > 1e-9 {
		t.Fatalf("starts[0] = %v, want %v", starts[0], want0)
	}
	// Then: closingSummary+farewell 束の開始 = topic0 束開始 + topic0 尺 + S
	wantClosing := want0 + 2 + s
	if math.Abs(closingStart-wantClosing) > 1e-9 {
		t.Fatalf("closingStart = %v, want %v", closingStart, wantClosing)
	}
	wantTotal := 6.0 + s*2
	if math.Abs(total-wantTotal) > 1e-9 {
		t.Fatalf("total = %v, want %v", total, wantTotal)
	}
}

func TestTimeline_returnsInconsistentEpisodeAssembly_whenSegmentCountMismatchesTopicCount(t *testing.T) {
	t.Parallel()

	// Given: topic 数 2 に対し segment 数が固定期待本数（1 + 2 + 1 = 4）と一致しない
	// When: timeline を組む
	starts, closingStart, total, err := build.Timeline([]float64{1, 2, 3}, 2)

	// Then: Op = inconsistent_episode_assembly の Domain Error。結果は返らない
	assertInconsistentEpisodeAssembly(t, err)
	if starts != nil || closingStart != 0 || total != 0 {
		t.Fatalf("starts/closingStart/total = %v / %v / %v, want nil / 0 / 0", starts, closingStart, total)
	}
}

func TestTimeline_returnsInconsistentEpisodeAssembly_whenTopicCountBelowOne(t *testing.T) {
	t.Parallel()

	// Given: topicCount = 0（下限割れ）
	// When: timeline を組む
	starts, closingStart, total, err := build.Timeline([]float64{1, 2}, 0)

	// Then: Op = inconsistent_episode_assembly の Domain Error
	assertInconsistentEpisodeAssembly(t, err)
	if starts != nil || closingStart != 0 || total != 0 {
		t.Fatalf("starts/closingStart/total = %v / %v / %v, want nil / 0 / 0", starts, closingStart, total)
	}
}

// --- MarshalManuscript ---

func TestMarshalManuscript_marshalsAllFields_whenInputsValid(t *testing.T) {
	t.Parallel()

	// Given: episodeID・date・title・durationSec・opening・draft・topicStartSecs・closing（summary と startSec）
	d := draftFixture()
	in := build.ManuscriptInput{
		EpisodeID:       "ep-fixed-0001",
		Date:            "2026-08-31",
		Title:           d.Title,
		DurationSec:     123,
		Opening:         "おはようございます。2026年8月31日です。",
		Draft:           d,
		TopicStartSecs:  []float64{10, 40},
		ClosingSummary:  d.ClosingSummary,
		ClosingStartSec: 100,
	}

	// When: JSON bytes を組む
	got, err := build.MarshalManuscript(in)

	// Then: Unmarshal して全 field が入る
	if err != nil {
		t.Fatalf("MarshalManuscript: %v", err)
	}
	var m struct {
		EpisodeID   string  `json:"episodeId"`
		Date        string  `json:"date"`
		Title       string  `json:"title"`
		DurationSec float64 `json:"durationSec"`
		Body        struct {
			Opening struct {
				Text     string  `json:"text"`
				StartSec float64 `json:"startSec"`
			} `json:"opening"`
			Topics []struct {
				Title    string  `json:"title"`
				Preface  string  `json:"preface"`
				Detail   string  `json:"detail"`
				StartSec float64 `json:"startSec"`
			} `json:"topics"`
			Closing struct {
				Summary  string  `json:"summary"`
				StartSec float64 `json:"startSec"`
			} `json:"closing"`
		} `json:"body"`
	}
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("Unmarshal: %v\n%s", err, got)
	}
	if m.EpisodeID != "ep-fixed-0001" || m.Date != "2026-08-31" || m.Title != d.Title {
		t.Fatalf("top-level fields = %+v", m)
	}
	if m.DurationSec != 123 {
		t.Fatalf("durationSec = %v, want 123", m.DurationSec)
	}
	if m.Body.Opening.Text != in.Opening || m.Body.Closing.Summary != d.ClosingSummary {
		t.Fatalf("body opening.text/closing.summary = %q / %q", m.Body.Opening.Text, m.Body.Closing.Summary)
	}
	if m.Body.Opening.StartSec != 0 {
		t.Fatalf("body opening.startSec = %v, want 0", m.Body.Opening.StartSec)
	}
	if m.Body.Closing.StartSec != 100 {
		t.Fatalf("body closing.startSec = %v, want 100", m.Body.Closing.StartSec)
	}
	if len(m.Body.Topics) != 2 {
		t.Fatalf("topics len = %d, want 2", len(m.Body.Topics))
	}
	if m.Body.Topics[0].Title != "話題いち" || m.Body.Topics[0].Preface != "前置きいち。" ||
		m.Body.Topics[0].Detail != "詳細いち。" || m.Body.Topics[0].StartSec != 10 {
		t.Fatalf("topics[0] = %+v", m.Body.Topics[0])
	}
	if m.Body.Topics[1].StartSec != 40 {
		t.Fatalf("topics[1].startSec = %v, want 40", m.Body.Topics[1].StartSec)
	}
}

func TestMarshalManuscript_returnsInconsistentEpisodeAssembly_whenTopicStartSecsCountDiffers(t *testing.T) {
	t.Parallel()

	// Given: topicStartSecs の数が draft.Topics と一致しない
	d := draftFixture()
	in := build.ManuscriptInput{
		EpisodeID:       "ep-1",
		Date:            "2026-08-31",
		Title:           d.Title,
		DurationSec:     1,
		Opening:         "x",
		Draft:           d,
		TopicStartSecs:  []float64{10},
		ClosingSummary:  d.ClosingSummary,
		ClosingStartSec: 100,
	}

	// When: JSON bytes を組む
	got, err := build.MarshalManuscript(in)

	// Then: Op = inconsistent_episode_assembly の Domain Error。bytes は返らない
	assertInconsistentEpisodeAssembly(t, err)
	if got != nil {
		t.Fatalf("got = %q, want nil", got)
	}
}

// assertStrings は 2 つの string slice が順序込みで一致することを確認する。
func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d\ngot:  %q\nwant: %q", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
