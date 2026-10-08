package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
)

// Scope: Sociable Unit
// 実物: gemini.SpeechSynthesizer（Gemini Interactions TTS Adapter）
// Double: http.RoundTripper の Spy（fakeRoundTripper）。待ちは backoffSleepFn / nowFn の差し替えで観測する。
// RetryReporter は retryReporterSpy。
//
// 並び: support → constructor → SynthesizeAll（合計予算・部分成功）→ synthesizeOne（retry loop・fallback へ渡す判定）→
// 待機（callGap・backoff）→ HTTP 取得・失敗分類・decode。

type retryReporterSpy struct {
	calls int
}

func (s *retryReporterSpy) Retry(step string, attempt, maxAttempts int, reason string) {
	s.calls++
}

// why: minSpeechDurationSec 未満だと decodeWAV が極小音声として一過性の失敗に落とすため、閾値を超える長さにする。
func minimalWAV() []byte {
	return synthHelperWAV(1, 24000, 16, 1.0)
}

func audioInteractionResponse(wav []byte) map[string]any {
	return map[string]any{
		"status": "completed",
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(wav)},
			}},
		},
	}
}

func isWAV(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}

type fakeClientCall struct {
	Method string
	URL    string
	Body   []byte
}

type fakeRoundTripper struct {
	responses []fakeClientResponse
	calls     []fakeClientCall
}

type fakeClientResponse struct {
	status int
	header http.Header
	body   []byte
	err    error
}

func (rt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	rt.calls = append(rt.calls, fakeClientCall{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   body,
	})
	index := len(rt.calls) - 1
	if index >= len(rt.responses) {
		return nil, fmt.Errorf("fakeRoundTripper: no response configured for call %d", index)
	}
	resp := rt.responses[index]
	if resp.err != nil {
		return nil, resp.err
	}
	rec := httptest.NewRecorder()
	for key, values := range resp.header {
		for _, v := range values {
			rec.Header().Add(key, v)
		}
	}
	rec.WriteHeader(resp.status)
	if resp.body != nil {
		_, _ = rec.Write(resp.body)
	}
	return rec.Result(), nil
}

func jsonBody(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return raw
}

func newFakeSynthesizer(responses ...fakeClientResponse) (*SpeechSynthesizer, *fakeRoundTripper) {
	rt := &fakeRoundTripper{responses: responses}
	synth := newSpeechSynthesizer(&http.Client{Transport: rt}, "gemini-fake-key", func(time.Duration) {}, &retryReporterSpy{})
	return synth, rt
}

// why: 合計予算・入口ガードは SynthesizeAll の test が持つので、ここは 1 セグメントの retry loop だけを MaxAttempts 上限で叩く。
func (s *SpeechSynthesizer) synthTestOne(ctx context.Context, text string) (models.SpeechAudio, error) {
	audio, _, err := s.synthesizeOne(ctx, text, MaxAttempts)
	return audio, err
}

// TestNewSpeechSynthesizerWithTuning_fallsBackToDefaultTuning_whenZeroValueFields は
// Tuning のゼロ値 field が既定値へフォールバックすることを検証する。
func TestNewSpeechSynthesizerWithTuning_fallsBackToDefaultTuning_whenZeroValueFields(t *testing.T) {
	// Given: 空 Tuning を注入した Synthesizer
	synth := NewSpeechSynthesizerWithTuning(&http.Client{}, "gemini-fake-key", TierFree, Tuning{}, nil)

	// Then: 各 tuning field が既定値
	if synth.callGap != defaultCallGap {
		t.Fatalf("callGap = %v, want %v", synth.callGap, defaultCallGap)
	}
	if synth.retryBackoffBase != defaultRetryBackoffBase {
		t.Fatalf("retryBackoffBase = %v, want %v", synth.retryBackoffBase, defaultRetryBackoffBase)
	}
	if synth.retryBackoffMax != defaultRetryBackoffMax {
		t.Fatalf("retryBackoffMax = %v, want %v", synth.retryBackoffMax, defaultRetryBackoffMax)
	}
}

// TestNewSpeechSynthesizer_usesDefaultTuning_whenConstructedPlainly は
// 既定 constructor が既定 tuning を使う（挙動不変）ことを検証する。
func TestNewSpeechSynthesizer_usesDefaultTuning_whenConstructedPlainly(t *testing.T) {
	// Given / When: 既定 constructor
	synth := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierFree, nil)

	// Then: tuning field はすべて既定値
	if synth.callGap != defaultCallGap || synth.retryBackoffBase != defaultRetryBackoffBase || synth.retryBackoffMax != defaultRetryBackoffMax {
		t.Fatalf("tuning = {%v, %v, %v}, want defaults {%v, %v, %v}",
			synth.callGap, synth.retryBackoffBase, synth.retryBackoffMax,
			defaultCallGap, defaultRetryBackoffBase, defaultRetryBackoffMax)
	}
}

// TestNewSpeechSynthesizer_storesGivenTier_whenConstructed は
// 渡した Tier がそのまま保持されることを検証する（config 変数名からの暗黙判定をやめ
// 明示引数にした Decision 2026-09-16T11-41-59 §1-5 の契約）。
func TestNewSpeechSynthesizer_storesGivenTier_whenConstructed(t *testing.T) {
	// Given / When: TierPaid を渡して構築する
	synth := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierPaid, nil)

	// Then: 渡した Tier がそのまま保持される
	if synth.tier != TierPaid {
		t.Fatalf("tier = %v, want %v", synth.tier, TierPaid)
	}
}

