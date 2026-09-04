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

// spySynth は SpeechSynthesizer の Spy。SynthesizeAll が受け取った texts 束と呼び出し回数を記録し、
// error を返すよう設定できる。各セグメントには既知尺の固定 WAV（wav）を返す。
type spySynth struct {
	calls      int
	texts      []string // 最後に SynthesizeAll へ渡された texts 束
	failAtCall int      // 0 なら成功。>0 なら SynthesizeAll が error を返す（WAV 列は返さない）
	wav        []byte
}

func (s *spySynth) SynthesizeAll(_ context.Context, texts []string) ([]models.SpeechAudio, error) {
	s.calls++
	s.texts = texts
	if s.failAtCall != 0 {
		return nil, fmt.Errorf("synthesize failed at segment %d", s.failAtCall)
	}
	audios := make([]models.SpeechAudio, len(texts))
	for i := range texts {
		audios[i] = models.SpeechAudio{Content: s.wav}
	}
	return audios, nil
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
	lookup *fakeCompletedEpisodeLookup
	writer *stubWriter
	synth  *spySynth
	episw  *fakeEpisodeWriter
}

// testDisplayLocation は表示タイムゾーンの test 用 Location。
// tzdata 非依存で環境に左右されないよう固定 offset(+9h) を使う。
// UTC 8/30 16:00 → JST 8/31 の跨ぎ検証も +9h で正しく成立する。
var testDisplayLocation = time.FixedZone("JST", 9*3600)

// newHarness は正常系 default（source 1 件、完成ペア無し、valid wire、尺 D 秒の固定 WAV）で harness を組む。
func newHarness(t *testing.T, segDurationSec float64) *harness {
	t.Helper()
	source := &fakeItemSource{items: []models.SourceItem{
		{SourceID: "x", OccurredAt: time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC), Context: "item_id: a1"},
	}}
	lookup := &fakeCompletedEpisodeLookup{}
	writer := &stubWriter{out: buildValidWireJSON()}
	synth := &spySynth{wav: fixedWavOfDuration(t, segDurationSec)}
	episw := &fakeEpisodeWriter{}
	uc := application.NewProduceEpisode(
		application.NewFetchSourceItems(source),
		lookup,
		writer,
		synth,
		application.NewWriteEpisode(episw),
		fixedEpisodeIDFunc,
		testDisplayLocation,
	)
	return &harness{uc: uc, source: source, lookup: lookup, writer: writer, synth: synth, episw: episw}
}

// --- 同日完成 skip ---

func TestProduceEpisodeRun_skipsWithoutFetch_whenCompletedPairExistsForDisplayDate(t *testing.T) {
	t.Parallel()

	// Given: 表示 date（JST）に完成ペアあり。now は UTC 8/30 16:00 → JST 8/31
	h := newHarness(t, 1.0)
	h.lookup.has = true
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), now)

	// Then: 成功。照会 date は JST 暦日。Fetch / TextWriter / Speech / WriteEpisode は呼ばない
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if h.lookup.calls != 1 {
		t.Fatalf("HasPair calls = %d, want 1", h.lookup.calls)
	}
	if h.lookup.lastDate != "2026-08-31" {
		t.Fatalf("HasPair date = %q, want 2026-08-31", h.lookup.lastDate)
	}
	if len(h.source.calls) != 0 {
		t.Fatalf("Fetch calls = %d, want 0", len(h.source.calls))
	}
	if h.writer.calls != 0 {
		t.Fatalf("TextWriter calls = %d, want 0", h.writer.calls)
	}
	if h.synth.calls != 0 {
		t.Fatalf("SynthesizeAll calls = %d, want 0", h.synth.calls)
	}
	if h.episw.calls != 0 {
		t.Fatalf("WriteEpisode calls = %d, want 0", h.episw.calls)
	}
}

func TestProduceEpisodeRun_continuesProduce_whenCompletedPairAbsent(t *testing.T) {
	t.Parallel()

	// Given: 完成ペア無し（Port が false。json only / wav only / 無しは Adapter が false に畳む）
	h := newHarness(t, 1.0)
	h.lookup.has = false
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), now)

	// Then: 通常 Produce 続行。HasPair は Fetch より前に 1 回。WriteEpisode 1 回
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if h.lookup.calls != 1 {
		t.Fatalf("HasPair calls = %d, want 1", h.lookup.calls)
	}
	if h.lookup.lastDate != "2026-08-31" {
		t.Fatalf("HasPair date = %q, want 2026-08-31", h.lookup.lastDate)
	}
	if len(h.source.calls) != 1 {
		t.Fatalf("Fetch calls = %d, want 1", len(h.source.calls))
	}
	if h.episw.calls != 1 {
		t.Fatalf("WriteEpisode calls = %d, want 1", h.episw.calls)
	}
}

