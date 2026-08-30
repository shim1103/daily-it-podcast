package config

// LookupEnv はprocess environmentのkeyを検索する契約である。
//
// os.LookupEnv と同 signature の注入 seam であり、testはこれを差し替えてenvを制御する。
type LookupEnv func(string) (string, bool)
