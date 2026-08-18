package gemini

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
)

func TestSynthesize_returnsInfrastructureError_whenOutputAudioMissingOnOK(t *testing.T) {
	t.Parallel()

	// Given: HTTP 200 だが output_audio が無い
	var calls int
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		writeJSON(t, w, http.StatusOK, map[string]any{"status": "ok"})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "audio 欠落")

	// Then: retry 後 Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != MaxAttempts {
		t.Fatalf("calls = %d, want %d", calls, MaxAttempts)
	}
	if len(probe.TargetURLs) != MaxAttempts {
		t.Fatalf("request count = %d", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenResponseBodyInvalidJSON(t *testing.T) {
	t.Parallel()

	// Given: 壊れた JSON
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-json`))
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "decode 失敗")

	// Then: retry 後 error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != MaxAttempts {
		t.Fatalf("request count = %d", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: nil client
	synth := NewSpeechSynthesizer(nil)

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "本文")

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSynthesize_retriesTooManyRequests_thenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目 429、2 回目成功
	var calls int
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			writeJSON(t, w, http.StatusTooManyRequests, map[string]any{"error": "RESOURCE_EXHAUSTED"})
			return
		}
		writeJSON(t, w, http.StatusOK, audioInteractionResponse(minimalPCM()))
	})

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "429 retry")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(probe.TargetURLs) != 2 {
		t.Fatalf("request count = %d, want 2", len(probe.TargetURLs))
	}
}

func TestSynthesize_doesNotRetry_whenStatusForbidden(t *testing.T) {
	t.Parallel()

	// Given: 403
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusForbidden, map[string]any{"error": "PERMISSION_DENIED"})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "403 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != 1 {
		t.Fatalf("request count = %d, want 1", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenUnexpectedStatus(t *testing.T) {
	t.Parallel()

	// Given: 404
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusNotFound, map[string]any{"error": "NOT_FOUND"})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "404")

	// Then: retry しない
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != 1 {
		t.Fatalf("request count = %d, want 1", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenBase64Invalid(t *testing.T) {
	t.Parallel()

	// Given: 不正 base64
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"output_audio": map[string]any{"data": "!!!not-base64!!!"},
		})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "bad b64")

	// Then: retry 後 error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != MaxAttempts {
		t.Fatalf("request count = %d", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenPCMLengthOdd(t *testing.T) {
	t.Parallel()

	// Given: 奇数 byte の PCM
	oddPCM := []byte{0x00, 0x01, 0x02}
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"output_audio": map[string]any{
				"data": base64.StdEncoding.EncodeToString(oddPCM),
			},
		})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "奇数 pcm")

	// Then: pcm 変換失敗（非 retry）
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != 1 {
		t.Fatalf("request count = %d, want 1", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenReceiverNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var synth *SpeechSynthesizer

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "本文")

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *Error", err, err)
	}
}

func TestSynthesize_returnsInfrastructureError_whenProxyDoFails(t *testing.T) {
	t.Parallel()

	// Given: 閉じた proxy

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"error": "UNAVAILABLE"})
	}))
	client := server.Client()
	proxyURL := server.URL
	server.Close()

	synth := newSpeechSynthesizerForTest(&agentsecrets.Client{
		HTTP:     client,
		ProxyURL: proxyURL,
	}, func(time.Duration) {})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "network 失敗")

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
}
