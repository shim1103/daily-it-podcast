package application_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/fetch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/writeepisode"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// --- Test Double ---

// stubWriter は TextWriter の Stub。raw wire JSON（out）と error を制御し、呼ばれた回数を記録する。
// Run が buildFn に渡す build.ManuscriptDraftFromWriterOutput をそのまま呼び出し側から受け取り、
// out を渡して draft へ解釈させる（Run が retry を持たない新設計：invalid-draft retry は
// TextWriter 実装側の責務であり、この Stub は 1 回の解釈結果をそのまま返す）。
type stubWriter struct {
	out   string
	err   error
	calls int
}

func (s *stubWriter) Write(_ context.Context, _ string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	s.calls++
	if s.err != nil {
		return models.ManuscriptDraft{}, s.err
	}
	return buildFn(s.out)
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
	_ port.WAVToMP3Encoder   = (*stubEncoder)(nil)
)

// stubEncoder は WAVToMP3Encoder の Stub。返す MP3 と error を制御し、呼び出しを記録する。
type stubEncoder struct {
	out     []byte
	err     error
	calls   int
	lastWAV []byte
}

func (s *stubEncoder) EncodeWAVToMP3(_ context.Context, wav []byte) ([]byte, error) {
	s.calls++
	s.lastWAV = append([]byte(nil), wav...)
	if s.err != nil {
		return nil, s.err
	}
	if s.out != nil {
		return s.out, nil
	}
	return stubMP3Bytes, nil
}

// stubMP3Bytes は stubEncoder 既定の戻り MP3。WriteEpisode へ渡る音声の観測用。
var stubMP3Bytes = []byte("stub-mp3-bytes")

// fixedEpisodeID は newEpisodeID Stub が返す固定 ID。
const fixedEpisodeID = "ep-fixed-0001"

func fixedEpisodeIDFunc() string { return fixedEpisodeID }

// progressCall は spyProgress が記録した 1 回の呼び出し（Start または Done）。
type progressCall struct {
	method string // "start" または "done"
	step   string
	detail string
}

// spyProgress は port.ProgressReporter の Spy。Start/Done を呼び出し順に記録する。
type spyProgress struct {
	calls []progressCall
}

func (s *spyProgress) Start(step string) {
	s.calls = append(s.calls, progressCall{method: "start", step: step})
}

func (s *spyProgress) Done(step, detail string) {
	s.calls = append(s.calls, progressCall{method: "done", step: step, detail: detail})
}

func (s *spyProgress) stepNames() []string {
	names := make([]string, 0)
	for _, c := range s.calls {
		names = append(names, c.step)
	}
	return names
}

var _ port.ProgressReporter = (*spyProgress)(nil)

// harness は Run の SU test 用に全 double を結線した UseCase と各 Spy を保持する。
type harness struct {
	uc       *application.ProduceEpisode
	source   *fakeItemSource
	lookup   *fakeCompletedEpisodeLookup
	writer   *stubWriter
	synth    *spySynth
	encode   *stubEncoder
	episw    *fakeEpisodeWriter
	progress *spyProgress
}

// testDisplayLocation は表示タイムゾーンの test 用 Location。
// tzdata 非依存で環境に左右されないよう固定 offset(+9h) を使う。
// UTC 8/30 16:00 → JST 8/31 の跨ぎ検証も +9h で正しく成立する。
var testDisplayLocation = time.FixedZone("JST", 9*3600)

// newHarness は正常系 default（source 1 件、完成ペア無し、valid wire、尺 D 秒の固定 WAV）で harness を組む。
func newHarness(t *testing.T, segDurationSec float64) *harness {
	t.Helper()
	source := &fakeItemSource{items: []models.SourceItem{
		{SourceID: "x", OccurredAt: time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC), Summary: "item_id: a1"},
	}}
	lookup := &fakeCompletedEpisodeLookup{}
	writer := &stubWriter{out: buildValidWireJSON()}
	synth := &spySynth{wav: fixedWavOfDuration(t, segDurationSec)}
	encode := &stubEncoder{}
	episw := &fakeEpisodeWriter{}
	progress := &spyProgress{}
	uc := application.NewProduceEpisode(
		fetch.NewFetchSourceItems(source),
		lookup,
		writer,
		synth,
		encode,
		writeepisode.NewWriteEpisode(episw),
		fixedEpisodeIDFunc,
		testDisplayLocation,
		progress,
		constants.DraftTopicCountTarget,
	)
	return &harness{uc: uc, source: source, lookup: lookup, writer: writer, synth: synth, encode: encode, episw: episw, progress: progress}
}

