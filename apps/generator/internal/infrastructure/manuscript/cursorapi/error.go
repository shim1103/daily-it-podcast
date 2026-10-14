package cursorapi

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"

const errorSource = "cursorapi"

func infraErr(op string, err error) error {
	return adaptererror.New(errorSource, op, err)
}