func TestSynthesizeAll_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: nil client
	synth := NewSpeechSynthesizer(nil, "gemini-edge-key", TierFree, nil)

	// When: SynthesizeAll する
	_, err := synth.SynthesizeAll(context.Background(), []string{"本文"})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSynthesizeAll_returnsInfrastructureError_whenReceiverNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var synth *SpeechSynthesizer

	// When: SynthesizeAll する
	_, err := synth.SynthesizeAll(context.Background(), []string{"本文"})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, infra)
	}
}

// TestSynthesizeAll_succeedsForAllTexts_whenEveryCallReturnsAudio は
// texts を順に Synthesize し、WAV 列（結合しない）を texts と同数返すことを検証する。
func TestSynthesizeAll_succeedsForAllTexts_whenEveryCallReturnsAudio(t *testing.T) {
	// Given: 3 本の text、各 1 回で成功
	texts := []string{"一本目", "二本目", "三本目"}
	responses := make([]fakeClientResponse, len(texts))
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body:   jsonBody(t, audioInteractionResponse(minimalWAV())),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: SynthesizeAll する
	got, err := synth.SynthesizeAll(context.Background(), texts)

	// Then: WAV 列を texts と同数返す。結合しない
	if err != nil {
		t.Fatalf("SynthesizeAll: %v", err)
	}
	if len(got) != len(texts) {
		t.Fatalf("audios = %d, want %d", len(got), len(texts))
	}
	for i, a := range got {
		if !isWAV(a.Content) {
			t.Fatalf("audios[%d] is not WAV: %d bytes", i, len(a.Content))
		}
	}
	if len(rt.calls) != len(texts) {
		t.Fatalf("call count = %d, want %d", len(rt.calls), len(texts))
	}
}

func TestSynthesizeAll_fillsDurationSecOfEachContent_whenEveryCallReturnsAudio(t *testing.T) {
	// Given: 尺の異なる既知長 WAV を返す 2 回の呼び出し
	wantDurations := []float64{1.5, 2.5}
	responses := make([]fakeClientResponse, len(wantDurations))
	for i, d := range wantDurations {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body:   jsonBody(t, audioInteractionResponse(synthHelperWAV(1, 24000, 16, d))),
		}
	}
	synth, _ := newFakeSynthesizer(responses...)

	// When: SynthesizeAll する
	got, err := synth.SynthesizeAll(context.Background(), []string{"一本目", "二本目"})

	// Then: 各要素の DurationSec が WAV から求めた秒数と一致する（0 のまま残らない）
	if err != nil {
		t.Fatalf("SynthesizeAll: %v", err)
	}
	if len(got) != len(wantDurations) {
		t.Fatalf("audios = %d, want %d", len(got), len(wantDurations))
	}
	for i, a := range got {
		if math.Abs(a.DurationSec-wantDurations[i]) > 1e-4 {
			t.Fatalf("audios[%d].DurationSec = %v, want %v", i, a.DurationSec, wantDurations[i])
		}
	}
}

// TestSynthesizeAll_returnsError_whenTextEmpty は空要素で Client を呼ばず error を返すことを検証する。
func TestSynthesizeAll_returnsError_whenTextEmpty(t *testing.T) {
	synth, rt := newFakeSynthesizer()

	// When: 空要素を含む texts
	_, err := synth.SynthesizeAll(context.Background(), []string{"  \t "})

	// Then: Infrastructure Error。Client は呼ばない
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 0 {
		t.Fatalf("unexpected calls: %#v", rt.calls)
	}
}

// TestSynthesizeAll_consumesBudgetAcrossSegments_thenErrorsWhenExhausted は
// 各セグメントが「retry してから成功」で呼び出しを積み上げ、合計が SynthesizeBudget へ達したら
// 以降のセグメントは Client を呼ばず即 error を返すことを検証する（残予算はセグメント横断）。
func TestSynthesizeAll_consumesBudgetAcrossSegments_thenErrorsWhenExhausted(t *testing.T) {
	// Given: 各セグメントは「503 → 成功」の 2 呼び出しで成功する。
	//        2 呼び出し × 5 セグメント = 10 = SynthesizeBudget。6 本目で予算切れ。
	const callsPerSegment = 2
	const fullSegments = SynthesizeBudget / callsPerSegment // 5
	if fullSegments*callsPerSegment != SynthesizeBudget {
		t.Fatalf("SynthesizeBudget = %d が callsPerSegment = %d で割り切れない。この test の前提が崩れている", SynthesizeBudget, callsPerSegment)
	}
	texts := make([]string, fullSegments+1)
	for i := range texts {
		texts[i] = fmt.Sprintf("セグメント%d", i)
	}

	var responses []fakeClientResponse
	for seg := 0; seg < fullSegments; seg++ {
		responses = append(responses,
			fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
			fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
		)
	}
	// 6 本目用の応答も足しておく（予算切れで呼ばれないことを assert する）。
	responses = append(responses, fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))})
	synth, rt := newFakeSynthesizer(responses...)

	// When: SynthesizeAll する
	_, err := synth.SynthesizeAll(context.Background(), texts)

	// Then: 合計呼び出しは SynthesizeBudget 回で頭打ち。6 本目は呼ばれない。
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != SynthesizeBudget {
		t.Fatalf("total call count = %d, want %d（SynthesizeBudget 合計上限で 6 本目は呼ばない）", len(rt.calls), SynthesizeBudget)
	}
	// Then: error 文言に呼んだ回数と予算が載る
	if !strings.Contains(err.Error(), "budget") {
		t.Fatalf("Error() = %q, want it to mention budget", err.Error())
	}
}

