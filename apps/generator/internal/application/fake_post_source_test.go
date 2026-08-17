package application_test

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.PostSource = (*fakePostSource)(nil)

type fakePostSource struct {
	calls []listByUserCall
	posts map[string][]models.Post
	errs  map[string]error
}

type listByUserCall struct {
	userID string
	since  time.Time
}

func (f *fakePostSource) ListByUser(_ context.Context, userID string, since time.Time) ([]models.Post, error) {
	f.calls = append(f.calls, listByUserCall{userID: userID, since: since})
	if err, ok := f.errs[userID]; ok {
		return nil, err
	}
	if posts, ok := f.posts[userID]; ok {
		return posts, nil
	}
	return []models.Post{}, nil
}
