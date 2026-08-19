package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
	"github.com/shim1103/daily-it-podcast/contracts"
)

var _ port.EpisodeWriter = (*EpisodeWriter)(nil)

type EpisodeWriter struct {
	client *agentsecrets.Client
}

// NewEpisodeWriter は Google Drive 書込 Adapter を返す。
//
// @require client != nil
// @ensure 秘密値は保持しない。
func NewEpisodeWriter(client *agentsecrets.Client) *EpisodeWriter {
	return &EpisodeWriter{client: client}
}

func (w *EpisodeWriter) Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	if w == nil || w.client == nil {
		return infraErr("write", fmt.Errorf("client is nil"))
	}
	if strings.TrimSpace(episodeID) == "" {
		return &domainerrors.EmptyEpisodeID{}
	}
	if len(audio.Content) == 0 {
		return &domainerrors.EmptyAudio{}
	}
	jsonEpisodeID, err := validateManuscript(manuscript)
	if err != nil {
		return err
	}
	if jsonEpisodeID != episodeID {
		return &domainerrors.EpisodeIDMismatch{Stem: episodeID, EpisodeID: jsonEpisodeID}
	}

	token, err := w.refreshToken(ctx)
	if err != nil {
		return err
	}
	if err := w.putFile(ctx, token, episodeID+jsonExt, jsonMIME, manuscript); err != nil {
		return err
	}
	if err := w.putFile(ctx, token, episodeID+wavExt, wavMIME, audio.Content); err != nil {
		return err
	}
	return nil
}

type denyRemoteLoader struct{}

func (denyRemoteLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("jsonschema remote load is disabled: %s", url)
}

var (
	manuscriptSchemaOnce sync.Once
	manuscriptSchema     *jsonschema.Schema
	manuscriptSchemaErr  error
)

func compiledManuscriptSchema() (*jsonschema.Schema, error) {
	manuscriptSchemaOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(contracts.ManuscriptSchema))
		if err != nil {
			manuscriptSchemaErr = err
			return
		}
		compiler := jsonschema.NewCompiler()
		compiler.UseLoader(denyRemoteLoader{})
		if err := compiler.AddResource(schemaResourceURL, doc); err != nil {
			manuscriptSchemaErr = err
			return
		}
		manuscriptSchema, manuscriptSchemaErr = compiler.Compile(schemaResourceURL)
	})
	return manuscriptSchema, manuscriptSchemaErr
}

func validateManuscript(manuscript []byte) (string, error) {
	sch, err := compiledManuscriptSchema()
	if err != nil {
		return "", infraErr("compile_schema", err)
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(manuscript))
	if err != nil {
		return "", &domainerrors.InvalidManuscript{Err: err}
	}
	if err := sch.Validate(inst); err != nil {
		return "", &domainerrors.InvalidManuscript{Err: err}
	}
	var parsed struct {
		EpisodeID string `json:"episodeId"`
	}
	if err := json.Unmarshal(manuscript, &parsed); err != nil {
		return "", &domainerrors.InvalidManuscript{Err: err}
	}
	return parsed.EpisodeID, nil
}

func (w *EpisodeWriter) refreshToken(ctx context.Context) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	res, err := w.client.Do(ctx, agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: TokenURL,
		Body:      strings.NewReader(form.Encode()),
		PassthroughHeaders: map[string]string{
			"Content-Type": formMIME,
		},
		Inject: agentsecrets.Inject{
			Form: map[string]string{
				"client_id":     secretnames.GoogleOAuthClientIDName,
				"client_secret": secretnames.GoogleOAuthClientSecretName,
				"refresh_token": secretnames.GoogleOAuthRefreshTokenName,
			},
		},
	})
	if err != nil {
		return "", infraErr("refresh_token", err)
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := readJSONBody(res, "refresh_token", &parsed, http.StatusOK); err != nil {
		return "", err
	}
	if strings.TrimSpace(parsed.AccessToken) == "" {
		return "", infraErr("refresh_token_decode", fmt.Errorf("access_token is empty"))
	}
	return parsed.AccessToken, nil
}

