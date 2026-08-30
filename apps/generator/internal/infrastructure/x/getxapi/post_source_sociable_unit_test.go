package getxapi_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

const getXAPITestAPIKey = "getxapi-test-real-value"

// stubClientResponse は RoundTrip 1 回分の応答または error を表す。
type stubClientResponse struct {
	Status int
	Body   string
	Err    error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
type stubRoundTripper struct {
	responses []stubClientResponse
	calls     int
}

func (rt *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	index := rt.calls
	rt.calls++
	if index >= len(rt.responses) {
		return nil, fmt.Errorf("stubRoundTripper: no response configured for call %d", index)
	}
	res := rt.responses[index]
	if res.Err != nil {
		return nil, res.Err
	}
	status := res.Status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader([]byte(res.Body))),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func newStubPostSource(responses ...stubClientResponse) (*getxapi.PostSource, *stubRoundTripper) {
	rt := &stubRoundTripper{responses: responses}
	return getxapi.NewPostSource(&http.Client{Transport: rt}, getXAPITestAPIKey), rt
}

func TestList_mapsSourceItemFields_whenVendorReturnsTweet(t *testing.T) {
	// Given: watch user と vendor response を返す stub
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	source, _ := newStubPostSource(stubClientResponse{
		Body: `{"tweets":[{"id":"tweet-1","url":"https://x.example/tweet-1","text":"本文","createdAt":"Wed Aug 19 10:00:00 +0000 2026","author":{"id":"author-1","name":"表示名"},"entities":{"urls":[{"expanded_url":"https://example.com"}]},"media":[{"url":"https://img.example/a.jpg"}]}],"has_more":false}`,
	})
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: SourceItem と Context の規約を満たす
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	wantOccurredAt := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	wantContext := strings.Join([]string{
		"item_id: tweet-1",
		"actor_id: author-1",
		"actor_name: 表示名",
		"text: 本文",
		"permalink: https://x.example/tweet-1",
		"links: https://example.com",
		"media: https://img.example/a.jpg",
	}, "\n")
	if len(got) != 1 {
		t.Fatalf("got = %+v, want one SourceItem", got)
	}
	if got[0].SourceID != x.SourceID ||
		!got[0].OccurredAt.Equal(wantOccurredAt) ||
		got[0].OccurredAt.Location() != time.UTC ||
		got[0].Context != wantContext {
		t.Fatalf("got[0] = %+v, want SourceID=%q OccurredAt=%v Context=%q", got[0], x.SourceID, wantOccurredAt, wantContext)
	}
}

func TestList_paginates_and_stops_atSince(t *testing.T) {
	// Given: 1ページ目は次ページを持ち、2ページ目は since より古い
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{Body: `{"tweets":[{"id":"new","createdAt":"Wed Aug 19 10:00:00 +0000 2026"}],"has_more":true,"next_cursor":"next"}`},
			{Body: `{"tweets":[{"id":"old","createdAt":"Tue Aug 18 10:00:00 +0000 2026"}],"has_more":false}`},
		},
	}
	source := getxapi.NewPostSource(&http.Client{Transport: rt}, getXAPITestAPIKey)

	// When: since を指定して List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC))

	// Then: 2ページ目まで取得し、古い要素は含めない
	if err != nil || len(got) != 1 || rt.calls != 2 {
		t.Fatalf("got = %+v, err = %v, calls = %d", got, err, rt.calls)
	}
}

func TestList_returnsInfrastructureError_whenClientNil(t *testing.T) {
	// Given: client が nil
	source := getxapi.NewPostSource(nil, getXAPITestAPIKey)
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: Infrastructure Error。Error / Unwrap が観測できる
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *getxapi.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *getxapi.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "getxapi:") {
		t.Fatalf("Error() = %q, want prefix getxapi:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestList_returnsError_whenResponseBodyIsInvalidJSON(t *testing.T) {
	// Given: watch user と、200 だが JSON でない body を返す stub
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	source, _ := newStubPostSource(stubClientResponse{Body: "not-json"})

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: decode 失敗で error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
	var infra *getxapi.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *getxapi.Error", err, err)
	}
}

func TestList_returnsError_whenConnectionFailsMidRequest(t *testing.T) {
	// Given: watch user と、RoundTrip が失敗する stub
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	source, _ := newStubPostSource(stubClientResponse{Err: errors.New("connection reset")})

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: do 失敗で error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
}

func TestList_returnsError_whenVendorStatusIsNotOK(t *testing.T) {
	// Given: watch user と HTTP error を返す stub
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	source, _ := newStubPostSource(stubClientResponse{Status: http.StatusBadGateway, Body: "失敗"})

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: 部分結果なしで error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
}
