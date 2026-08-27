package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
)

// ProduceEpisodeFactory は検証済みConfigからproduction UseCaseを組み立てるconstructor契約である。
//
// @require configはGeneratorのconfiguration boundaryで検証済みである。
// @ensure 戻りはconfigのcapabilityごとにproduction Adapterを結線したUseCaseである。
type ProduceEpisodeFactory func(config.Config) *application.ProduceEpisode
