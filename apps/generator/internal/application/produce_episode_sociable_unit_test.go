package application_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// --- Test Double ---

// stubWriter は TextWriter の Stub。返す string と error を制御し、呼ばれた回数を記録する。
// outs が非空なら呼び出し順に返し、尽きたら最後の要素を繰り返す。
type stubWriter struct {
	out   string
	outs  []string
	err   error
	calls int
}

func (s *stubWriter) Write(_ context.Context, _ string) (string, error) {
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	if n := len(s.outs); n > 0 {
		i := s.calls - 1
		if i >= n {
			i = n - 1
		}
		return s.outs[i], nil
	}
	return s.out, nil
}

// spySynth は SpeechSynthesizer の Spy。呼び出し text を順に記録し、指定 call 番号で error を返せる。
// 各成功呼び出しには既知尺の固定 WAV（wav）を返す。
type spySynth struct {
	texts      []string
	failAtCall int // 1-origin。0 なら失敗しない
	wav        []byte
}

func (s *spySynth) Synthesize(_ context.Context, text string) (models.SpeechAudio, error) {
	s.texts = append(s.texts, text)
	if s.failAtCall != 0 && len(s.texts) == s.failAtCall {
		return models.SpeechAudio{}, fmt.Errorf("synthesize failed at call %d", s.failAtCall)
	}
	return models.SpeechAudio{Content: s.wav}, nil
}

var (
	_ port.TextWriter        = (*stubWriter)(nil)
	_ port.SpeechSynthesizer = (*spySynth)(nil)
)

// fixedEpisodeID は newEpisodeID Stub が返す固定 ID。
const fixedEpisodeID = "ep-fixed-0001"

func fixedEpisodeIDFunc() string { return fixedEpisodeID }

// harness は Run の SU test 用に全 double を結線した UseCase と各 Spy を保持する。
type harness struct {
	uc     *application.ProduceEpisode
	source *fakeItemSource
	writer *stubWriter
	synth  *spySynth
	episw  *fakeEpisodeWriter
}

// testDisplayLocation は表示タイムゾーンの test 用 Location。
// tzdata 非依存で環境に左右されないよう固定 offset(+9h) を使う。
// UTC 8/30 16:00 → JST 8/31 の跨ぎ検証も +9h で正しく成立する。
var testDisplayLocation = time.FixedZone("JST", 9*3600)

// newHarness は正常系 default（source 1 件、valid wire、尺 D 秒の固定 WAV）で harness を組む。
func newHarness(t *testing.T, segDurationSec float64) *harness {
	t.Helper()
	source := &fakeItemSource{items: []models.SourceItem{
		{SourceID: "x", OccurredAt: time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC), Context: "item_id: a1"},
	}}
	writer := &stubWriter{out: buildValidWireJSON()}
	synth := &spySynth{wav: fixedWavOfDuration(t, segDurationSec)}
	episw := &fakeEpisodeWriter{}
	uc := application.NewProduceEpisode(
		application.NewFetchSourceItems(source),
		writer,
		synth,
		application.NewWriteEpisode(episw),
		fixedEpisodeIDFunc,
		testDisplayLocation,
	)
	return &harness{uc: uc, source: source, writer: writer, synth: synth, episw: episw}
}

// --- 正常系 ---

