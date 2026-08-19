package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/twitterapiio"
)

// NewTwitterAPIIOItemSource は TwitterAPI.io Adapter を組み立てる。
//
// @ensure 戻りは port.ItemSource。
func NewTwitterAPIIOItemSource() port.ItemSource {
	return twitterapiio.NewPostSource(&agentsecrets.Client{})
}