// newHarnessWithTopicCount は newHarness の topicCount 可変版。wire は topicCount 件の
// topic を持つ valid JSON で組む（newHarness 既定の validWireTopicCount 固定とは独立に、
// Run が uc.topicCount を使って buildFn を呼ぶことを検証するための harness）。
func newHarnessWithTopicCount(t *testing.T, topicCount int) *harness {
	t.Helper()
	source := &fakeItemSource{items: []models.SourceItem{
		{SourceID: "x", OccurredAt: time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC), Summary: "item_id: a1"},
	}}
	lookup := &fakeCompletedEpisodeLookup{}
	writer := &stubWriter{out: buildValidWireJSONWithTopicCount(topicCount)}
	synth := &spySynth{wav: fixedWavOfDuration(t, 1.0)}
	encode := &stubEncoder{}
	episw := &fakeEpisodeWriter{}
	progress := &spyProgress{}
	uc := application.NewProduceEpisode(
		fetch.NewFetchSourceItems(source),
		lookup,
		writer,
		synth,
		encode,
		writeepisode.NewWriteEpisode(episw),
		fixedEpisodeIDFunc,
		testDisplayLocation,
		progress,
		topicCount,
	)
	return &harness{uc: uc, source: source, lookup: lookup, writer: writer, synth: synth, encode: encode, episw: episw, progress: progress}
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

	// Then: WriteEpisode が 1 回、episodeID は Stub 値、audio は encoder の MP3（生 WAV ではない）
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
	if h.encode.calls != 1 {
		t.Fatalf("EncodeWAVToMP3 calls = %d, want 1", h.encode.calls)
	}
	if !bytes.Equal(h.episw.audio.Content, stubMP3Bytes) {
		t.Fatalf("audio.Content = %q, want stub MP3 %q", h.episw.audio.Content, stubMP3Bytes)
	}
	if bytes.Equal(h.episw.audio.Content, h.encode.lastWAV) {
		t.Fatal("audio.Content が encode 前 WAV と同一（MP3 へ置換されていない）")
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
	// body.opening.text は greeting + 束境界 + intro（TTS が読む原稿そのものを契約へ入れる）。
	wantOpening := wantGreeting + bundleSep + wireIntroOf(t, h.writer.out)
	if m.Body.Opening.Text != wantOpening {
		t.Fatalf("body.opening.text = %q, want %q", m.Body.Opening.Text, wantOpening)
	}
	// body.opening.startSec は先頭 segment なので 0。
	if m.Body.Opening.StartSec != 0 {
		t.Fatalf("body.opening.startSec = %v, want 0", m.Body.Opening.StartSec)
	}
	if len(m.Body.Topics) != validWireTopicCount {
		t.Fatalf("body.topics count = %d, want %d", len(m.Body.Topics), validWireTopicCount)
	}
	// title は draft.Title（＝wire の title）がそのまま渡る
	if wantTitle := wireTitleOf(t, h.writer.out); m.Title != wantTitle {
		t.Fatalf("title = %q, want wire title %q", m.Title, wantTitle)
	}
	// body.ending.text は closingSummary + 束境界 + farewell（date 注入済み。TTS が読む原稿そのものを契約へ入れる）。
	wantEnding := wireClosingSummaryOf(t, h.writer.out) + bundleSep + wantFarewell
	if m.Body.Ending.Text != wantEnding {
		t.Fatalf("body.ending.text = %q, want %q", m.Body.Ending.Text, wantEnding)
	}
	if strings.Contains(m.Body.Ending.Text, "%s") {
		t.Fatalf("body.ending.text contains raw %%s: %q", m.Body.Ending.Text)
	}
	// body.ending.startSec は末尾束（closingSummary+farewell）の開始累積秒。topic の後なので > 0。
	if m.Body.Ending.StartSec <= 0 {
		t.Fatalf("body.ending.startSec = %v, want > 0", m.Body.Ending.StartSec)
	}
}

