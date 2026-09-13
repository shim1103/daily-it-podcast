package runtime

import "os"

// LookupEnv は config.Load が使う親環境アクセス手段を返す。
// production では os.LookupEnv。
//
// @ensure 戻りは os.LookupEnv。
func LookupEnv() func(key string) (string, bool) {
	return os.LookupEnv
}