func (w *EpisodeWriter) putFile(ctx context.Context, token, name, mime string, content []byte) error {
	fileID, err := w.findFileID(ctx, token, name)
	if err != nil {
		return err
	}
	if fileID == "" {
		fileID, err = w.createMetadata(ctx, token, name, mime)
		if err != nil {
			return err
		}
	}
	return w.uploadMedia(ctx, token, fileID, mime, content)
}

func (w *EpisodeWriter) findFileID(ctx context.Context, token, name string) (string, error) {
	q := url.Values{}
	q.Set("q", "name = '"+escapeDriveQueryValue(name)+"' and trashed = false")
	q.Set("fields", "files(id)")
	res, err := w.doDrive(ctx, http.MethodGet, FilesURL+"?"+q.Encode(), token, "", nil, agentsecrets.Inject{})
	if err != nil {
		return "", infraErr("list", err)
	}
	var parsed struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	if err := readJSONBody(res, "list", &parsed, http.StatusOK); err != nil {
		return "", err
	}
	if len(parsed.Files) == 0 {
		return "", nil
	}
	return parsed.Files[0].ID, nil
}

type fileMetadata struct {
	Name     string   `json:"name"`
	MimeType string   `json:"mimeType"`
	Parents  []string `json:"parents"`
}

func (w *EpisodeWriter) createMetadata(ctx context.Context, token, name, mime string) (string, error) {
	// why: parents の値は AgentSecrets の Body inject が埋める。code に folder ID を置かない。
	meta, err := json.Marshal(fileMetadata{
		Name:     name,
		MimeType: mime,
		Parents:  []string{""},
	})
	if err != nil {
		return "", infraErr("create_marshal", err)
	}
	res, err := w.doDrive(ctx, http.MethodPost, FilesURL, token, jsonMIME, bytes.NewReader(meta), agentsecrets.Inject{
		Body: map[string]string{"parents.0": secretnames.DriveFolderIDName},
	})
	if err != nil {
		return "", infraErr("create", err)
	}
	var parsed struct {
		ID string `json:"id"`
	}
	if err := readJSONBody(res, "create", &parsed, http.StatusOK, http.StatusCreated); err != nil {
		return "", err
	}
	if strings.TrimSpace(parsed.ID) == "" {
		return "", infraErr("create_decode", fmt.Errorf("file id is empty"))
	}
	return parsed.ID, nil
}

func (w *EpisodeWriter) uploadMedia(ctx context.Context, token, fileID, mime string, content []byte) error {
	target := UploadURL + "/" + url.PathEscape(fileID) + "?uploadType=media"
	res, err := w.doDrive(ctx, http.MethodPatch, target, token, mime, bytes.NewReader(content), agentsecrets.Inject{})
	if err != nil {
		return infraErr("upload", err)
	}
	defer func() { _ = res.Body.Close() }()
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		return infraErr("upload_read", err)
	}
	if res.StatusCode != http.StatusOK {
		return infraErr("upload_http", fmt.Errorf("status %d", res.StatusCode))
	}
	return nil
}

func (w *EpisodeWriter) doDrive(ctx context.Context, method, target, token, contentType string, body io.Reader, inject agentsecrets.Inject) (*http.Response, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	return w.client.Do(ctx, agentsecrets.Request{
		Method:             method,
		TargetURL:          target,
		Body:               body,
		PassthroughHeaders: headers,
		Inject:             inject,
	})
}

func readJSONBody(res *http.Response, op string, dest any, okStatuses ...int) error {
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return infraErr(op+"_read", err)
	}
	ok := false
	for _, status := range okStatuses {
		if res.StatusCode == status {
			ok = true
			break
		}
	}
	if !ok {
		return infraErr(op+"_http", fmt.Errorf("status %d", res.StatusCode))
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return infraErr(op+"_decode", err)
	}
	return nil
}

func escapeDriveQueryValue(name string) string {
	name = strings.ReplaceAll(name, `\`, `\\`)
	return strings.ReplaceAll(name, `'`, `\'`)
}