func TestProduceEpisodeRun_writesEpisodeWithAssembledManuscriptAndAudio_whenAllStepsSucceed(t *testing.T) {
	t.Parallel()

	// Given: 全 step 成功。now は UTC で JST へ跨ぐ時刻（UTC 8/30 16:00 → JST 8/31）
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), now)

	// Then: WriteEpisode が 1 回、episodeID は Stub 値、audio 非空
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if h.episw.calls != 1 {
		t.Fatalf("WriteEpisode calls = %d, want 1", h.episw.calls)
	}
	if h.episw.episodeID != fixedEpisodeID {
		t.Fatalf("episodeID = %q, want %q", h.episw.episodeID, fixedEpisodeID)
	}
	if len(h.episw.audio.Content) == 0 {
		t.Fatal("audio.Content is empty")
	}

	// Then: Run が組んだ manuscript bytes の形（入力伝播の検証。schema 適合検証は WriteEpisode の責務なのでしない）
	m := unmarshalManuscript(t, h.episw.manuscript)
	if m.EpisodeID != fixedEpisodeID {
		t.Fatalf("manuscript.episodeId = %q, want %q", m.EpisodeID, fixedEpisodeID)
	}
	if m.Date != "2026-08-31" {
		t.Fatalf("manuscript.date = %q, want JST calendar day 2026-08-31", m.Date)
	}
	wantGreeting := fmt.Sprintf(constants.OpeningGreetingTemplate, "2026年8月31日")
	if m.Body.Opening != wantGreeting {
		t.Fatalf("body.opening = %q, want %q", m.Body.Opening, wantGreeting)
	}
	if len(m.Body.Topics) != validWireTopicCount {
		t.Fatalf("body.topics count = %d, want %d", len(m.Body.Topics), validWireTopicCount)
	}
	// title は draft.Title（＝wire の title）がそのまま渡る
	if wantTitle := wireTitleOf(t, h.writer.out); m.Title != wantTitle {
		t.Fatalf("title = %q, want wire title %q", m.Title, wantTitle)
	}
	// body.closing は draft.ClosingSummary（wire の closingSummary）そのもの。
	// farewell（fmt.Sprintf(ClosingFarewell, spokenDate)）は音声のみで body.closing に含めない。
	wantClosing := wireClosingSummaryOf(t, h.writer.out)
	if m.Body.Closing != wantClosing {
		t.Fatalf("body.closing = %q, want wire closingSummary %q", m.Body.Closing, wantClosing)
	}
	if farewell := fmt.Sprintf(constants.ClosingFarewell, "2026年8月31日"); m.Body.Closing == farewell {
		t.Fatalf("body.closing must be draft.ClosingSummary, not the farewell line %q", farewell)
	}
	if strings.Contains(m.Body.Closing, "%s") {
		t.Fatalf("body.closing contains raw %%s: %q", m.Body.Closing)
	}
}

func TestProduceEpisodeRun_synthesizesSegmentsInGreetingIntroTopicsClosingOrder_whenDraftHasTopics(t *testing.T) {
	t.Parallel()

	// Given: 複数 topic の draft
	topicCount := validWireTopicCount
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: Synthesize の text 列は Greeting, Intro, (Preface, Detail)×topic, ClosingSummary, ClosingFarewell の順
	// farewell（date 注入済み）は常に末尾 1 segment。
	wantCount := 3 + topicCount*2 + 1 // Greeting + Intro + ClosingSummary + Farewell + Preface/Detail×topic
	if len(h.synth.texts) != wantCount {
		t.Fatalf("Synthesize calls = %d, want %d\ntexts=%q", len(h.synth.texts), wantCount, h.synth.texts)
	}

	m := unmarshalManuscript(t, h.episw.manuscript)
	if h.synth.texts[0] != m.Body.Opening {
		t.Fatalf("texts[0] = %q, want greeting %q", h.synth.texts[0], m.Body.Opening)
	}
	for i := 0; i < topicCount; i++ {
		preface := h.synth.texts[2+i*2]
		detail := h.synth.texts[2+i*2+1]
		if preface != m.Body.Topics[i].Preface {
			t.Fatalf("texts[%d] = %q, want topic[%d].preface %q", 2+i*2, preface, i, m.Body.Topics[i].Preface)
		}
		if detail != m.Body.Topics[i].Detail {
			t.Fatalf("texts[%d] = %q, want topic[%d].detail %q", 2+i*2+1, detail, i, m.Body.Topics[i].Detail)
		}
	}
	if h.synth.texts[2+topicCount*2] != m.Body.Closing {
		t.Fatalf("closingSummary segment = %q, want %q", h.synth.texts[2+topicCount*2], m.Body.Closing)
	}
	// 末尾は date 注入済みの farewell 文（生 template ではない）。
	wantFarewell := fmt.Sprintf(constants.ClosingFarewell, "2026年8月31日")
	if last := h.synth.texts[len(h.synth.texts)-1]; last != wantFarewell {
		t.Fatalf("last segment = %q, want farewell %q", last, wantFarewell)
	}
}