// TestSynthesizeAll_capsLastSegmentAttemptsToRemainingBudget は
// 予算が MaxAttempts 未満しか残っていないセグメントは、その残予算ぶんだけ retry して打ち切ることを検証する。
func TestSynthesizeAll_capsLastSegmentAttemptsToRemainingBudget(t *testing.T) {
	// Given: 1 本目が「do → 503 → 成功」で 3 消費し、残予算を SynthesizeBudget-3。
	//        …ではなく、残予算 1 の状況を作るため、SynthesizeBudget-1 回ぶん成功で埋める設計にする。
	//        簡単のため: (SynthesizeBudget-1) 本のセグメントを各 1 回成功で消費 → 残予算 1。
	//        最後のセグメントは 503 を返し続ける → 残予算 1 なので 1 回だけ呼んで打ち切り。
	texts := make([]string, SynthesizeBudget) // SynthesizeBudget-1 本成功 + 1 本失敗
	for i := range texts {
		texts[i] = fmt.Sprintf("セグメント%d", i)
	}
	var responses []fakeClientResponse
	for i := 0; i < SynthesizeBudget-1; i++ {
		responses = append(responses, fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))})
	}
	// 最後のセグメント用に 503 を複数積む（1 回しか呼ばれないはず）。
	responses = append(responses,
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
	)
	synth, rt := newFakeSynthesizer(responses...)

	// When: SynthesizeAll する
	_, err := synth.SynthesizeAll(context.Background(), texts)

	// Then: 合計 SynthesizeBudget 回。最後のセグメントは残予算 1 のため 1 回だけ呼んで打ち切り。
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != SynthesizeBudget {
		t.Fatalf("total call count = %d, want %d（最後のセグメントは残予算 1 回のみ）", len(rt.calls), SynthesizeBudget)
	}
}

// TestSynthesizeAll_capsSingleSegmentAtMaxAttempts_whenBudgetRemains は
// 予算が潤沢でも 1 セグメントは MaxAttempts 相当で打ち切ることを検証する
// （1 セグメントの暴走で予算全部を食わせない二段構え）。
func TestSynthesizeAll_capsSingleSegmentAtMaxAttempts_whenBudgetRemains(t *testing.T) {
	// Given: 2 本の text。1 本目は常に 503（同種 retryable）→ 同種 2 連続で 2 回打ち切り。
	responses := []fakeClientResponse{
		{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: SynthesizeAll する
	_, err := synth.SynthesizeAll(context.Background(), []string{"暴走セグメント", "健全セグメント"})

	// Then: 1 本目で失敗して即 return（2 本目は呼ばれない）。合計 2 回。
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（1 本目の同種 2 連続打ち切りで即 return）", len(rt.calls))
	}
}

// TestSynthesizeAll_consumesPaidBudget_whenTierPaid は TierPaid を渡した SpeechSynthesizer が
// SynthesizeBudget ではなく SynthesizeBudgetPaid を合計上限として使うことを検証する。
func TestSynthesizeAll_consumesPaidBudget_whenTierPaid(t *testing.T) {
	// Given: SynthesizeBudget（free 上限）を超える本数の text。すべて 1 回で成功する。
	texts := make([]string, SynthesizeBudget+1)
	for i := range texts {
		texts[i] = fmt.Sprintf("セグメント%d", i)
	}
	responses := make([]fakeClientResponse, len(texts))
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))}
	}
	synth, rt := newFakeSynthesizer(responses...)
	synth.tier = TierPaid

	// When: SynthesizeAll する
	got, err := synth.SynthesizeAll(context.Background(), texts)

	// Then: free budget（SynthesizeBudget）を超えても成功する。呼び出し回数は texts と同数
	if err != nil {
		t.Fatalf("SynthesizeAll: %v", err)
	}
	if len(got) != len(texts) {
		t.Fatalf("audios = %d, want %d", len(got), len(texts))
	}
	if len(rt.calls) != len(texts) {
		t.Fatalf("call count = %d, want %d（TierPaid は SynthesizeBudgetPaid まで許す）", len(rt.calls), len(texts))
	}
}

// TestSynthesizeBudget_returnsPaidBudget_whenTierPaid は synthesizeBudget() の tier 分岐そのものを検証する。
func TestSynthesizeBudget_returnsPaidBudget_whenTierPaid(t *testing.T) {
	t.Parallel()

	// Given: tier だけが異なる 2 つの SpeechSynthesizer
	free := &SpeechSynthesizer{tier: TierFree}
	paid := &SpeechSynthesizer{tier: TierPaid}

	// Then: TierFree は SynthesizeBudget、TierPaid は SynthesizeBudgetPaid
	if got := free.synthesizeBudget(); got != SynthesizeBudget {
		t.Fatalf("free.synthesizeBudget() = %d, want %d", got, SynthesizeBudget)
	}
	if got := paid.synthesizeBudget(); got != SynthesizeBudgetPaid {
		t.Fatalf("paid.synthesizeBudget() = %d, want %d", got, SynthesizeBudgetPaid)
	}
	if SynthesizeBudgetPaid != SynthesizeBudget*2 {
		t.Fatalf("SynthesizeBudgetPaid = %d, want %d（free の 2 倍）", SynthesizeBudgetPaid, SynthesizeBudget*2)
	}
}