func TestProduceEpisodeRun_returnsErrorWithoutFetch_whenCompletedEpisodeLookupFails(t *testing.T) {
	t.Parallel()

	// Given: 照会が error
	boom := errors.New("lookup boom")
	h := newHarness(t, 1.0)
	h.lookup.err = boom
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), now)

	// Then: その error を伝播。Fetch 以降は呼ばない
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if len(h.source.calls) != 0 {
		t.Fatalf("Fetch calls = %d, want 0", len(h.source.calls))
	}
	if h.writer.calls != 0 || h.synth.calls != 0 || h.episw.calls != 0 {
		t.Fatalf("downstream was called: writer=%d synth=%d episw=%d", h.writer.calls, h.synth.calls, h.episw.calls)
	}
}

// --- 正常系 ---

func TestProduceEpisodeRun_writesEpisodeWithAssembledManuscriptAndAudio_whenAllStepsSucceed(t *testing.T) {
	t.Parallel()

	// Given: 全 step 成功。now は UTC で JST へ跨ぐ時刻（UTC 8/30 16:00 → JST 8/31）
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	gotID, err := h.uc.Run(context.Background(), now)

	// Then: WriteEpisode が 1 回、episodeID は Stub 値、audio 非空
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotID != fixedEpisodeID {
		t.Fatalf("Run episodeID = %q, want %q", gotID, fixedEpisodeID)
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
	wantFarewell := fmt.Sprintf(constants.ClosingFarewell, "2026年8月31日")
	bundleSep := "\n\n\n"
	wantOpening := wantGreeting + bundleSep + wireIntroOf(t, h.writer.out)
	if m.Body.Opening != wantOpening {
		t.Fatalf("body.opening = %q, want %q", m.Body.Opening, wantOpening)
	}
	if len(m.Body.Topics) != validWireTopicCount {
		t.Fatalf("body.topics count = %d, want %d", len(m.Body.Topics), validWireTopicCount)
	}
	// title は draft.Title（＝wire の title）がそのまま渡る
	if wantTitle := wireTitleOf(t, h.writer.out); m.Title != wantTitle {
		t.Fatalf("title = %q, want wire title %q", m.Title, wantTitle)
	}
	// body.ending は closingSummary + farewell（TTS が読む原稿そのものを入れる）
	wantEnding := wireClosingSummaryOf(t, h.writer.out) + bundleSep + wantFarewell
	if m.Body.Ending != wantEnding {
		t.Fatalf("body.ending = %q, want %q", m.Body.Ending, wantEnding)
	}
	if strings.Contains(m.Body.Ending, "%s") {
		t.Fatalf("body.ending contains raw %%s: %q", m.Body.Ending)
	}
}

func TestProduceEpisodeRun_synthesizesTopicPlusTwoBundles_whenDraftHasTopics(t *testing.T) {
	t.Parallel()

	// Given: 複数 topic の draft
	topicCount := validWireTopicCount
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if _, err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: SynthesizeAll は 1 回だけ呼ばれ、渡された texts 束は 1 + topic 数 + 1 本。
	// texts[0] = greeting+intro、各 topic = preface+detail、末尾 = closingSummary+farewell（いずれも改行連結）。
	if h.synth.calls != 1 {
		t.Fatalf("SynthesizeAll calls = %d, want 1", h.synth.calls)
	}
	wantCount := 1 + topicCount + 1
	if len(h.synth.texts) != wantCount {
		t.Fatalf("SynthesizeAll texts 束 = %d, want %d\ntexts=%q", len(h.synth.texts), wantCount, h.synth.texts)
	}

	m := unmarshalManuscript(t, h.episw.manuscript)
	// TTS 束の先頭・末尾は body.opening / body.ending と同一（読み上げ原稿を契約へ入れた結果）。
	if h.synth.texts[0] != m.Body.Opening {
		t.Fatalf("texts[0] = %q, want body.opening %q", h.synth.texts[0], m.Body.Opening)
	}
	// 中間束は topic ごとの preface + 改行3個 + detail。
	bundleSep := "\n\n\n"
	for i := 0; i < topicCount; i++ {
		want := m.Body.Topics[i].Preface + bundleSep + m.Body.Topics[i].Detail
		if got := h.synth.texts[1+i]; got != want {
			t.Fatalf("texts[%d] = %q, want topic[%d] bundle %q", 1+i, got, i, want)
		}
	}
	if last := h.synth.texts[len(h.synth.texts)-1]; last != m.Body.Ending {
		t.Fatalf("last segment = %q, want body.ending %q", last, m.Body.Ending)
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
	if _, err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: topic[0] 束の startSec = D(greeting+intro 束)+S
	m := unmarshalManuscript(t, h.episw.manuscript)
	want0 := d + s
	if math.Abs(m.Body.Topics[0].StartSec-want0) > 1e-9 {
		t.Fatalf("topics[0].startSec = %v, want %v", m.Body.Topics[0].StartSec, want0)
	}
	// topic[1] 束の startSec = topic[0] 束の startSec + D(topic0 束)+S
	want1 := want0 + d + s
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
	_, err := h.uc.Run(context.Background(), time.Now())

	// Then: Op = no_source_items の Domain Error。TextWriter/Speech/WriteEpisode いずれも呼ばれない
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpNoSourceItems {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpNoSourceItems)
	}
	if h.writer.calls != 0 {
		t.Fatalf("TextWriter calls = %d, want 0", h.writer.calls)
	}
	if h.synth.calls != 0 {
		t.Fatalf("SynthesizeAll calls = %d, want 0", h.synth.calls)
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
	_, err := h.uc.Run(context.Background(), time.Now())

	// Then: その error を伝播。WriteEpisode は呼ばれない
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if h.writer.calls != 0 || h.synth.calls != 0 || h.episw.calls != 0 {
		t.Fatalf("downstream was called: writer=%d synth=%d episw=%d", h.writer.calls, h.synth.calls, h.episw.calls)
	}
}

func TestProduceEpisodeRun_returnsErrorWithoutWriting_whenTextWriterFails(t *testing.T) {
	t.Parallel()

	// Given: TextWriter が error
	boom := errors.New("writer boom")
	h := newHarness(t, 1.0)
	h.writer.err = boom

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。Speech/WriteEpisode は呼ばれない
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if h.synth.calls != 0 {
		t.Fatalf("SynthesizeAll calls = %d, want 0", h.synth.calls)
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
	_, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: 上限まで再試行したうえで Op = invalid_manuscript_draft。Speech/WriteEpisode は呼ばれない
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpInvalidManuscriptDraft {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpInvalidManuscriptDraft)
	}
	if h.writer.calls != application.TextWriterMaxAttempts {
		t.Fatalf("TextWriter calls = %d, want %d", h.writer.calls, application.TextWriterMaxAttempts)
	}
	if h.synth.calls != 0 || h.episw.calls != 0 {
		t.Fatalf("downstream was called: synth=%d episw=%d", h.synth.calls, h.episw.calls)
	}
}

func TestProduceEpisodeRun_retriesTextWriter_whenFirstDraftInvalidThenValid(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は壊れた wire、2 回目は valid wire
	h := newHarness(t, 1.0)
	seq := &seqWriter{outs: []string{`{"title": "あ", "intro":`, buildValidWireJSON()}}
	h.uc = application.NewProduceEpisode(
		application.NewFetchSourceItems(h.source),
		h.lookup,
		seq,
		h.synth,
		application.NewWriteEpisode(h.episw),
		fixedEpisodeIDFunc,
		testDisplayLocation,
	)

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

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

	// Given: SpeechSynthesizer.SynthesizeAll が error を返す（retry 予算切れなど、Adapter 内で確定した失敗）
	h := newHarness(t, 1.0)
	h.synth.failAtCall = 2

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。SynthesizeAll は 1 回だけ。WriteEpisode は呼ばれない
	if err == nil {
		t.Fatal("Run: want error, got nil")
	}
	if h.synth.calls != 1 {
		t.Fatalf("SynthesizeAll calls = %d, want 1", h.synth.calls)
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
	_, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。WriteEpisode は 1 回呼ばれている
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if h.episw.calls != 1 {
		t.Fatalf("WriteEpisode calls = %d, want 1", h.episw.calls)
	}
}
