package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

const tokenURL = "https://oauth2.googleapis.com/token"

// TokenSource は Google OAuth refresh で access token を取得する。
type TokenSource struct {
	client *agentsecrets.Client
}

// NewTokenSource は AgentSecrets proxy を使う Google OAuth TokenSource を返す。
//
// @ensure OAuth secret の実値を保持しない。
func NewTokenSource(client *agentsecrets.Client) *TokenSource {
	return &TokenSource{client: client}
}

// Token は refresh token を使って非空の access token を返す。
//
// @require ctx != nil。client != nil。
// @ensure 失敗時は元の原因を cause chain に保持した *Error を返す。
func (s *TokenSource) Token(ctx context.Context) (string, error) {
	if s == nil || s.client == nil {
		return "", infraErr("refresh", fmt.Errorf("client is nil"))
	}

	res, err := s.client.Do(ctx, agentsecrets.Request{
		Method:    http.MethodPost,
		TargetURL: tokenURL,
		Body:      strings.NewReader("grant_type=refresh_token"),
		PassthroughHeaders: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
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