func TestSameGeminiOp_returnsTrue_whenSourceAndOpBothMatch(t *testing.T) {
	t.Parallel()

	// Given: Source も Op も等しい 2 つの Infrastructure Error
	prev := adaptererror.New("gemini", "http_status", errors.New("status 503"))
	cur := adaptererror.New("gemini", "http_status", errors.New("status 503"))

	// When / Then: 同種と判定する
	if !sameGeminiOp(prev, cur) {
		t.Fatal("sameGeminiOp() = false, want true")
	}
}

func TestSameGeminiOp_returnsFalse_whenOpMatchesButSourceDiffers(t *testing.T) {
	t.Parallel()

	// Given: Op は同じ共通語彙だが Source（発生 Adapter）が違う 2 つの Infrastructure Error
	prev := adaptererror.New("gemini", "http_status", errors.New("status 503"))
	cur := adaptererror.New("geminiapi", "http_status", errors.New("status 503"))

	// When / Then: 別 Adapter 由来なので同種ではない
	if sameGeminiOp(prev, cur) {
		t.Fatal("sameGeminiOp() = true, want false（Source が違う）")
	}
}

func TestSynthesizeOne_returnsInfrastructureError_whenTextEmptyAfterTrim(t *testing.T) {

	// Given: 空本文。Client は呼ばれない想定なので response 設定は不要
	synth, rt := newFakeSynthesizer()

	// When: trim 後空の text を渡す
	_, err := synth.synthTestOne(context.Background(), "  \t\n  ")

	// Then: Infrastructure Error。Client は呼ばない
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "gemini:") {
		t.Fatalf("Error() = %q, want prefix gemini:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
	if len(rt.calls) != 0 {
		t.Fatalf("unexpected calls: %#v", rt.calls)
	}
}

func TestSynthesizeOne_retriesTransientError_thenSucceeds(t *testing.T) {

	// Given: 1 回目 503、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	)

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "retry テスト")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestSynthesizeOne_returnsInfrastructureError_whenSameRetryableOpRepeatsTwice(t *testing.T) {

	// Given: 常に 503（同種 retryable error = Op "http_status"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "打ち切りテスト")

	// Then: 同種 error 2 連続で打ち切り、call は 2 回で、fallback へ渡す番兵と Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v（5xx は fallback）", err)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestSynthesizeOne_retriesUpToMaxAttempts_whenRetryableOpAlternates(t *testing.T) {

	// Given: 503（http_status）と transport error（do）が交互。Op が変わるので連続打ち切りに掛からない
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		fakeClientResponse{err: fmt.Errorf("connection reset")},
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
	)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "交互リトライ")

	// Then: MaxAttempts 回まで回って Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != MaxAttempts {
		t.Fatalf("call count = %d, want %d（Op 交互は打ち切らない）", len(rt.calls), MaxAttempts)
	}
}

func TestSynthesizeOne_retriesMissingAudio_whenStatusInternalError(t *testing.T) {

	// Given: 1 回目 500（audio 欠落）、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusInternalServerError, body: jsonBody(t, map[string]any{"error": "internal"})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	)

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "500 retry")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestSynthesizeOne_retriesTooManyRequests_thenSucceeds(t *testing.T) {

	// Given: 1 回目 429（公式 error.code が rate_limit_exceeded）、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusTooManyRequests, body: jsonBody(t, map[string]any{"error": map[string]any{"code": "rate_limit_exceeded"}})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	)

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "429 retry")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestSynthesizeOne_wrapsSourceExhausted_when429RetriesAreExhausted(t *testing.T) {
	for _, tier := range []Tier{TierFree, TierPaid} {
		t.Run(fmt.Sprintf("tier_%d", tier), func(t *testing.T) {
			// Given: 回復を明示した同じ429（rate_limit_exceeded）が続き、retryを使い切る
			synth, rt := newFakeSynthesizer(
				fakeClientResponse{status: http.StatusTooManyRequests, body: jsonBody(t, map[string]any{"error": map[string]any{"code": "rate_limit_exceeded"}})},
				fakeClientResponse{status: http.StatusTooManyRequests, body: jsonBody(t, map[string]any{"error": map[string]any{"code": "rate_limit_exceeded"}})},
			)
			synth.tier = tier

			// When
			_, err := synth.synthTestOne(context.Background(), "429使い切り")

			// Then
			if !errors.Is(err, port.ErrSourceExhausted) {
				t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v", err)
			}
			if len(rt.calls) != 2 {
				t.Fatalf("call count = %d、期待値 = 2", len(rt.calls))
			}
		})
	}
}

func TestSynthesizeOne_wrapsSourceExhaustedWithoutWaiting_when429DeclaresNoRecovery(t *testing.T) {

	// Given: 回復の明示（Retry-After も rate_limit_exceeded も）が無い 429
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusTooManyRequests, body: jsonBody(t, map[string]any{"error": "RESOURCE_EXHAUSTED"})},
	)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "回復の明示なし")

	// Then: retry せず 1 call で fallback へ渡す番兵を返す
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_retries429_whenRetryAfterHeaderDeclaresRecovery(t *testing.T) {

	// Given: error.code は未知だが Retry-After 付きの 429、その後成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{
			status: http.StatusTooManyRequests,
			header: http.Header{"Retry-After": {"1"}},
			body:   jsonBody(t, map[string]any{"error": "RESOURCE_EXHAUSTED"}),
		},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	)

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "Retry-After で回復")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 || len(rt.calls) != 2 {
		t.Fatalf("content=%d bytes, call count = %d, want 非空 / 2", len(got.Content), len(rt.calls))
	}
}

