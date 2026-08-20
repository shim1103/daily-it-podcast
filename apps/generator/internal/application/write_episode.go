package application

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/contracts"
)

type WriteEpisode struct {
	writer port.EpisodeWriter
}

// NewWriteEpisode は検証済み episode を EpisodeWriter へ渡す UseCase を返す。
//
// @require writer != nil
// @ensure 戻りは非 nil。
func NewWriteEpisode(writer port.EpisodeWriter) *WriteEpisode {
	return &WriteEpisode{writer: writer}
}

// Run は原稿と音声を検証してから EpisodeWriter へ 1 回だけ渡す。
//
// @require uc != nil かつ uc.writer != nil。
// @ensure 検証失敗時は writer.Write を呼ばない。成功時だけ writer.Write を呼ぶ。
func (uc *WriteEpisode) Run(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	if strings.TrimSpace(episodeID) == "" {
		return &domainerrors.EmptyEpisodeID{}
	}
	if len(audio.Content) == 0 {
		return &domainerrors.EmptyAudio{}
	}

	decoded, err := decodeManuscript(manuscript)
	if err != nil {
		return &domainerrors.InvalidManuscript{Err: err}
	}
	schema, err := jsonschema.CompileString("manuscript.schema.json", string(contracts.ManuscriptSchema))
	if err != nil {
		return &domainerrors.InvalidManuscript{Err: err}
	}
	if err := schema.Validate(decoded); err != nil {
		return &domainerrors.InvalidManuscript{Err: err}
	}

	fields, ok := decoded.(map[string]any)
	if !ok {
		return &domainerrors.InvalidManuscript{Err: io.ErrUnexpectedEOF}
	}
	manuscriptEpisodeID, ok := fields["episodeId"].(string)
	if !ok || manuscriptEpisodeID != episodeID {
		return &domainerrors.EpisodeIDMismatch{
			Expected: episodeID,
			Actual:   manuscriptEpisodeID,
		}
	}
	return uc.writer.Write(ctx, episodeID, manuscript, audio)
}

func decodeManuscript(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return document, nil
}
