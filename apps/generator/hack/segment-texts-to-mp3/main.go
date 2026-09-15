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

// main は完成稿 json から TTS→sec 更新→TEST Drive へ json+mp3 を書く使い捨て入口。
//
// @require MANUSCRIPT_JSON。TEST_GEMINI_API_KEY。Drive OAuth 4 key（TEST_* を script が正規名へ map）。
// @ensure TEST folder に {episodeId}.json（sec のみ更新）と .mp3。原稿 text は変えない。
func main() {
	rawManuscript := strings.TrimSpace(os.Getenv("MANUSCRIPT_JSON"))
	apiKey := strings.TrimSpace(os.Getenv("TEST_GEMINI_API_KEY"))
	clientID := strings.TrimSpace(os.Getenv(config.GoogleOAuthClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(config.GoogleOAuthClientSecretEnv))
	refreshToken := strings.TrimSpace(os.Getenv(config.GoogleOAuthRefreshTokenEnv))
	folderID := strings.TrimSpace(os.Getenv(config.DriveFolderIDEnv))
	if rawManuscript == "" || apiKey == "" ||
		clientID == "" || clientSecret == "" || refreshToken == "" || folderID == "" {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: MANUSCRIPT_JSON / TEST_GEMINI_API_KEY / Drive OAuth 4 key が必要\n")
		os.Exit(1)
	}

	manuscript := []byte(rawManuscript)
	episodeID, err := EpisodeIDFromManuscript(manuscript)
	if err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: %v\n", err)
		os.Exit(1)
	}
	texts, err := SegmentTextsFromManuscript(manuscript)
	if err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: %v\n", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 30 * time.Minute}
	speech := gemini.NewSpeechSynthesizer(httpClient, apiKey)
	enc := ffmpeg.NewEncoder(runtime.LookPath(), runtime.CommandRun())

	fmt.Printf("segment-texts-to-mp3: synthesize+encode episode=%s segments=%d\n", episodeID, len(texts))
	result, err := EncodeSegmentTextsToMP3(context.Background(), texts, speech, enc.EncodeWAVToMP3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: %v\n", err)
		os.Exit(1)
	}

	updated, err := ApplyManuscriptSecs(manuscript, result.TopicStartSecs, result.EndingStartSec, result.DurationSec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: sec update: %v\n", err)
		os.Exit(1)
	}

	uploader := &driveUploader{
		http:     httpClient,
		tokens:   oauth.NewTokenSource(httpClient, clientID, clientSecret, refreshToken),
		folderID: folderID,
	}
	ctx := context.Background()
	if err := uploader.PutJSON(ctx, episodeID, updated); err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: upload json: %v\n", err)
		os.Exit(1)
	}
	if err := uploader.PutMP3(ctx, episodeID, result.MP3); err != nil {
		fmt.Fprintf(os.Stderr, "segment-texts-to-mp3: upload mp3: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("segment-texts-to-mp3: ok %s.{json,mp3} durationSec=%.2f mp3Bytes=%d\n",
		episodeID, result.DurationSec, len(result.MP3))
}