func TestSynthesizeOne_wrapsSourceExhausted_whenTransportErrorRepeatsTwice(t *testing.T) {

	// Given: client.Do error が続く（client 起因か server 起因かは区別できない）
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{err: fmt.Errorf("connection reset")},
		fakeClientResponse{err: fmt.Errorf("connection reset")},
	)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "通信断")

	// Then: 同種 2 連続で打ち切り、次 source へ渡せる番兵を返す
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v", err)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestSynthesizeOne_wrapsSourceExhausted_whenAudioIsMissingRepeatedly(t *testing.T) {

	// Given: HTTP 200 だが audio が欠落した応答が続く
	missing := jsonBody(t, map[string]any{"status": "completed", "steps": []any{}})
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusOK, body: missing},
		fakeClientResponse{status: http.StatusOK, body: missing},
	)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "audio 欠落")

	// Then: 同種 2 連続で打ち切り、次 source へ渡せる番兵を返す
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v", err)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestSynthesizeOne_doesNotRetry_whenStatusBadRequest(t *testing.T) {

	// Given: 400
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusBadRequest,
		body:   jsonBody(t, map[string]any{"error": "INVALID_ARGUMENT"}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "400 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_doesNotRetry_whenStatusForbidden(t *testing.T) {

	// Given: 403
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusForbidden,
		body:   jsonBody(t, map[string]any{"error": "PERMISSION_DENIED"}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "403 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_returnsInfrastructureError_whenUnexpectedStatus(t *testing.T) {

	// Given: 404
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusNotFound,
		body:   jsonBody(t, map[string]any{"error": "NOT_FOUND"}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "404")

	// Then: retry しない
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_doesNotRetry_whenProhibitedContent(t *testing.T) {

	// Given: PROHIBITED_CONTENT
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"code": "PROHIBITED_CONTENT"}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "禁止テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_returnsInfrastructureError_whenUpstreamDoFails(t *testing.T) {

	// Given: Client.Do が常に transport error を返す（同種 retryable = Op "do"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{err: fmt.Errorf("connection refused")}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "network 失敗")

	// Then: 同種 2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestRetryDelay_growsFromBaseAndCapsAtMax(t *testing.T) {
	t.Parallel()

	s := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierFree, nil)

	if got := s.retryDelay(1); got != defaultRetryBackoffBase {
		t.Fatalf("retryDelay(1) = %v, want %v", got, defaultRetryBackoffBase)
	}
	if got := s.retryDelay(2); got != 2*defaultRetryBackoffBase {
		t.Fatalf("retryDelay(2) = %v, want %v", got, 2*defaultRetryBackoffBase)
	}
	if got := s.retryDelay(10); got != defaultRetryBackoffMax {
		t.Fatalf("retryDelay(10) = %v, want %v", got, defaultRetryBackoffMax)
	}
	if defaultRetryBackoffBase < 60*time.Second {
		t.Fatalf("defaultRetryBackoffBase = %v, want >= 60s（429 対策）", defaultRetryBackoffBase)
	}
}

func TestParseRetryAfter_returnsDuration_whenSecondsHeaderPresent(t *testing.T) {
	t.Parallel()

	s := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierFree, nil)

	h := http.Header{}
	h.Set("Retry-After", "90")
	if got := s.parseRetryAfter(h); got != 90*time.Second {
		t.Fatalf("parseRetryAfter = %v, want 90s", got)
	}
}

func TestParseRetryAfter_returnsZero_whenHeaderMissingOrInvalid(t *testing.T) {
	t.Parallel()

	s := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierFree, nil)

	if got := s.parseRetryAfter(http.Header{}); got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
	h := http.Header{}
	h.Set("Retry-After", "nope")
	if got := s.parseRetryAfter(h); got != 0 {
		t.Fatalf("invalid = %v, want 0", got)
	}
}

// TestWaitCallGap_usesInjectedCallGap_whenTuningProvided は
// NewSpeechSynthesizerWithTuning で短い CallGap を注入すると、連続 Synthesize の
// 待機が既定値ではなく注入値どおりに縮むことを検証する。
func TestWaitCallGap_usesInjectedCallGap_whenTuningProvided(t *testing.T) {
	// Given: 常に成功応答を返す fake client と、待機呼び出しを記録する sleep spy
	rt := &fakeRoundTripper{responses: []fakeClientResponse{
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	}}
	var sleeps []time.Duration
	const injectedGap = 5 * time.Millisecond
	synth := NewSpeechSynthesizerWithTuning(&http.Client{Transport: rt}, "gemini-fake-key", TierFree, Tuning{CallGap: injectedGap}, nil)
	synth.backoffSleepFn = func(d time.Duration) { sleeps = append(sleeps, d) }
	// nowFn は常に同じ時刻。elapsed=0 なので毎回 callGap 全量の待機が入るはず。
	synth.nowFn = func() time.Time { return time.Unix(0, 0) }

	// When: 2 回連続で Synthesize する
	for i := 0; i < 2; i++ {
		if _, err := synth.synthTestOne(context.Background(), "注入テスト"); err != nil {
			t.Fatalf("Synthesize(%d): %v", i+1, err)
		}
	}

	// Then: 待機は 1 度だけ（2 回目の呼び出し前）で、注入値どおり。既定値ではない。
	if len(sleeps) != 1 {
		t.Fatalf("sleep 呼び出し回数 = %d (%v), want 1", len(sleeps), sleeps)
	}
	if sleeps[0] != injectedGap {
		t.Fatalf("sleep = %v, want %v（注入 CallGap）", sleeps[0], injectedGap)
	}
}

