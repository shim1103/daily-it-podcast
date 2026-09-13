package publickey

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"

// why: この package の Infrastructure Error の発生源名。adaptererror.New へ渡す。
const errorSource = "publickey"

func infraErr(op string, err error) error {
	return adaptererror.New(errorSource, op, err)
}
