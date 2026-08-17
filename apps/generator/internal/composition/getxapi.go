package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

// NewGetXAPIPostSource は GetXAPI Adapter を組み立てる。
//
// @ensure 戻りは port.PostSource。
func NewGetXAPIPostSource() port.PostSource {
	return getxapi.NewPostSource(&agentsecrets.Client{})
}