// TestWaitCallGap_skipsWait_whenElapsedExceedsInjectedGap は
// 前回呼び出しからの経過が注入 CallGap を超えていれば待機しないことを検証する。
func TestWaitCallGap_skipsWait_whenElapsedExceedsInjectedGap(t *testing.T) {
	// Given: 成功応答 2 回分の fake client
	rt := &fakeRoundTripper{responses: []fakeClientResponse{
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	}}
	var sleeps []time.Duration
	synth := NewSpeechSynthesizerWithTuning(&http.Client{Transport: rt}, "gemini-fake-key", TierFree, Tuning{CallGap: time.Second}, nil)
	synth.backoffSleepFn = func(d time.Duration) { sleeps = append(sleeps, d) }
	// nowFn は呼ぶたびに 10s 進む。経過 >> CallGap なので待機は入らない。
	base := time.Unix(0, 0)
	calls := 0
	synth.nowFn = func() time.Time {
		calls++
		return base.Add(time.Duration(calls) * 10 * time.Second)
	}

	// When: 2 回連続で Synthesize する
	for i := 0; i < 2; i++ {
		if _, err := synth.synthTestOne(context.Background(), "経過超過テスト"); err != nil {
			t.Fatalf("Synthesize(%d): %v", i+1, err)
		}
	}

	// Then: callGap 待機は発生しない
	if len(sleeps) != 0 {
		t.Fatalf("sleep 呼び出し = %v, want なし（経過が CallGap を超過）", sleeps)
	}
}

// why: System で `decode_wav: output audio is missing` が MaxAttempts 尽きても
//
//	応答本文が error に載らず原因（finish_reason / safety / body 内 quota）が読めない（run 33581258235）。
//	HTTP 200 で audio 欠落のとき、error は本文の bounded snippet を含む。
func TestSynthesize_includesResponseBodySnippet_whenOutputAudioMissingOnOK(t *testing.T) {

	// Given: HTTP 200 だが output_audio が無く、代わりに finish_reason を持つ本文
	const marker = "SAFETY_BLOCKLIST_TRIGGERED"
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"finish_reason": marker,
				"note":          "no audio produced",
			}),
		}
	}
	synth, _ := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "audio 欠落の原因を知りたい")

	// Then: error 文言に本文の marker が含まれる
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), marker) {
		t.Fatalf("Error() = %q, want it to contain response body marker %q", err.Error(), marker)
	}
}

func TestBuildInput_structuresTranscriptWithSpeechMetadata_whenCallingProxy(t *testing.T) {

	// Given: 成功応答を返す Client Stub
	const transcript = "朗読する本文だけ"
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, audioInteractionResponse(minimalWAV())),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), transcript)

	// Then: user_input + text + speech_metadata annotation が input へ入る
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d", len(rt.calls))
	}
	var req map[string]any
	if err := json.Unmarshal(rt.calls[0].Body, &req); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	input, _ := req["input"].([]any)
	if len(input) != 1 {
		t.Fatalf("input len = %d, want 1", len(input))
	}
	firstInput, _ := input[0].(map[string]any)
	if firstInput["type"] != "user_input" {
		t.Fatalf("input[0].type = %v, want user_input", firstInput["type"])
	}
	contentList, _ := firstInput["content"].([]any)
	if len(contentList) != 1 {
		t.Fatalf("content len = %d, want 1", len(contentList))
	}
	firstContent, _ := contentList[0].(map[string]any)
	if firstContent["type"] != "text" {
		t.Fatalf("content[0].type = %v, want text", firstContent["type"])
	}
	if firstContent["text"] != transcript {
		t.Fatalf("content[0].text = %v, want %q", firstContent["text"], transcript)
	}
	annotations, _ := firstContent["annotations"].([]any)
	if len(annotations) != 1 {
		t.Fatalf("annotations len = %d, want 1", len(annotations))
	}
	firstAnnotation, _ := annotations[0].(map[string]any)
	if firstAnnotation["type"] != "speech_metadata" {
		t.Fatalf("annotation[0].type = %v, want speech_metadata", firstAnnotation["type"])
	}
	genCfg, _ := req["generation_config"].(map[string]any)
	speechCfg, _ := genCfg["speech_config"].([]any)
	if len(speechCfg) != 1 {
		t.Fatalf("speech_config = %#v", speechCfg)
	}
	voiceCfg, _ := speechCfg[0].(map[string]any)
	if voiceCfg["voice"] != VoiceName {
		t.Fatalf("voice = %v, want %q", voiceCfg["voice"], VoiceName)
	}
	if req["model"] != ModelID {
		t.Fatalf("model = %v, want %q", req["model"], ModelID)
	}
}