func TestProduceEpisodeRun_reportsStepProgressInOrder_whenAllStepsSucceed(t *testing.T) {
	t.Parallel()

	// Given: 全 step 成功
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if _, err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: 先頭は fetch_source_items、末尾は write_episode、already_produced は出ない。
	// 順序: concat_wav → encode_wav_to_mp3 → write_episode。
	got := h.progress.stepNames()
	if len(got) == 0 {
		t.Fatal("Progress が一度も呼ばれていない")
	}
	if got[0] != "fetch_source_items" {
		t.Fatalf("steps[0] = %q, want %q", got[0], "fetch_source_items")
	}
	if last := got[len(got)-1]; last != "write_episode" {
		t.Fatalf("steps[last] = %q, want %q", last, "write_episode")
	}
	concatIdx, encodeIdx, writeIdx := -1, -1, -1
	for i, s := range got {
		if s == "already_produced" {
			t.Fatalf("通常経路で already_produced が出た: %q", got)
		}
		if s == "concat_wav" && concatIdx < 0 {
			concatIdx = i
		}
		if s == "encode_wav_to_mp3" && encodeIdx < 0 {
			encodeIdx = i
		}
		if s == "write_episode" && writeIdx < 0 {
			writeIdx = i
		}
	}
	if concatIdx < 0 {
		t.Fatalf("concat_wav が無い: %q", got)
	}
	if encodeIdx < 0 {
		t.Fatalf("encode_wav_to_mp3 が無い: %q", got)
	}
	if writeIdx < 0 {
		t.Fatalf("write_episode が無い: %q", got)
	}
	if !(concatIdx < encodeIdx && encodeIdx < writeIdx) {
		t.Fatalf("順序が concat→encode→write ではない: concat=%d encode=%d write=%d steps=%q", concatIdx, encodeIdx, writeIdx, got)
	}
}

func TestProduceEpisodeRun_reportsFetchResultCountAsDetail_whenSourcesFetched(t *testing.T) {
	t.Parallel()

	// Given: 情報源 1 件（newHarness の既定）
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if _, err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: fetch_source_items の通知は完了後で detail に件数を含む（完全一致は brittle なので緩く見る）
	var fetchDetail string
	for _, c := range h.progress.calls {
		if c.step == "fetch_source_items" {
			fetchDetail = c.detail
		}
	}
	if !strings.Contains(fetchDetail, "件") {
		t.Fatalf("fetch_source_items detail = %q, want it to contain %q", fetchDetail, "件")
	}
}

func TestProduceEpisodeRun_reportsWriteEpisodeWithEpisodeIDAsDetail_whenWriteSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 全 step 成功
	h := newHarness(t, 1.0)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if _, err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: write_episode の detail は発行済み episodeID
	last := h.progress.calls[len(h.progress.calls)-1]
	if last.step != "write_episode" || last.detail != fixedEpisodeID {
		t.Fatalf("last progress = %+v, want {write_episode %s}", last, fixedEpisodeID)
	}
}

