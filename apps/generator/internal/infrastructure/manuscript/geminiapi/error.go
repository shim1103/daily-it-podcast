package geminiapi

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"

const errorSource = "geminiapi"

func geminiErr(op string, err error) error {
	return adaptererror.New(errorSource, op, err)
}