func TestFetchWAV_includesResponseBodySnippet_whenClientErrorStatus(t *testing.T) {

	// Given: 403 応答の body に切り分け用の理由が入っている
	const reason = "PERMISSION_DENIED"
	synth, _ := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusForbidden,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"status": reason}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "権限エラーの原因を知りたい")

	// Then: http_status Infra Error に応答 body の snippet が載る（System 失敗の切り分け用）
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "response body:") {
		t.Fatalf("error message %q does not carry response body snippet", err.Error())
	}
	if !strings.Contains(err.Error(), reason) {
		t.Fatalf("error message %q does not carry response body reason %q", err.Error(), reason)
	}
}

func TestFetchPCM_wrapsSourceExhausted_whenStatusUnauthorized(t *testing.T) {
	// Given: 401 応答
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusUnauthorized,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"status": "UNAUTHENTICATED"}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "認証切れ")

	// Then: port.ErrSourceExhausted を wrap した error。retry しない
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("error = %v, want errors.Is(err, port.ErrSourceExhausted) == true", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1（fallback へ渡す失敗は retry しない）", len(rt.calls))
	}
}

func TestFetchPCM_wrapsSourceExhausted_whenStatusForbidden(t *testing.T) {
	// Given: 403 応答
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusForbidden,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"status": "PERMISSION_DENIED"}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "権限なし")

	// Then: port.ErrSourceExhausted を wrap した error。retry しない
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("error = %v, want errors.Is(err, port.ErrSourceExhausted) == true", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1（fallback へ渡す失敗は retry しない）", len(rt.calls))
	}
}

func TestFetchPCM_wrapsSourceExhausted_whenStatusBadRequestWithQuotaExceeded(t *testing.T) {
	// Given: 400 応答。body の error.code が公式仕様どおり quota_exceeded
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusBadRequest,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"code": "quota_exceeded"}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "quota 切れ")

	// Then: port.ErrSourceExhausted を wrap した error。retry しない
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("error = %v, want errors.Is(err, port.ErrSourceExhausted) == true", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1（fallback へ渡す失敗は retry しない）", len(rt.calls))
	}
}

func TestFetchPCM_wrapsSourceExhausted_whenStatusTooManyRequestsWithQuotaExceeded(t *testing.T) {
	// Given: 429 応答。body の error.code が公式仕様どおり quota_exceeded（message は任意文言）
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusTooManyRequests,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"code": "quota_exceeded", "message": "任意の文言"}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "日次 quota 切れ")

	// Then: port.ErrSourceExhausted を wrap した error。retry しない
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("error = %v, want errors.Is(err, port.ErrSourceExhausted) == true", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1（quota_exceeded は retry しない）", len(rt.calls))
	}
}

func TestFetchPCM_retries_whenStatusTooManyRequestsWithRateLimitExceeded(t *testing.T) {
	// Given: 1 回目が 429 + rate_limit_exceeded、2 回目が成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{
			status: http.StatusTooManyRequests,
			body:   jsonBody(t, map[string]any{"error": map[string]any{"code": "rate_limit_exceeded"}}),
		},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalWAV()))},
	)

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "一過性 rate limit")

	// Then: 2 回目で成功する
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 || len(rt.calls) != 2 {
		t.Fatalf("content=%d bytes, call count = %d, want 非空 / 2", len(got.Content), len(rt.calls))
	}
}

func TestFetchPCM_doesNotWrapSourceExhausted_whenStatusBadRequestWithoutQuotaExceeded(t *testing.T) {
	// Given: 400 応答。body の error.code は通常の invalid argument
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusBadRequest,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"code": "invalid_argument"}}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "普通の 400")

	// Then: port.ErrSourceExhausted ではない通常の Infrastructure Error。retry しない
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("error = %v, want errors.Is(err, port.ErrSourceExhausted) == false（quota_exceeded を含まない 400）", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestClassifyFailedStatus_returnsRetryKind_perStatusAndErrorCode(t *testing.T) {

	// Given: 失敗 status・Retry-After・公式 error.code の組
	body := func(code string) []byte {
		return jsonBody(t, map[string]any{"error": map[string]any{"code": code}})
	}
	cases := []struct {
		name     string
		status   int
		header   http.Header
		body     []byte
		wantKind speechFetchRetryKind
	}{
		{"429 rate_limit_exceeded", http.StatusTooManyRequests, nil, body("rate_limit_exceeded"), wavRetryRateLimited},
		{"429 Retry-After 付き", http.StatusTooManyRequests, http.Header{"Retry-After": {"2"}}, body("unknown"), wavRetryRateLimited},
		{"429 回復の明示なし", http.StatusTooManyRequests, nil, body("unknown"), wavRetryFallback},
		{"429 quota_exceeded は Retry-After があっても fallback", http.StatusTooManyRequests, http.Header{"Retry-After": {"2"}}, body("quota_exceeded"), wavRetryFallback},
		{"400 quota_exceeded", http.StatusBadRequest, nil, body("quota_exceeded"), wavRetryFallback},
		{"401", http.StatusUnauthorized, nil, body("authentication"), wavRetryFallback},
		{"403", http.StatusForbidden, nil, body("permission_denied"), wavRetryFallback},
		{"400 は bug", http.StatusBadRequest, nil, body("invalid_request"), wavRetryNone},
		{"404 は bug", http.StatusNotFound, nil, body("not_found"), wavRetryNone},
		{"500", http.StatusInternalServerError, nil, body("internal"), wavRetryTransient},
		{"503 service_unavailable", http.StatusServiceUnavailable, nil, body("service_unavailable"), wavRetryTransient},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			synth, _ := newFakeSynthesizer()

			// When: 分類する
			gotKind, _ := synth.classifyFailedStatus(tc.status, tc.header, tc.body)

			// Then: 方針が決まる
			if gotKind != tc.wantKind {
				t.Fatalf("classifyFailedStatus(%d, %v, %s) = %v, want %v", tc.status, tc.header, tc.body, gotKind, tc.wantKind)
			}
		})
	}
}

