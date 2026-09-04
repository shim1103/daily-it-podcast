package gdrive

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func newStubLookup(rt *stubRoundTripper, tokens TokenSource) *CompletedEpisodeLookup {
	return NewCompletedEpisodeLookup(&http.Client{Transport: rt}, tokens, testFolderID)
}

func TestHasPair_returnsTrue_whenSameStemJsonAndWavMatchDate(t *testing.T) {
	t.Parallel()

	// Given: folder に同 stem の json+wav。json の date が照会日と一致
	const stem = "ep-complete-1"
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, MatchPath: "files?", Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []map[string]any{
					{"id": "json-id", "name": stem + ".json"},
					{"id": "wav-id", "name": stem + ".wav"},
				},
			})},
			{MatchMethod: http.MethodGet, MatchPath: "alt=media", Status: http.StatusOK, Body: `{"date":"2026-08-31","episodeId":"` + stem + `"}`},
		},
	}
	lookup := newStubLookup(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 同日で照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: true。list と media download が走る
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if !got {
		t.Fatal("HasPair = false, want true")
	}
	if len(listCalls(rt.calls)) == 0 {
		t.Fatal("list was not called")
	}
	var sawMedia bool
	for _, c := range rt.calls {
		if strings.Contains(c.TargetURL, "alt=media") {
			sawMedia = true
		}
	}
	if !sawMedia {
		t.Fatal("media download was not called")
	}
}

func TestHasPair_returnsFalse_whenJsonOnlyForDate(t *testing.T) {
	t.Parallel()

	// Given: 同日 date の json のみ（対応 wav 無し）
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, MatchPath: "files?", Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []map[string]any{
					{"id": "json-id", "name": "ep-json-only.json"},
				},
			})},
		},
	}
	lookup := newStubLookup(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false。media download しない（完成候補が無い）
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
	for _, c := range rt.calls {
		if strings.Contains(c.TargetURL, "alt=media") {
			t.Fatalf("media download was called: %s", c.TargetURL)
		}
	}
}

func TestHasPair_returnsFalse_whenWavOnly(t *testing.T) {
	t.Parallel()

	// Given: wav のみ（date を持つ json が無い）
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, MatchPath: "files?", Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []map[string]any{
					{"id": "wav-id", "name": "ep-wav-only.wav"},
				},
			})},
		},
	}
	lookup := newStubLookup(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}

func TestHasPair_returnsFalse_whenNoFiles(t *testing.T) {
	t.Parallel()

	// Given: folder が空
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, MatchPath: "files?", Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []any{},
			})},
		},
	}
	lookup := newStubLookup(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}

func TestHasPair_returnsFalse_whenPairDateDiffers(t *testing.T) {
	t.Parallel()

	// Given: 完成ペアはあるが date が別日
	const stem = "ep-other-day"
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, MatchPath: "files?", Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []map[string]any{
					{"id": "json-id", "name": stem + ".json"},
					{"id": "wav-id", "name": stem + ".wav"},
				},
			})},
			{MatchMethod: http.MethodGet, MatchPath: "alt=media", Status: http.StatusOK, Body: `{"date":"2026-08-30","episodeId":"` + stem + `"}`},
		},
	}
	lookup := newStubLookup(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 別日で照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}

func TestHasPair_returnsInfrastructureError_whenTokenSourceFails(t *testing.T) {
	t.Parallel()

	// Given: TokenSource が error
	rt := &stubRoundTripper{}
	lookup := newStubLookup(rt, stubTokenSource{err: errors.New("refresh failed")})

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: Infrastructure Error。Client は 0 回
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	if len(rt.calls) != 0 {
		t.Fatalf("unexpected client calls: %+v", rt.calls)
	}
}

func TestHasPair_scopesListQueryToFolder(t *testing.T) {
	t.Parallel()

	// Given: 空 folder
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, MatchPath: "files?", Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []any{},
			})},
		},
	}
	lookup := newStubLookup(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 照会する
	_, err := lookup.HasPair(context.Background(), "2026-08-31")
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}

	// Then: list q に folder parents が入る
	lists := listCalls(rt.calls)
	if len(lists) != 1 {
		t.Fatalf("list calls = %d, want 1", len(lists))
	}
	if !strings.Contains(lists[0].TargetURL, "in+parents") && !strings.Contains(lists[0].TargetURL, "in%20parents") && !strings.Contains(lists[0].TargetURL, "in parents") {
		// url.Values Encode は space を + にする
		if !strings.Contains(lists[0].TargetURL, testFolderID) {
			t.Fatalf("list q missing folder: %s", lists[0].TargetURL)
		}
	}
	if !strings.Contains(lists[0].TargetURL, testFolderID) {
		t.Fatalf("list q missing folder id: %s", lists[0].TargetURL)
	}
}
