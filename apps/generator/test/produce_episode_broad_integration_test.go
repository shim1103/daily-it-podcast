// Scope: Broad Integration
// 実物境界: ProduceEpisode 合成経路（composite ItemSource → Cursor TextWriter → Gemini → OAuth+gdrive WriteEpisode）
// Double: 真外部のみ httptest TLS redirect / fake Cursor agent。production Adapter 型・順序は composition と同型。
// @require dummy secret のみ。各 vendor 境界 I/O の枝網羅は Narrow / Sociable Unit が所有する。
// @ensure 合成 postcondition のみ assert する（成功時書込 1 組、0 件・途中失敗時書込なし、代表 call 回数）。
// @invariant Authorization header 形式・request path 枝・schema 深掘りは assert しない。
package test

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

func TestProduceEpisodeBroadIntegration_uploadsEpisodeArtifactsOnce_whenAllProductionAdaptersSucceed(t *testing.T) {
	// Given: 全 production Adapter が success double へ接続された合成 UseCase
	h := newBroadProduceEpisodeHarness(t, broadProduceEpisodeConfig{})
	wantSynth := integrationSynthesizeCallCount(broadIntegrationTopicCount)

	// When: Run する
	_, err := h.uc.Run(context.Background(), integrationTestFixedNow)

	// Then: error なし。TextWriter 1、Synthesize = 2+topic（topic+2 束）、Drive upload 2（json+wav の 1 組）
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	assertBroadDownstreamCalls(t, h, 1, wantSynth, 2)
}

func TestProduceEpisodeBroadIntegration_returnsNoSourceItemsWithoutDownstreamCalls_whenFetchReturnsEmpty(t *testing.T) {
	// Given: 情報源が 0 件を返す合成 UseCase
	h := newBroadProduceEpisodeHarness(t, broadProduceEpisodeConfig{emptySources: true})

	// When: Run する
	_, err := h.uc.Run(context.Background(), integrationTestFixedNow)

	// Then: Op = no_source_items。TextWriter / Synthesize / Drive 書込は 0
	var de *domainerrors.Error
	if !errors.As(err, &de) || de.Op != domainerrors.OpNoSourceItems {
		t.Fatalf("err = %v, want Domain Error Op = %q", err, domainerrors.OpNoSourceItems)
	}
	assertIntegrationSecretsNotLeaked(t, err)
	assertBroadDownstreamCalls(t, h, 0, 0, 0)
}

func TestProduceEpisodeBroadIntegration_writesNothing_whenTextWriterFails(t *testing.T) {
	// Given: fake Cursor agent が exit 1 する合成 UseCase
	h := newBroadProduceEpisodeHarness(t, broadProduceEpisodeConfig{cursorFail: true})

	// When: Run する
	_, err := h.uc.Run(context.Background(), integrationTestFixedNow)

	// Then: error あり。Synthesize / Drive 書込は 0
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	assertIntegrationSecretsNotLeaked(t, err)
	assertBroadDownstreamCalls(t, h, -1, 0, 0)
}

func TestProduceEpisodeBroadIntegration_writesNothing_whenSynthesizeFails(t *testing.T) {
	// Given: Gemini が 2 回目で 400 を返す合成 UseCase
	h := newBroadProduceEpisodeHarness(t, broadProduceEpisodeConfig{geminiFailAt: 2})

	// When: Run する
	_, err := h.uc.Run(context.Background(), integrationTestFixedNow)

	// Then: error あり。Drive 書込は 0
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	assertIntegrationSecretsNotLeaked(t, err)
	assertBroadDownstreamCalls(t, h, -1, 2, 0)
}