func TestProduceEpisodeRun_setsTopicStartSecFromCumulativeSegmentDurationsWithSilence_whenMultipleTopics(t *testing.T) {
	t.Parallel()

	// Given: 各 segment 尺 D 秒、無音 S 秒
	const d = 2.0
	s := constants.SegmentSilenceSec
	h := newHarness(t, d)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: topic[0].startSec = D(greeting)+S+D(intro)+S
	m := unmarshalManuscript(t, h.episw.manuscript)
	want0 := d + s + d + s
	if math.Abs(m.Body.Topics[0].StartSec-want0) > 1e-9 {
		t.Fatalf("topics[0].startSec = %v, want %v", m.Body.Topics[0].StartSec, want0)
	}
	// topic[1].startSec = topic[0].startSec + D(preface0)+S+D(detail0)+S
	want1 := want0 + d + s + d + s
	if math.Abs(m.Body.Topics[1].StartSec-want1) > 1e-9 {
		t.Fatalf("topics[1].startSec = %v, want %v", m.Body.Topics[1].StartSec, want1)
	}

	// Then: durationSec = 全 segment 尺合計 + S*(segment数-1)
	segCount := len(h.synth.texts)
	wantDuration := d*float64(segCount) + s*float64(segCount-1)
	if math.Abs(m.DurationSec-wantDuration) > 1e-9 {
		t.Fatalf("durationSec = %v, want %v (segCount=%d)", m.DurationSec, wantDuration, segCount)
	}
}

// --- 異常系 ---

func TestProduceEpisodeRun_returnsNoSourceItemsWithoutWriting_whenFetchReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Given: ItemSource が空 slice
	h := newHarness(t, 1.0)
	h.source.items = []models.SourceItem{}

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Now())

	// Then: Op = no_source_items の Domain Error。TextWriter/Speech/WriteEpisode いずれも呼ばれない
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpNoSourceItems {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpNoSourceItems)
	}
	if h.writer.calls != 0 {
		t.Fatalf("TextWriter calls = %d, want 0", h.writer.calls)
	}
	if len(h.synth.texts) != 0 {
		t.Fatalf("Synthesize calls = %d, want 0", len(h.synth.texts))
	}
	if h.episw.calls != 0 {
		t.Fatalf("WriteEpisode calls = %d, want 0", h.episw.calls)
	}
}

func TestProduceEpisodeRun_returnsErrorWithoutWriting_whenFetchFails(t *testing.T) {
	t.Parallel()

	// Given: ItemSource が error
	boom := errors.New("fetch boom")
	h := newHarness(t, 1.0)
	h.source.err = boom

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Now())

	// Then: その error を伝播。WriteEpisode は呼ばれない
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if h.writer.calls != 0 || len(h.synth.texts) != 0 || h.episw.calls != 0 {
		t.Fatalf("downstream was called: writer=%d synth=%d episw=%d", h.writer.calls, len(h.synth.texts), h.episw.calls)
	}
}

func TestProduceEpisodeRun_returnsErrorWithoutWriting_whenTextWriterFails(t *testing.T) {
	t.Parallel()

	// Given: TextWriter が error
	boom := errors.New("writer boom")
	h := newHarness(t, 1.0)
	h.writer.err = boom

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。Speech/WriteEpisode は呼ばれない
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if len(h.synth.texts) != 0 {
		t.Fatalf("Synthesize calls = %d, want 0", len(h.synth.texts))
	}
	if h.episw.calls != 0 {
		t.Fatalf("WriteEpisode calls = %d, want 0", h.episw.calls)
	}
}

