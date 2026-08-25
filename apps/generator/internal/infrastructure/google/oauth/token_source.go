package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

const tokenURL = "https://oauth2.googleapis.com/token"

// TokenSource は Google OAuth refresh で access token を取得する。
type TokenSource struct {
	client             secrettransport.Client
	clientIDSecret     secrettransport.SecretRef
	clientSecretSecret secrettransport.SecretRef
	refreshTokenSecret secrettransport.SecretRef
}

// NewTokenSource は Google OAuth TokenSource を返す。
//
// @ensure OAuth secret の実値を保持しない。secret 名の知識は持たず、各 SecretRef の参照だけを保持する。
func NewTokenSource(
	client secrettransport.Client,
	clientIDSecret secrettransport.SecretRef,
	clientSecretSecret secrettransport.SecretRef,
	refreshTokenSecret secrettransport.SecretRef,
) *TokenSource {
	return &TokenSource{
		client:             client,
		clientIDSecret:     clientIDSecret,
		clientSecretSecret: clientSecretSecret,
		refreshTokenSecret: refreshTokenSecret,
	}
}

// Token は refresh token を使って非空の access token を返す。
//
// @require ctx != nil。client != nil。
// @ensure 失敗時は元の原因を cause chain に保持した *Error を返す。
func (s *TokenSource) Token(ctx context.Context) (string, error) {
	if s == nil || s.client == nil {
		return "", infraErr("refresh", fmt.Errorf("client is nil"))
	}

	res, err := s.client.Do(ctx, secrettransport.Request{
		Method:    http.MethodPost,
		TargetURL: tokenURL,
		Body:      []byte("grant_type=refresh_token"),
		PassthroughHeaders: []secrettransport.Header{
			{Name: "Content-Type", Value: "application/x-www-form-urlencoded"},
		},
		Inject: secrettransport.Inject{
			Form: []secrettransport.FieldInjection{
				{Field: "client_id", Secret: s.clientIDSecret},
				{Field: "client_secret", Secret: s.clientSecretSecret},
				{Field: "refresh_token", Secret: s.refreshTokenSecret},
			},
		},
	})
	if err != nil {
		return "", infraErr("refresh_proxy", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", infraErr("refresh_read", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", infraErr("refresh_http", fmt.Errorf("status %d", res.StatusCode))
	}

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", infraErr("refresh_decode", err)
	}
	if strings.TrimSpace(response.AccessToken) == "" {
		return "", infraErr("refresh_decode", fmt.Errorf("access token is empty"))
	}
	return response.AccessToken, nil
}