func TestDecodeWAV_extractsAudioFromStepsContentData_whenRealResponseShape(t *testing.T) {

	// Given: 実 Interactions API そっくりの応答（steps[].content[].data に audio base64）
	wav := minimalWAV()
	body := map[string]any{
		"id":           "v1_real_shape",
		"object":       "interaction",
		"model":        ModelID,
		"status":       "completed",
		"service_tier": "standard",
		"created":      "2026-09-02T08:36:33Z",
		"updated":      "2026-09-02T08:36:33Z",
		"usage": map[string]any{
			"total_tokens":        125,
			"total_output_tokens": 99,
			"output_tokens_by_modality": []map[string]any{
				{"modality": "audio", "tokens": 99},
			},
		},
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(wav)},
			}},
		},
	}
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, body),
	})

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "実応答形テスト")

	// Then: steps 構造から audio を取り出し非空 WAV を返す
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if !isWAV(got.Content) {
		t.Fatalf("Content is not WAV: %d bytes", len(got.Content))
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestDecodeWAV_returnsInfrastructureError_whenOutputAudioMissingOnOK(t *testing.T) {

	// Given: HTTP 200 だが output_audio が無い（同種 retryable = Op "decode_wav"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body:   jsonBody(t, map[string]any{"status": "ok"}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "audio 欠落")

	// Then: 同種 2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
	// Then: error 文言に body のトップレベルキー一覧が載り、fixture の "status" が含まれる
	if !strings.Contains(err.Error(), "top-level keys:") {
		t.Fatalf("Error() = %q, want it to contain %q", err.Error(), "top-level keys:")
	}
	if !strings.Contains(err.Error(), "status") {
		t.Fatalf("Error() = %q, want it to list top-level key %q", err.Error(), "status")
	}
}

func TestDecodeWAV_retriesTooShortWAV_asTransientDecodeFailure(t *testing.T) {

	// Given: HTTP 200 だが尺が minSpeechDurationSec 未満の極小 WAV（0.1s）。
	//        Gemini の一過性劣化なので decode_wav 相当の retryable として retry される。
	tinyWAV := synthHelperWAV(1, 24000, 16, 0.1)
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"steps": []map[string]any{
					{"content": []map[string]any{
						{"data": base64.StdEncoding.EncodeToString(tinyWAV)},
					}},
				},
			}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "極小 WAV")

	// Then: 同種 retryable（Op "decode_wav"）2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if infra.Op != "decode_wav" {
		t.Fatalf("Op = %q, want %q（極小 WAV は decode_wav 系 retryable）", infra.Op, "decode_wav")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestDecodeWAV_returnsInfrastructureError_whenResponseBodyInvalidJSON(t *testing.T) {

	// Given: 壊れた JSON（同種 retryable = Op "decode_wav"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusOK, body: []byte(`not-json`)}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "decode 失敗")

	// Then: 同種 2 連続で打ち切り error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestDecodeWAV_returnsInfrastructureError_whenBase64Invalid(t *testing.T) {

	// Given: 不正 base64（同種 retryable = Op "decode_wav"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"steps": []map[string]any{
					{"content": []map[string]any{
						{"data": "!!!not-base64!!!"},
					}},
				},
			}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "bad b64")

	// Then: 同種 2 連続で打ち切り error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestQuotaExceeded_returnsTrue_whenErrorCodeIsQuotaExceeded(t *testing.T) {
	t.Parallel()

	// Given: 公式仕様どおりの flat 構造で error.code が quota_exceeded
	raw := jsonBody(t, map[string]any{"error": map[string]any{"code": "quota_exceeded", "message": "quota exceeded"}})

	// Then: true
	if !quotaExceeded(raw) {
		t.Fatalf("quotaExceeded(%s) = false, want true", raw)
	}
}

func TestQuotaExceeded_returnsFalse_whenErrorCodeDiffers(t *testing.T) {
	t.Parallel()

	// Given: error.code が quota_exceeded ではない
	raw := jsonBody(t, map[string]any{"error": map[string]any{"code": "invalid_argument"}})

	// Then: false
	if quotaExceeded(raw) {
		t.Fatalf("quotaExceeded(%s) = true, want false", raw)
	}
}

func TestQuotaExceeded_returnsFalse_whenBodyInvalidJSON(t *testing.T) {
	t.Parallel()

	// Given: JSON として parse できない body
	raw := []byte("not-json")

	// Then: false（parse 失敗を quota_exceeded と誤判定しない）
	if quotaExceeded(raw) {
		t.Fatal("quotaExceeded(invalid JSON) = true, want false")
	}
}

func TestToSpeechAudio_returnsInfrastructureError_whenWAVCorrupt(t *testing.T) {
	t.Parallel()

	// Given: 壊れた WAV bytes
	corrupt := []byte("not-a-wav-byte-stream")

	// When: toSpeechAudio する
	_, err := toSpeechAudio(corrupt)

	// Then: wav_duration Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, infra)
	}
	if infra.Op != "wav_duration" {
		t.Fatalf("Op = %q, want %q", infra.Op, "wav_duration")
	}
}