func TestProduceEpisodeRun_reportsOnlyAlreadyProduced_whenCompletedPairExists(t *testing.T) {
	t.Parallel()

	// Given: 表示 date に完成ペアあり
	h := newHarness(t, 1.0)
	h.lookup.has = true
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	if _, err := h.uc.Run(context.Background(), now); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then: already_produced だけが報告され、他の step は出ない
	got := h.progress.stepNames()
	if len(got) != 1 || got[0] != "already_produced" {
		t.Fatalf("steps = %q, want [already_produced]", got)
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
	// TTS 束の先頭・末尾は body.opening.text / body.ending.text と同一（読み上げ原稿を契約へ入れた結果）。
	if h.synth.texts[0] != m.Body.Opening.Text {
		t.Fatalf("texts[0] = %q, want body.opening.text %q", h.synth.texts[0], m.Body.Opening.Text)
	}
	// 中間束は topic ごとの preface + 改行3個 + detail。
	bundleSep := "\n\n\n"
	for i := 0; i < topicCount; i++ {
		want := m.Body.Topics[i].Preface + bundleSep + m.Body.Topics[i].Detail
		if got := h.synth.texts[1+i]; got != want {
			t.Fatalf("texts[%d] = %q, want topic[%d] bundle %q", 1+i, got, i, want)
		}
	}
	if last := h.synth.texts[len(h.synth.texts)-1]; last != m.Body.Ending.Text {
		t.Fatalf("last segment = %q, want body.ending.text %q", last, m.Body.Ending.Text)
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

// --- topicCount 反映 ---

func TestProduceEpisodeRun_writesEpisodeWithConfiguredTopicCount_whenWireTopicCountMatchesUseCaseTopicCount(t *testing.T) {
	t.Parallel()

	// Given: UseCase の topicCount = 3、wire も 3 topic
	const topicCount = 3
	h := newHarnessWithTopicCount(t, topicCount)
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), now)

	// Then: 成功し、manuscript の topic 数は UseCase の topicCount と一致する
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	m := unmarshalManuscript(t, h.episw.manuscript)
	if len(m.Body.Topics) != topicCount {
		t.Fatalf("body.topics count = %d, want %d", len(m.Body.Topics), topicCount)
	}
}

func TestProduceEpisodeRun_returnsInvalidManuscriptDraft_whenWireTopicCountDiffersFromUseCaseTopicCount(t *testing.T) {
	t.Parallel()

	// Given: UseCase の topicCount = 3 だが wire は固定 validWireTopicCount 件
	// （validWireTopicCount == constants.DraftTopicCountTarget であり 3 と異なる前提）
	const topicCount = 3
	if validWireTopicCount == topicCount {
		t.Fatalf("test 前提が崩れている: validWireTopicCount(%d) == topicCount(%d)", validWireTopicCount, topicCount)
	}
	h := newHarnessWithTopicCount(t, topicCount)
	h.writer.out = buildValidWireJSON() // validWireTopicCount 件の wire（topicCount と不一致）
	now := time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), now)

	// Then: buildFn（build.ManuscriptDraftFromWriterOutput(raw, uc.topicCount)）が
	// topic 数不一致を検出し、Op = invalid_manuscript_draft で失敗する
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpInvalidManuscriptDraft {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpInvalidManuscriptDraft)
	}
	if h.episw.calls != 0 {
		t.Fatalf("WriteEpisode calls = %d, want 0", h.episw.calls)
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

	// Given: TextWriter が壊れた JSON を返す（invalid-draft retry は TextWriter 実装側の責務であり
	// Run 自体は持たないため、stubWriter は 1 回 buildFn を呼ぶだけで invalid 判定を確定させる）
	h := newHarness(t, 1.0)
	h.writer.out = `{"title": "あ", "intro":`

	// When: Run を呼ぶ
	_, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: buildFn（build.ManuscriptDraftFromWriterOutput）の Op = invalid_manuscript_draft がそのまま伝播する。
	// TextWriter は 1 回だけ呼ばれ、Speech/WriteEpisode は呼ばれない
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpInvalidManuscriptDraft {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpInvalidManuscriptDraft)
	}
	if h.writer.calls != 1 {
		t.Fatalf("TextWriter calls = %d, want 1", h.writer.calls)
	}
	if h.synth.calls != 0 || h.episw.calls != 0 {
		t.Fatalf("downstream was called: synth=%d episw=%d", h.synth.calls, h.episw.calls)
	}
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

func TestProduceEpisodeRun_returnsErrorWithoutWriting_whenEncodeFails(t *testing.T) {
	t.Parallel()

	// Given: WAVToMP3Encoder が error
	boom := errors.New("encode boom")
	h := newHarness(t, 1.0)
	h.encode.err = boom

	// When: Run を呼ぶ
	gotID, err := h.uc.Run(context.Background(), time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC))

	// Then: その error を伝播。WriteEpisode は呼ばない。episodeID は空。
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if gotID != "" {
		t.Fatalf("episodeID = %q, want empty", gotID)
	}
	if h.encode.calls != 1 {
		t.Fatalf("EncodeWAVToMP3 calls = %d, want 1", h.encode.calls)
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
