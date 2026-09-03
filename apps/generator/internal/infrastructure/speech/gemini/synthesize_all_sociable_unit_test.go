package gemini

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestSynthesizeAll_succeedsForAllTexts_whenEveryCallReturnsAudio は
// texts を順に Synthesize し、WAV 列（結合しない）を texts と同数返すことを検証する。
func TestSynthesizeAll_succeedsForAllTexts_whenEveryCallReturnsAudio(t *testing.T) {
	// Given: 3 本の text、各 1 回で成功
	texts := []string{"一本目", "二本目", "三本目"}
	responses := make([]fakeClientResponse, len(texts))
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body:   jsonBody(t, audioInteractionResponse(minimalPCM())),
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
	// Given: 各セグメントは「do error → 503 → 成功」の 3 呼び出しで成功する（Op 交互なので連続打ち切り回避）。
	//        3 呼び出し × 5 セグメント = 15 = SynthesizeBudget。6 本目で予算切れ。
	const callsPerSegment = 3
	const fullSegments = SynthesizeBudget / callsPerSegment // 5
	texts := make([]string, fullSegments+1)
	for i := range texts {
		texts[i] = fmt.Sprintf("セグメント%d", i)
	}

	var responses []fakeClientResponse
	for seg := 0; seg < fullSegments; seg++ {
		responses = append(responses,
			fakeClientResponse{err: fmt.Errorf("connection reset")},
			fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
			fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
		)
	}
	// 6 本目用の応答も足しておく（予算切れで呼ばれないことを assert する）。
	responses = append(responses, fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))})
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
		responses = append(responses, fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))})
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
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
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
