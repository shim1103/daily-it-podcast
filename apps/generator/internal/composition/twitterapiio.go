package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/twitterapiio"
)

// NewTwitterAPIIOPostSource は TwitterAPI.io Adapter を組み立てる。
//
// @ensure 戻りは port.PostSource。
func NewTwitterAPIIOPostSource() port.PostSource {
	return twitterapiio.NewPostSource(&agentsecrets.Client{})
}