func TestProduceEpisodeRun_returnsInvalidManuscriptDraftWithoutWriting_whenWriterOutputIsInvalid(t *testing.T) {
	t.Parallel()

	// Given: TextWriter が壊れた JSON を返す
	h := newHarness(t, 1.0)
	h.writer.out = `{"title": "あ", "intro":`

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: 上限まで再試行したうえで Op = invalid_manuscript_draft。Speech/WriteEpisode は呼ばれない
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpInvalidManuscriptDraft {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpInvalidManuscriptDraft)
	}
	if h.writer.calls != application.TextWriterMaxAttempts {
		t.Fatalf("TextWriter calls = %d, want %d", h.writer.calls, application.TextWriterMaxAttempts)
	}
	if len(h.synth.texts) != 0 || h.episw.calls != 0 {
		t.Fatalf("downstream was called: synth=%d episw=%d", len(h.synth.texts), h.episw.calls)
	}
}

func TestProduceEpisodeRun_retriesTextWriter_whenFirstDraftInvalidThenValid(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は壊れた wire、2 回目は valid wire
	h := newHarness(t, 1.0)
	seq := &seqWriter{outs: []string{`{"title": "あ", "intro":`, buildValidWireJSON()}}
	h.uc = application.NewProduceEpisode(
		application.NewFetchSourceItems(h.source),
		seq,
		h.synth,
		application.NewWriteEpisode(h.episw),
		fixedEpisodeIDFunc,
		testDisplayLocation,
	)

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: 2 回目で成功し WriteEpisode 1 回。2 回目 brief に前回 reject 理由が入る
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if seq.calls != 2 {
		t.Fatalf("TextWriter calls = %d, want 2", seq.calls)
	}
	if h.episw.calls != 1 {
		t.Fatalf("WriteEpisode calls = %d, want 1", h.episw.calls)
	}
	if len(seq.briefs) < 2 {
		t.Fatalf("briefs = %d, want >= 2", len(seq.briefs))
	}
	if !strings.Contains(seq.briefs[1], "Previous attempt rejected") {
		t.Fatalf("2 回目 brief に reject 理由が無い: %q", seq.briefs[1])
	}
	if !strings.Contains(seq.briefs[1], "invalid_manuscript_draft") && !strings.Contains(seq.briefs[1], "looking for beginning") {
		t.Fatalf("2 回目 brief に前回 error 本文が無い: %q", seq.briefs[1])
	}
}

// seqWriter は呼び出し順に out を返し、受け取った brief を記録する。
type seqWriter struct {
	outs   []string
	briefs []string
	calls  int
}

func (s *seqWriter) Write(_ context.Context, brief string) (string, error) {
	s.briefs = append(s.briefs, brief)
	i := s.calls
	s.calls++
	if i >= len(s.outs) {
		i = len(s.outs) - 1
	}
	return s.outs[i], nil
}

func TestProduceEpisodeRun_returnsErrorWithoutWriting_whenSynthesizeFails(t *testing.T) {
	t.Parallel()

	// Given: Speech が 2 回目の呼び出しで error
	h := newHarness(t, 1.0)
	h.synth.failAtCall = 2

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。WriteEpisode は呼ばれない。3 回目以降の Synthesize もない
	if err == nil {
		t.Fatal("Run: want error, got nil")
	}
	if len(h.synth.texts) != 2 {
		t.Fatalf("Synthesize calls = %d, want 2 (stops at first failure)", len(h.synth.texts))
	}
	if h.episw.calls != 0 {
		t.Fatalf("WriteEpisode calls = %d, want 0", h.episw.calls)
	}
}

func TestProduceEpisodeRun_returnsErrorWithoutWriting_whenEpisodeWriterFails(t *testing.T) {
	t.Parallel()

	// Given: EpisodeWriter が error
	boom := errors.New("episode writer boom")
	h := newHarness(t, 1.0)
	h.episw.err = boom

	// When: Run を呼ぶ
	err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。WriteEpisode は 1 回呼ばれている
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if h.episw.calls != 1 {
		t.Fatalf("WriteEpisode calls = %d, want 1", h.episw.calls)
	}
}
