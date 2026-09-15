package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/audio/ffmpeg"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/runtime"
)

// main は segmentTexts JSON 配列から TEST Drive へ mp3 だけ書く使い捨て入口。
//
// @require EPISODE_ID / SEGMENT_TEXTS_JSON。TEST_GEMINI_API_KEY。Drive OAuth 4 key（TEST_* を script が正規名へ map）。
// @ensure TEST folder に {episodeId}.mp3 のみ。json は書かない。
func main() {
	episodeID := strings.TrimSpace(os.Getenv("EPISODE_ID"))
	rawTexts := strings.TrimSpace(os.Getenv("SEGMENT_TEXTS_JSON"))
	apiKey := strings.TrimSpace(os.Getenv("TEST_GEMINI_API_KEY"))
	clientID := strings.TrimSpace(os.Getenv(config.GoogleOAuthClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(config.GoogleOAuthClientSecretEnv))
	refreshToken := strings.TrimSpace(os.Getenv(config.GoogleOAuthRefreshTokenEnv))
	folderID := strings.TrimSpace(os.Getenv(config.DriveFolderIDEnv))
	if episodeID == "" || rawTexts == "" || apiKey == "" ||
		clientID == "" || clientSecret == "" || refreshToken == "" || folderID == "" {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: EPISODE_ID / SEGMENT_TEXTS_JSON / TEST_GEMINI_API_KEY / Drive OAuth 4 key が必要\n")
		os.Exit(1)
	}

	texts, err := ParseSegmentTextsJSON(rawTexts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: %v\n", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 30 * time.Minute}
	speech := gemini.NewSpeechSynthesizer(httpClient, apiKey)
	enc := ffmpeg.NewEncoder(runtime.LookPath(), runtime.CommandRun())

	fmt.Printf("segment-texts-to-mp3: synthesize+encode episode=%s segments=%d\n", episodeID, len(texts))
	mp3, err := EncodeSegmentTextsToMP3(context.Background(), texts, speech, enc.EncodeWAVToMP3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: %v\n", err)
		os.Exit(1)
	}

	uploader := &mp3Uploader{
		http:     httpClient,
		tokens:   oauth.NewTokenSource(httpClient, clientID, clientSecret, refreshToken),
		folderID: folderID,
	}
	if err := uploader.PutMP3(context.Background(), episodeID, mp3); err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: upload: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("segment-texts-to-mp3: ok %s.mp3 bytes=%d\n", episodeID, len(mp3))
}
